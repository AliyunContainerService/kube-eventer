package prometheus

import (
	v1 "k8s.io/api/core/v1"
	"strings"
)

// always can be used when no finer-grained reason is reached. It should be the last one when used.
func always(_ *v1.Event) bool {
	return true
}

func isPodImagePullBackOff(event *v1.Event) bool {
	if event.Reason == "Failed" {
		return strings.Contains(event.Message, "ImagePullBackOff") || strings.Contains(event.Message, "ErrImagePull")
	}
	if event.Reason == "BackOff" {
		return strings.Contains(event.Message, "Back-off pulling image")
	}
	return false
}

func isFailCreatePodExceedQuota(event *v1.Event) bool {
	return event.Reason == "FailedCreate" && strings.Contains(event.Message, "exceeded quota")
}

func isResourceInsufficient(event *v1.Event) bool {
	return event.Reason == "FailedScheduling" && strings.Contains(event.Message, "Insufficient")
}

func isPodFailStart(event *v1.Event) bool {
	return event.Reason == "Failed" &&
		event.InvolvedObject.Kind == "Pod" &&
		!strings.Contains(event.Message, "ImagePullBackOff") &&
		!strings.Contains(event.Message, "ErrImagePull") &&
		!strings.Contains(event.Message, "Failed to pull image")
}

func isPodCrash(event *v1.Event) bool {
	return event.Reason == "BackOff" &&
		strings.Contains(event.Message, "Back-off restarting failed container")
}

func isDiskProvisionFailSize(event *v1.Event) bool {
	return event.Reason == "ProvisioningFailed" &&
		strings.Contains(event.Message, "disk size is not supported")
}

func isDiskProvisionFail(event *v1.Event) bool {
	return event.Reason == "ProvisioningFailed" &&
		!strings.Contains(event.Message, "disk size is not supported")
}

func isFailedBindingNoStorageClass(event *v1.Event) bool {
	return event.Reason == "FailedBinding" && strings.Contains(event.Message, "no storage class is set")
}

func isNodePLEGUnhealthy(event *v1.Event) bool {
	return event.Reason == "NodeNotReady" && event.InvolvedObject.Kind == "Node" && strings.Contains(event.Message, "PLEG is not healthy")
}

func isNodeNotReady(event *v1.Event) bool {
	return event.Reason == "NodeNotReady" && event.InvolvedObject.Kind == "Node" && !strings.Contains(event.Message, "PLEG is not healthy")
}

func isClusterIPNotEnough(event *v1.Event) bool {
	return strings.Contains(event.Message, "IpNotEnough")
}
