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
	PodPending            AbnormalEventKind = "pod_pending"
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
type RecordFunc func(kind AbnormalEventKind, event *v1.Event)
type CleanFunc func(kind AbnormalEventKind, event *v1.Event)

var DefaultRecordFunc RecordFunc = recordAbnormalEvent
var DefaultCleanFunc CleanFunc = cleanAbnormalEvent

type JudgeEvent struct {
	kind   AbnormalEventKind
	judge  JudgeEventFunc
	record RecordFunc
	clean  CleanFunc
}

var (
	eventCounter         *prometheus.CounterVec
	abnormalEventCounter *prometheus.CounterVec
	abnormalEventTime    *prometheus.GaugeVec
	abnormalEventLastTS  *prometheus.GaugeVec

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
		"ImagePullBackOff": {{kind: PodImagePullBackOff, judge: isPodImagePullBackOff}},
		"Scheduled":        {{kind: PodPending, judge: isPodPending, record: recordPodPending, clean: Noop}},
		"Pulling":          {{kind: PodPending, judge: isPodPendingClear, record: recordPodPendingClear, clean: Noop}},
		"Created":          {{kind: PodPending, judge: isPodPendingClear, record: recordPodPendingClear, clean: Noop}},
		"Failed":           {{kind: PodFailStart, judge: isPodFailStart}},
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
	eventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eventer",
			Subsystem: "events",
			Name:      "total",
		},
		[]string{"reason", "type"},
	)
	errorEventLabels := []string{
		"event_kind",
		"reason",
		"type",
		"involved_object_kind",
		"involved_object_api_version",
		"involved_object_name",
		"involved_object_namespace",
		"involved_object_resource_version",
		"involved_object_field_path",
		"related_kind",
		"related_api_version",
		"related_name",
		"related_namespace",
		"related_resource_version",
		"related_field_path",
	}
	abnormalEventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "eventer",
		Subsystem: "events",
		Name:      "abnormal_count",
	}, errorEventLabels)
	abnormalEventTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "eventer",
		Subsystem: "events",
		Name:      "abnormal_time_seconds",
	}, errorEventLabels)
	abnormalEventLastTS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "eventer",
		Subsystem: "events",
		Name:      "abnormal_last_ts_seconds",
	}, errorEventLabels)
	prometheus.MustRegister(eventCounter)
	prometheus.MustRegister(abnormalEventCounter)
	prometheus.MustRegister(abnormalEventTime)
	prometheus.MustRegister(abnormalEventLastTS)
}

func event2Labels(kind AbnormalEventKind, event *v1.Event) []string {
	related := event.Related
	if related == nil {
		related = &v1.ObjectReference{}
	}
	return []string{
		string(kind),
		event.Reason,
		event.Type,
		event.InvolvedObject.Kind,
		event.InvolvedObject.APIVersion,
		event.InvolvedObject.Name,
		event.InvolvedObject.Namespace,
		event.InvolvedObject.ResourceVersion,
		event.InvolvedObject.FieldPath,
		related.Kind,
		related.APIVersion,
		related.Name,
		related.Namespace,
		related.ResourceVersion,
		related.FieldPath,
	}
}

func eventCounterInc(reason, eventType string) {
	eventCounter.WithLabelValues(reason, eventType).Inc()
}

func cleanAbnormalEvent(kind AbnormalEventKind, event *v1.Event) {
	labels := event2Labels(kind, event)
	abnormalEventCounter.DeleteLabelValues(labels...)
	abnormalEventTime.DeleteLabelValues(labels...)
	abnormalEventLastTS.DeleteLabelValues(labels...)
}

func triageEvent(event *v1.Event) (AbnormalEventKind, RecordFunc, CleanFunc, bool) {
	if kind, ok := reasonToEventKind[event.Reason]; ok {
		return kind, DefaultRecordFunc, DefaultCleanFunc, true
	}
	if judgements, ok := reasonToEventKindFunc[event.Reason]; ok {
		for _, j := range judgements {
			if j.judge(event) {
				record, clean := j.record, j.clean
				if record == nil {
					record = DefaultRecordFunc
				}
				if clean == nil {
					clean = DefaultCleanFunc
				}
				return j.kind, record, clean, true
			}
		}
	}
	return "", nil, nil, false
}

// RecordEvent records event to prometheus metrics
func RecordEvent(event *v1.Event) {
	eventCounterInc(event.Reason, event.Type)
	if kind, record, _, ok := triageEvent(event); ok {
		record(kind, event)
	}
}

// CleanEvent cleans event from prometheus metrics
func CleanEvent(event *v1.Event) {
	if kind, _, clean, ok := triageEvent(event); ok {
		clean(kind, event)
	}
}
