package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
)

const (
	LeaseLockName      = "eventer-lease-lock"
	LeaseLockNamespace = "kube-system"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	// Leader id, needs to be unique
	id, err := os.Hostname()
	if err != nil {
		klog.Fatal(err)
	}
	id = id + "_" + string(uuid.NewUUID())
	klog.Infof("current replica id is %s\n", id)

	var startedLeading atomic.Bool

	// leader election uses the Kubernetes API by writing to a
	// lock object, which can be a LeaseLock object (preferred),
	// a ConfigMap, or an Endpoints (deprecated) object.
	// Conflicting writes are detected and each client handles those actions
	// independently.
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatal(err)
	}
	client := clientset.NewForConfigOrDie(config)

	end := make(<-chan struct{})
	run := func(ctx context.Context) {
		// complete your controller loop here
		klog.Info("kube-eventer start...")
		end = eventer(ctx)
	}

	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      LeaseLockName,
			Namespace: LeaseLockNamespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	// start the leader election code loop
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				startedLeading.Store(true)
				run(ctx)
			},
			OnStoppedLeading: func() {
				// we can do cleanup here, but note that this callback is always called
				// when the LeaderElector exits, even if it did not start leading.
				// Therefore, we should check if we actually started leading before
				// performing any cleanup operations to avoid unexpected behavior.
				klog.Infof("leader lost: %s", id)

				// Example check to ensure we only perform cleanup if we actually started leading
				if startedLeading.Load() {
					// Perform cleanup operations here
					// For example, releasing resources, closing connections, etc.
					klog.Info("Performing cleanup operations...")
					<-end
				} else {
					klog.Info("No cleanup needed as we never started leading.")
				}
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == id {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
}
