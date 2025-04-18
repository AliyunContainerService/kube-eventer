package util

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/klog"
)

// NewLeaderElection starts the leader election code loop
func NewLeaderElection(
	run func(ctx context.Context) error,
	client clientset.Interface,
	LeaderElectionConfig *componentbaseconfig.LeaderElectionConfiguration,
	ctx context.Context,
) error {
	var id string

	if hostname, err := os.Hostname(); err != nil {
		// on errors, make sure we're unique
		id = string(uuid.NewUUID())
	} else {
		// add a uniquifier so that two processes on the same host don't accidentally both become active
		id = hostname + "_" + string(uuid.NewUUID())
	}

	klog.V(3).Infof("Assigned unique lease holder id: %s", id)

	if len(LeaderElectionConfig.ResourceNamespace) == 0 {
		return fmt.Errorf("namespace may not be empty")
	}

	if len(LeaderElectionConfig.ResourceName) == 0 {
		return fmt.Errorf("name may not be empty")
	}

	lock, err := resourcelock.New(
		LeaderElectionConfig.ResourceLock,
		LeaderElectionConfig.ResourceNamespace,
		LeaderElectionConfig.ResourceName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create leader election lock: %v", err)
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   LeaderElectionConfig.LeaseDuration.Duration,
		RenewDeadline:   LeaderElectionConfig.RenewDeadline.Duration,
		RetryPeriod:     LeaderElectionConfig.RetryPeriod.Duration,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.V(1).Info("Started leading")
				err := run(ctx)
				if err != nil {
					klog.Error(err)
				}
			},
			OnStoppedLeading: func() {
				klog.V(1).Info("Leader lost")
			},
			OnNewLeader: func(identity string) {
				// Just got the lock
				if identity == id {
					return
				}
				klog.V(1).Infof("New leader elected: %v", identity)
			},
		},
	})
	return nil
}

// BindLeaderElectionFlags binds the LeaderElectionConfiguration struct fields to a flagset
func BindLeaderElectionFlags(l *componentbaseconfig.LeaderElectionConfiguration, fs *flag.FlagSet) {
	fs.BoolVar(&l.LeaderElect, "leader-elect", l.LeaderElect, ""+
		"Start a leader election client and gain leadership before "+
		"executing the main loop. Enable this when running replicated "+
		"components for high availability.")
	fs.DurationVar(&l.LeaseDuration.Duration, "leader-elect-lease-duration", l.LeaseDuration.Duration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline.Duration, "leader-elect-renew-deadline", l.RenewDeadline.Duration, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod.Duration, "leader-elect-retry-period", l.RetryPeriod.Duration, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	fs.StringVar(&l.ResourceLock, "leader-elect-resource-lock", l.ResourceLock, ""+
		"The type of resource object that is used for locking during "+
		"leader election. Supported options are 'leases'.")
	fs.StringVar(&l.ResourceName, "leader-elect-resource-name", l.ResourceName, ""+
		"The name of resource object that is used for locking during "+
		"leader election.")
	fs.StringVar(&l.ResourceNamespace, "leader-elect-resource-namespace", l.ResourceNamespace, ""+
		"The namespace of resource object that is used for locking during "+
		"leader election.")
}

func DefaultLeaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:       true,
		LeaseDuration:     metav1.Duration{Duration: 60 * time.Second},
		RenewDeadline:     metav1.Duration{Duration: 15 * time.Second},
		RetryPeriod:       metav1.Duration{Duration: 5 * time.Second},
		ResourceLock:      "leases",
		ResourceName:      "kube-eventer",
		ResourceNamespace: "kube-system",
	}
}
