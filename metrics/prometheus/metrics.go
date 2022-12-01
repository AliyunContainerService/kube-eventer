package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
)

type AbnormalEventKind string

var (
	// Namespace Lavel Event
	PodEvict              AbnormalEventKind = "pod_evict"
	PodImagePullBackOff   AbnormalEventKind = "pod_image_pull_back_off"
	PodOOM                AbnormalEventKind = "pod_oom"
	ResourceInsufficient  AbnormalEventKind = "resource_insufficient"
	PodFailStart          AbnormalEventKind = "pod_fail_start"
	PodCrash              AbnormalEventKind = "pod_crash"
	PodFailScheduling     AbnormalEventKind = "pod_fail_scheduling"
	DiskProvisionFailSize AbnormalEventKind = "disk_provision_fail_size"
	DiskProvisionFail     AbnormalEventKind = "disk_provision_fail"
	VolumeFailMount       AbnormalEventKind = "volume_fail_mount"

	// Node Level Event
	NodeOOM           AbnormalEventKind = "node_oom"
	NodeRebooted      AbnormalEventKind = "node_rebooted"
	NodeDiskPressure  AbnormalEventKind = "node_disk_pressure"
	NodeDockerHung    AbnormalEventKind = "node_docker_hung"
	NodePSHung        AbnormalEventKind = "node_ps_hung"
	NodeGPUXIPError   AbnormalEventKind = "node_gpu_xip_error"
	NodeFDPressure    AbnormalEventKind = "node_fd_pressure"
	NodePLEGUnhealthy AbnormalEventKind = "node_pleg_unhealthy"
	NodeNPTDown       AbnormalEventKind = "node_npt_down"
	NodeNotReady      AbnormalEventKind = "node_not_ready"
	ConnTrackFull     AbnormalEventKind = "conntrack_full"

	// Core Component Event
	CcmSLBSyncFail          AbnormalEventKind = "ccm_slb_sync_fail"
	CcmSLBUnavailable       AbnormalEventKind = "ccm_slb_unavailable"
	CcmSLBDeleteFail        AbnormalEventKind = "ccm_slb_delete_fail"
	CcmCreateRouteFail      AbnormalEventKind = "ccm_create_route_fail"
	CcmSyncRouteFail        AbnormalEventKind = "ccm_sync_route_fail"
	CcmAddNodeFail          AbnormalEventKind = "ccm_add_node_fail"
	CcmDeleteNodeFail       AbnormalEventKind = "ccm_delete_node_fail"
	CcmSLBAnnotationChanged AbnormalEventKind = "ccm_slb_annotation_changed"
	CcmSLBSpecChanged       AbnormalEventKind = "ccm_slb_spec_changed"
	CSISlowIO               AbnormalEventKind = "csi_slow_io"
	CSIDeviceBusy           AbnormalEventKind = "csi_device_busy"
	CSIIOHang               AbnormalEventKind = "csi_io_hang"
	CNIAllocIPFail          AbnormalEventKind = "cni_alloc_ip_fail"
	CNIAllocResourceFail    AbnormalEventKind = "cni_alloc_resource_fail"
	CNIResourceInvalid      AbnormalEventKind = "cni_resource_invalid"
	CNIParseFail            AbnormalEventKind = "cni_parse_fail"
	CNIDisposeResourceFail  AbnormalEventKind = "cni_dispose_resource_fail"
	ClusterIPNotEnough      AbnormalEventKind = "cluster_ip_not_enough"
)

type JudgeEventFunc func(event *v1.Event) bool

type JudgeEvent struct {
	kind  AbnormalEventKind
	judge JudgeEventFunc
}

var (
	normalEventCounter   *prometheus.CounterVec
	abnormalEventCounter *prometheus.CounterVec

	reasonToEventKind = map[string]AbnormalEventKind{
		// Namespace level
		"Evicted":            PodEvict,
		"OOMKilling":         PodOOM,
		"PodOOMKilling":      PodOOM,
		"FailedMount":        VolumeFailMount,
		"FailedAttachVolume": VolumeFailMount,
		"FailedScheduling":   PodFailScheduling,

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
		"AllocResourceFailed":      CNIAllocResourceFail,
		"ResourceInvalid":          CNIResourceInvalid,
		"ParseFailed":              CNIParseFail,
		"DisposeResourceFailed":    CNIDisposeResourceFail,
	}

	reasonToEventKindFunc = map[string][]JudgeEvent{
		"Failed": {
			{kind: PodFailStart, judge: isPodFailStart},
			{kind: PodImagePullBackOff, judge: isPodImagePullBackOff},
		},
		"BackOff":          {{kind: PodCrash, judge: isPodCrash}},
		"FailedScheduling": {{kind: ResourceInsufficient, judge: isResourceInsufficient}},
		"ProvisioningFailed": {
			{kind: DiskProvisionFailSize, judge: isDiskProvisionFailSize},
			{kind: DiskProvisionFail, judge: isDiskProvisionFail},
		},
		"NodeNotReady": {
			{kind: NodePLEGUnhealthy, judge: isNodePLEGUnhealthy},
			{kind: NodeNotReady, judge: isNodeNotReady},
		},
		"AllocResourceFailed": {{kind: ClusterIPNotEnough, judge: isClusterIPNotEnough}},
		"ResourceInvalid":     {{kind: ClusterIPNotEnough, judge: isClusterIPNotEnough}},
	}
)

func init() {
	normalEventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eventer",
			Subsystem: "events",
			Name:      "normal_total",
		},
		[]string{"reason", "namespace"},
	)
	errorEventLabels := []string{
		"event_kind",
		"kind",
		"name",
		"namespace",
	}
	abnormalEventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "eventer",
		Subsystem: "events",
		Name:      "warning_total",
	}, errorEventLabels)
	prometheus.MustRegister(normalEventCounter)
	prometheus.MustRegister(abnormalEventCounter)
}

func event2Labels(kind AbnormalEventKind, event *v1.Event) []string {
	return []string{
		string(kind),
		event.InvolvedObject.Kind,
		event.InvolvedObject.Name,
		event.InvolvedObject.Namespace,
	}
}

func eventCounterInc(reason, namespace string) {
	normalEventCounter.WithLabelValues(reason, namespace).Inc()
}

func cleanAbnormalEvent(kind AbnormalEventKind, event *v1.Event) {
	labels := event2Labels(kind, event)
	abnormalEventCounter.DeleteLabelValues(labels...)
}

func recordAbnormalEvent(kind AbnormalEventKind, event *v1.Event) {
	labels := event2Labels(kind, event)
	abnormalEventCounter.WithLabelValues(labels...).Inc()
}

func triageEvent(event *v1.Event) (AbnormalEventKind, bool) {
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
		recordAbnormalEvent(kind, event)
	} else {
		eventCounterInc(event.Reason, event.Namespace)
	}
}

// CleanEvent cleans event from prometheus metrics
func CleanEvent(event *v1.Event) {
	if kind, ok := triageEvent(event); ok {
		cleanAbnormalEvent(kind, event)
	}
}
