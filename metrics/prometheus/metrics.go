package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
)

type AbnormalEventReason string

var (
	// Namespace Lavel Event
	PodEvict                    AbnormalEventReason = "PodEvict"
	PodImagePullBackOff         AbnormalEventReason = "PodImagePullBackOff"
	PodOOM                      AbnormalEventReason = "PodOOM"
	ResourceInsufficient        AbnormalEventReason = "ResourceInsufficient"
	PodFailStart                AbnormalEventReason = "PodFailStart"
	PodCrash                    AbnormalEventReason = "PodCrash"
	PodFailScheduling           AbnormalEventReason = "PodFailScheduling"
	DiskProvisionFailSize       AbnormalEventReason = "DiskProvisionFailSize"
	DiskProvisionFail           AbnormalEventReason = "DiskProvisionFail"
	FailedBindingNoStorageClass AbnormalEventReason = "FailedBindingNoStorageClass"
	VolumeFailMount             AbnormalEventReason = "VolumeFailMount"
	FailCreatePodExceedQuota    AbnormalEventReason = "FailCreatePodExceedQuota"

	// Node Level Event
	NodeOOM           AbnormalEventReason = "NodeOOM"
	NodeRebooted      AbnormalEventReason = "NodeRebooted"
	NodeDiskPressure  AbnormalEventReason = "NodeDiskPressure"
	NodeDockerHung    AbnormalEventReason = "NodeDockerHung"
	NodePSHung        AbnormalEventReason = "NodePSHung"
	NodeGPUXIPError   AbnormalEventReason = "NodeGPUXIPError"
	NodeFDPressure    AbnormalEventReason = "NodeFDPressure"
	NodePLEGUnhealthy AbnormalEventReason = "NodePLEGUnhealthy"
	NodeNPTDown       AbnormalEventReason = "NodeNPTDown"
	NodeNotReady      AbnormalEventReason = "NodeNotReady"
	ConnTrackFull     AbnormalEventReason = "ConnTrackFull"

	// Core Component Event
	CcmSLBSyncFail          AbnormalEventReason = "CcmSLBSyncFail"
	CcmSLBUnavailable       AbnormalEventReason = "CcmSLBUnavailable"
	CcmSLBDeleteFail        AbnormalEventReason = "CcmSLBDeleteFail"
	CcmCreateRouteFail      AbnormalEventReason = "CcmCreateRouteFail"
	CcmSyncRouteFail        AbnormalEventReason = "CcmSyncRouteFail"
	CcmAddNodeFail          AbnormalEventReason = "CcmAddNodeFail"
	CcmDeleteNodeFail       AbnormalEventReason = "CcmDeleteNodeFail"
	CcmSLBAnnotationChanged AbnormalEventReason = "CcmSLBAnnotationChanged"
	CcmSLBSpecChanged       AbnormalEventReason = "CcmSLBSpecChanged"
	CSISlowIO               AbnormalEventReason = "CSISlowIO"
	CSIDeviceBusy           AbnormalEventReason = "CSIDeviceBusy"
	CSIIOHang               AbnormalEventReason = "CSIIOHang"
	CNIAllocIPFail          AbnormalEventReason = "CNIAllocIPFail"
	CNIAllocResourceFail    AbnormalEventReason = "CNIAllocResourceFail"
	CNIResourceInvalid      AbnormalEventReason = "CNIResourceInvalid"
	CNIParseFail            AbnormalEventReason = "CNIParseFail"
	CNIDisposeResourceFail  AbnormalEventReason = "CNIDisposeResourceFail"
	ClusterIPNotEnough      AbnormalEventReason = "ClusterIPNotEnough"
)

type JudgeEventFunc func(event *v1.Event) bool

type JudgeEvent struct {
	kind  AbnormalEventReason
	judge JudgeEventFunc
}

