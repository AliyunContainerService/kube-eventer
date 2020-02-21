package kubernetes

import (
	clientv1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	kuberecord "k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

// CreateEventRecorder creates an event recorder to send custom events to Kubernetes to be recorded for targeted Kubernetes objects
func CreateEventRecorder(kubeClient clientset.Interface) kuberecord.EventRecorder {
	eventBroadcaster := kuberecord.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	if _, isfake := kubeClient.(*fake.Clientset); !isfake {
		eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: corev1.New(kubeClient.CoreV1().RESTClient()).Events("")})
	}
	return eventBroadcaster.NewRecorder(scheme.Scheme, clientv1.EventSource{Component: "kube-eventer"})
}
