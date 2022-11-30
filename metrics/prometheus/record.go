package prometheus

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

var (
	pendingTimers map[string]*time.Timer
)

func init() {
	pendingTimers = make(map[string]*time.Timer)
}

func recordAbnormalEvent(kind AbnormalEventKind, event *v1.Event) {
	labels := event2Labels(kind, event)
	abnormalEventCounter.WithLabelValues(labels...).Inc()
}

func recordPodPending(kind AbnormalEventKind, event *v1.Event) {
	key := string(event.InvolvedObject.UID)
	labels := event2Labels(kind, event)
	timer := time.AfterFunc(5*time.Minute, func() {
		abnormalEventCounter.WithLabelValues(labels...).Inc()
		delete(pendingTimers, key)
	})
	pendingTimers[key] = timer
}

func recordPodPendingClear(kind AbnormalEventKind, event *v1.Event) {
	// Pulling is the first event after scheduled. In case image has already been on node, detect Created together.
	key := string(event.InvolvedObject.UID)
	labels := event2Labels(kind, event)
	if timer, ok := pendingTimers[key]; ok {
		timer.Stop()
		delete(pendingTimers, key)
	}
	abnormalEventCounter.DeleteLabelValues(labels...)
}
func Noop(_ AbnormalEventKind, _ *v1.Event) {}
