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
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AliyunContainerService/kube-eventer/api"
	"github.com/AliyunContainerService/kube-eventer/common/flags"
	"github.com/AliyunContainerService/kube-eventer/manager"
	"github.com/AliyunContainerService/kube-eventer/sinks"
	"github.com/AliyunContainerService/kube-eventer/sources"
	"github.com/AliyunContainerService/kube-eventer/version"
	"k8s.io/klog"
)

const ServerShutdownTimeout = 20 * time.Second

var (
	argFrequency    = flag.Duration("frequency", 30*time.Second, "The resolution at which Eventer pushes events to sinks")
	argMaxProcs     = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	argSources      flags.Uris
	argSinks        flags.Uris
	argVersion      bool
	argEventMetrics bool
	argHealthzIP    = flag.String("healthz-ip", "0.0.0.0", "ip eventer health check service uses")
	argHealthzPort  = flag.Uint("healthz-port", 8084, "port eventer health check listens on")
)

func eventer(ctx context.Context) <-chan struct{} {
	end := make(chan struct{})
	defer func(end chan<- struct{}) {
		end <- struct{}{}
	}(end)

	flag.Var(&argSources, "source", "source(s) to read events from")
	flag.Var(&argSinks, "sink", "external sink(s) that receive events")
	flag.BoolVar(&argVersion, "version", false, "print version info and exit")
	flag.BoolVar(&argEventMetrics, "event-metrics", true, "whether to collect and export event metrics")
	flag.Parse()

	if argVersion {
		fmt.Println(version.VersionInfo())
		os.Exit(0)
	}

	setMaxProcs()

	klog.Info(strings.Join(os.Args, " "))
	klog.Info(version.VersionInfo())
	if err := validateFlags(); err != nil {
		klog.Fatal(err)
	}

	// sources
	if len(argSources) != 1 {
		klog.Fatal("Wrong number of sources specified")
	}
	sourceFactory := sources.NewSourceFactory()
	sources, err := sourceFactory.BuildAll(ctx, argSources, argEventMetrics)
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
	manager, err := manager.NewManager(ctx, sources[0], sinkManager, *argFrequency)
	if err != nil {
		klog.Fatalf("Failed to create main manager: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.Start()
	}()

	err = startHTTPServer(ctx)
	if err != nil {
		klog.Fatalf("Failed to start http server: %v", err)
	}
	klog.Info("HTTP server gracefully shutdown")

	// wait for all goroutine to avoid leak
	wg.Wait()

	return end
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
