// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate ./hooks/run_extpoints.sh

package main

import (
	ctx "context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AliyunContainerService/kube-eventer/api"
	"github.com/AliyunContainerService/kube-eventer/common/flags"
	kubeconfig "github.com/AliyunContainerService/kube-eventer/common/kubernetes"
	"github.com/AliyunContainerService/kube-eventer/manager"
	"github.com/AliyunContainerService/kube-eventer/sinks"
	"github.com/AliyunContainerService/kube-eventer/sources"
	"github.com/AliyunContainerService/kube-eventer/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/klog"
)

var (
	argFrequency = flag.Duration("frequency", 30*time.Second, "The resolution at which Eventer pushes events to sinks")
	argMaxProcs  = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	argSources   flags.Uris

	argSinks       flags.Uris
	argVersion     bool
	argHealthzIP   = flag.String("healthz-ip", "0.0.0.0", "ip eventer health check service uses")
	argHealthzPort = flag.Uint("healthz-port", 8084, "port eventer health check listens on")
	namespace      = flag.String("namespace", "kube-system", "Namespace in which kube-eventer run.")
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Var(&argSources, "source", "source(s) to read events from")
	flag.Var(&argSinks, "sink", "external sink(s) that receive events")
	flag.BoolVar(&argVersion, "version", false, "print version info and exit")
	leaderElection := defaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = true

	bindLeaderElectionFlags(&leaderElection)

	flag.Parse()

	if argVersion {
		fmt.Println(version.VersionInfo())
		os.Exit(0)
	}

	klog.Infof(strings.Join(os.Args, " "))
	klog.Info(version.VersionInfo())

	setMaxProcs()

	go startHTTPServer()

	if len(argSources) != 1 {
		klog.Fatal("Wrong number of sources specified")
	}

	if err := validateFlags(); err != nil {
		klog.Fatal(err)
	}

	if !leaderElection.LeaderElect {
		run()
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Unable to get hostname: %v", err)
		}

		cfg, err := kubeconfig.GetKubeClientConfig(&argSources[0].Val)
		if err != nil {
			klog.Fatalf("Get KubeClientConfig Error: %v", err)
		}
		kubeClient := kubeclient.NewForConfigOrDie(cfg)

		lock, err := resourcelock.New(
			leaderElection.ResourceLock,
			*namespace,
			"kube-eventer",
			kubeClient.CoreV1(),
			kubeClient.CoordinationV1(),
			resourcelock.ResourceLockConfig{
				Identity:      id,
				EventRecorder: kubeconfig.CreateEventRecorder(kubeClient),
			},
		)
		if err != nil {
			klog.Fatalf("Unable to create leader election lock: %v", err)
		}

		leaderelection.RunOrDie(ctx.TODO(), leaderelection.LeaderElectionConfig{
			Lock:          lock,
			LeaseDuration: leaderElection.LeaseDuration.Duration,
			RenewDeadline: leaderElection.RenewDeadline.Duration,
			RetryPeriod:   leaderElection.RetryPeriod.Duration,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ ctx.Context) {
					// Since we are committing a suicide after losing
					// mastership, we can safely ignore the argument.
					run()
				},
				OnStoppedLeading: func() {
					klog.Fatalf("lost master")
				},
			},
		})
	}
}

func run() {
	quitChannel := make(chan struct{}, 0)

	sourceFactory := sources.NewSourceFactory()
	sources, err := sourceFactory.BuildAll(argSources)
	if err != nil {
		klog.Fatalf("Failed to create sources: %v", err)
	}
	if len(sources) != 1 {
		klog.Fatal("Requires exactly 1 source")
	}

	// sinks
	sinksFactory := sinks.NewSinkFactory()
	sinkList := sinksFactory.BuildAll(argSinks)
	if len([]flags.Uri(argSinks)) != 0 && len(sinkList) == 0 {
		klog.Fatal("No available sink to use")
	}

	for _, sink := range sinkList {
		klog.Infof("Starting with %s sink", sink.Name())
	}
	sinkManager, err := sinks.NewEventSinkManager(sinkList, sinks.DefaultSinkExportEventsTimeout, sinks.DefaultSinkStopTimeout)
	if err != nil {
		klog.Fatalf("Failed to create sink manager: %v", err)
	}

	// main manager
	manager, err := manager.NewManager(sources[0], sinkManager, *argFrequency)
	if err != nil {
		klog.Fatalf("Failed to create main manager: %v", err)
	}

	manager.Start()
	klog.Infof("Starting eventer")

	<-quitChannel
}

func startHTTPServer() {
	klog.Info("Starting eventer http service")

	klog.Fatal(http.ListenAndServe(net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort))), nil))
}

func validateFlags() error {
	var minFrequency = 5 * time.Second

	if *argFrequency < minFrequency {
		return fmt.Errorf("frequency needs to be no less than %s, supplied %s", minFrequency,
			*argFrequency)
	}

	if *argFrequency > api.MaxEventsScrapeDelay {
		return fmt.Errorf("frequency needs to be no greater than %s, supplied %s",
			api.MaxEventsScrapeDelay, *argFrequency)
	}

	return nil
}

func setMaxProcs() {
	// Allow as many threads as we have cores unless the user specified a value.
	var numProcs int
	if *argMaxProcs < 1 {
		numProcs = runtime.NumCPU()
	} else {
		numProcs = *argMaxProcs
	}
	runtime.GOMAXPROCS(numProcs)

	// Check if the setting was successful.
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != numProcs {
		klog.Warningf("Specified max procs of %d but using %d", numProcs, actualNumProcs)
	}
}

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

func defaultLeaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:   false,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.LeasesResourceLock,
	}
}

func bindLeaderElectionFlags(l *componentbaseconfig.LeaderElectionConfiguration) {
	flag.BoolVar(&l.LeaderElect, "leader-elect", l.LeaderElect, ""+
		"Start a leader election client and gain leadership before "+
		"executing the main loop. Enable this when running replicated "+
		"components for high availability.")
	flag.DurationVar(&l.LeaseDuration.Duration, "leader-elect-lease-duration", l.LeaseDuration.Duration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	flag.DurationVar(&l.RenewDeadline.Duration, "leader-elect-renew-deadline", l.RenewDeadline.Duration, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	flag.DurationVar(&l.RetryPeriod.Duration, "leader-elect-retry-period", l.RetryPeriod.Duration, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	flag.StringVar(&l.ResourceLock, "leader-elect-resource-lock", l.ResourceLock, ""+
		"The type of resource object that is used for locking during "+
		"leader election. Supported options are `leases`(default), `endpoints` and `configmaps`.")
	flag.StringVar(&l.ResourceName, "leader-elect-resource-name", l.ResourceName, ""+
		"The name of resource object that is used for locking during "+
		"leader election.")
	flag.StringVar(&l.ResourceNamespace, "leader-elect-resource-namespace", l.ResourceNamespace, ""+
		"The namespace of resource object that is used for locking during "+
		"leader election.")
}
