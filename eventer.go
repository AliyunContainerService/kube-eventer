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
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AliyunContainerService/kube-eventer/api"
	"github.com/AliyunContainerService/kube-eventer/common/flags"
	"github.com/AliyunContainerService/kube-eventer/common/kubernetes"
	"github.com/AliyunContainerService/kube-eventer/manager"
	"github.com/AliyunContainerService/kube-eventer/sinks"
	"github.com/AliyunContainerService/kube-eventer/sources"
	"github.com/AliyunContainerService/kube-eventer/util"
	"github.com/AliyunContainerService/kube-eventer/version"
	"k8s.io/klog"
)

// A Pod is granted a term to terminate gracefully, which defaults to 30 seconds.
const ServerShutdownTimeout = 20 * time.Second

var (
	argFrequency            = flag.Duration("frequency", 30*time.Second, "The resolution at which Eventer pushes events to sinks")
	argMaxProcs             = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	argSources              flags.Uris
	argSinks                flags.Uris
	argVersion              bool
	argEventMetrics         bool
	argHealthzIP            = flag.String("healthz-ip", "0.0.0.0", "ip eventer health check service uses")
	argHealthzPort          = flag.Uint("healthz-port", 8084, "port eventer health check listens on")
	argLeaderElectionConfig = util.DefaultLeaderElectionConfiguration()
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Var(&argSources, "source", "source(s) to read events from")
	flag.Var(&argSinks, "sink", "external sink(s) that receive events")
	flag.BoolVar(&argVersion, "version", false, "print version info and exit")
	flag.BoolVar(&argEventMetrics, "event-metrics", true, "whether to collect and export event metrics")
	util.BindLeaderElectionFlags(&argLeaderElectionConfig, flag.CommandLine)

	flag.Parse()

	if argVersion {
		fmt.Println(version.VersionInfo())
		os.Exit(0)
	}

	klog.Infof(strings.Join(os.Args, " "))
	klog.Info(version.VersionInfo())
	if err := validateFlags(); err != nil {
		klog.Fatalf("validate flags, err: %v", err)
	}

	srcUri := argSources[0]
	if srcUri.Key != "kubernetes" {
		klog.Fatalf("source %s does not recognized", srcUri.Key)
	}
	kubeclient, err := kubernetes.GetKubernetesClient(&srcUri.Val)
	if err != nil {
		klog.Fatalf("create kubernetes client, err: %v", err)
	}

	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx := context.Background()

	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		if err = startHTTPServer(ctx); err != nil {
			klog.Fatalf("start http server, err: %v", err)
		}
		klog.Info("HTTP server shutdown gracefully")
	}()

	if argLeaderElectionConfig.LeaderElect {
		err = util.NewLeaderElection(run, kubeclient, &argLeaderElectionConfig, ctx)
		if err != nil {
			klog.Fatalf("leader election, err: %v", err)
		}
		return
	}
	if err = run(ctx); err != nil {
		klog.Fatal(err)
	}

	wg.Wait()
}

func run(ctx context.Context) error {
	setMaxProcs()

	// sources
	sourceFactory := sources.NewSourceFactory()
	sources, err := sourceFactory.BuildAll(argSources, argEventMetrics)
	if err != nil {
		return fmt.Errorf("create source, err: %w", err)
	}
	if len(sources) != 1 {
		return errors.New("require exactly 1 source")
	}

	// sinks
	sinksFactory := sinks.NewSinkFactory()
	sinkList := sinksFactory.BuildAll(argSinks)
	if len([]flags.Uri(argSinks)) != 0 && len(sinkList) == 0 {
		return errors.New("no available sink to use")
	}

	for _, sink := range sinkList {
		klog.Infof("Starting with %s sink", sink.Name())
	}
	sinkManager, err := sinks.NewEventSinkManager(sinkList, sinks.DefaultSinkExportEventsTimeout, sinks.DefaultSinkStopTimeout)
	if err != nil {
		return fmt.Errorf("create sink manager, err: %w", err)
	}

	// main manager
	manager, err := manager.NewManager(sources[0], sinkManager, *argFrequency)
	if err != nil {
		return fmt.Errorf("create main manager, err: %w", err)
	}

	manager.Start()
	klog.Info("Manager started")
	defer func() {
		manager.Stop()
		// TODO: How to ensure all sinks are stopped correctly?
		// Currently simply implemented through time.Sleep function.
		// time.Sleep(time.Second * 5)
		klog.Info("Manager stopped")
	}()

	<-ctx.Done()
	return nil
}

func startHTTPServer(ctx context.Context) error {
	srv := http.Server{Addr: net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort)))}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	serverErr := make(chan error, 1)
	go func() {
		// Capture ListenAndServe errors such as "port already in use".
		// However, when a server is gracefully shutdown, it is safe to ignore errors
		// returned from this method (given the select logic below), because
		// Shutdown causes ListenAndServe to always return http.ErrServerClosed.
		klog.Info("Starting eventer http service")
		serverErr <- srv.ListenAndServe()
	}()
	var err error
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
	case err = <-serverErr:
	}
	return err
}

func validateFlags() error {
	var minFrequency = 5 * time.Second

	if *argHealthzPort > 65534 {
		return fmt.Errorf("invalid port supplied for healthz %d", *argHealthzPort)
	}
	if *argFrequency < minFrequency {
		return fmt.Errorf("frequency needs to be no less than %s, supplied %s", minFrequency,
			*argFrequency)
	}

	if *argFrequency > api.MaxEventsScrapeDelay {
		return fmt.Errorf("frequency needs to be no greater than %s, supplied %s",
			api.MaxEventsScrapeDelay, *argFrequency)
	}

	if len(argSources) != 1 {
		klog.Fatal("Wrong number of sources specified")
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