var (
	normalEventCounter  *prometheus.CounterVec
	errorEventCounter   *prometheus.CounterVec
	warningEventCounter *prometheus.CounterVec

	reasonToEventKind = map[string]AbnormalEventReason{
		// Namespace level
		"Evicted":            PodEvict,
		"OOMKilling":         PodOOM,
		"PodOOMKilling":      PodOOM,
		"FailedMount":        VolumeFailMount,
		"FailedAttachVolume": VolumeFailMount,

		// Node level
		"SystemOOM":             NodeOOM,
		"Rebooted":              NodeRebooted,
		"NodeHasDiskPressure":   NodeDiskPressure,
		"DockerHung":            NodeDockerHung,
		"PSProcessIsHung":       NodePSHung,
		"NodeHasNvidiaXidError": NodeGPUXIPError,
		"NodeHasFDPressure":     NodeFDPressure,
		"PIDPressure":           NodeFDPressure,
		"NodeHasPIDPressure":    NodeFDPressure,
		"NTPIsDown":             NodeNPTDown,
		"ConntrackFull":         ConnTrackFull,

		// Core Component
		"SyncLoadBalancerFailed":   CcmSLBSyncFail,
		"DeleteLoadBalancerFailed": CcmSLBDeleteFail,
		"CreateRouteFailed":        CcmCreateRouteFail,
		"SYncRouteFailed":          CcmSyncRouteFail,
		"AddNodeFailed":            CcmAddNodeFail,
		"DeteleNodeFailed":         CcmDeleteNodeFail,
		"UnAvailableLoadBalancer":  CcmSLBUnavailable,
		"AnnotationChanged":        CcmSLBAnnotationChanged,
		"ServiceSpecChanged":       CcmSLBSpecChanged,
		"SlowIO":                   CSISlowIO,
		"DeviceBusy":               CSIDeviceBusy,
		"IOHang":                   CSIIOHang,
		"AllocIPFailed":            CNIAllocIPFail,
		"ParseFailed":              CNIParseFail,
		"DisposeResourceFailed":    CNIDisposeResourceFail,
	}

	reasonToEventKindFunc = map[string][]JudgeEvent{
		"Failed": {
			{kind: PodFailStart, judge: isPodFailStart},
			{kind: PodImagePullBackOff, judge: isPodImagePullBackOff},
		},
		"BackOff": {
			{kind: PodCrash, judge: isPodCrash},
			{kind: PodImagePullBackOff, judge: isPodImagePullBackOff},
		},
		"FailedCreate": {{kind: FailCreatePodExceedQuota, judge: isFailCreatePodExceedQuota}},
		"FailedScheduling": {
			{kind: ResourceInsufficient, judge: isResourceInsufficient},
			{kind: PodFailScheduling, judge: always},
		},
		"ProvisioningFailed": {
			{kind: DiskProvisionFailSize, judge: isDiskProvisionFailSize},
			{kind: DiskProvisionFail, judge: isDiskProvisionFail},
		},
		"FailedBinding": {
			{kind: FailedBindingNoStorageClass, judge: isFailedBindingNoStorageClass},
		},
		"NodeNotReady": {
			{kind: NodePLEGUnhealthy, judge: isNodePLEGUnhealthy},
			{kind: NodeNotReady, judge: isNodeNotReady},
		},
		"AllocResourceFailed": {
			{kind: ClusterIPNotEnough, judge: isClusterIPNotEnough},
			{kind: CNIAllocResourceFail, judge: always},
		},
		"ResourceInvalid": {
			{kind: ClusterIPNotEnough, judge: isClusterIPNotEnough},
			{kind: CNIResourceInvalid, judge: always},
		},
	}
)

func InitMetrics() {
	normalEventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eventer",
			Subsystem: "events",
			Name:      "normal_total",
		},
		[]string{"reason", "namespace", "kind"},
	)
	errorEventLabels := []string{
		"reason",
		"kind",
		"name",
		"namespace",
	}
	errorEventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "eventer",
		Subsystem: "events",
		Name:      "error_total",
	}, errorEventLabels)
	warningEventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "eventer",
		Subsystem: "events",
		Name:      "warning_total",
	}, errorEventLabels)

	prometheus.MustRegister(normalEventCounter)
	prometheus.MustRegister(warningEventCounter)
	prometheus.MustRegister(errorEventCounter)
}

func event2Labels(kind AbnormalEventReason, event *v1.Event) []string {
	return []string{
		string(kind),
		event.InvolvedObject.Kind,
		event.InvolvedObject.Name,
		event.InvolvedObject.Namespace,
	}
}

func eventCounterInc(reason, namespace, kind string) {
	normalEventCounter.WithLabelValues(reason, namespace, kind).Inc()
}

func recordErrorEvent(reason AbnormalEventReason, event *v1.Event) {
	labels := event2Labels(reason, event)
	errorEventCounter.WithLabelValues(labels...).Inc()
}

func recordWarningEvent(reason AbnormalEventReason, event *v1.Event) {
	labels := event2Labels(reason, event)
	warningEventCounter.WithLabelValues(labels...).Inc()
}

func triageEvent(event *v1.Event) (AbnormalEventReason, bool) {
	if kind, ok := reasonToEventKind[event.Reason]; ok {
		return kind, true
	}
	if judgements, ok := reasonToEventKindFunc[event.Reason]; ok {
		for _, j := range judgements {
			if j.judge(event) {
				return j.kind, true
			}
		}
	}
	return "", false
}

// RecordEvent records event to prometheus metrics
func RecordEvent(event *v1.Event) {
	if kind, ok := triageEvent(event); ok {
		recordErrorEvent(kind, event)
	} else {
		if event.Type == v1.EventTypeWarning {
			recordWarningEvent(AbnormalEventReason(event.Reason), event)
		} else {
			eventCounterInc(event.Reason, event.Namespace, event.InvolvedObject.Kind)
		}
	}
}
