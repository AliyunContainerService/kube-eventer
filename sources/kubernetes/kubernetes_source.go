// Copyright 2015 Google Inc. All Rights Reserved.
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

package kubernetes

import (
	metrics "github.com/AliyunContainerService/kube-eventer/metrics/prometheus"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/AliyunContainerService/kube-eventer/common/kubernetes"
	"github.com/AliyunContainerService/kube-eventer/core"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"

	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
)

const (
	// Number of object pointers. Big enough so it won't be hit anytime soon with reasonable GetNewEvents frequency.
	LocalEventsBufferSize = 100000
)

var (
	// Last time of event since unix epoch in seconds
	lastEventTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "eventer",
			Subsystem: "scraper",
			Name:      "last_time_seconds",
			Help:      "Last time of event since unix epoch in seconds.",
		})
	totalEventsNum = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "eventer",
			Subsystem: "scraper",
			Name:      "events_total_number",
			Help:      "The total number of events.",
		})
	scrapEventsDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "eventer",
			Subsystem: "scraper",
			Name:      "duration_milliseconds",
			Help:      "Time spent scraping events in milliseconds.",
		})
)

func init() {
	prometheus.MustRegister(lastEventTimestamp)
	prometheus.MustRegister(totalEventsNum)
	prometheus.MustRegister(scrapEventsDuration)
}

// Implements core.EventSource interface.
type KubernetesEventSource struct {
	// Large local buffer, periodically read.
	localEventsBuffer chan *kubeapi.Event

	stopChannel chan struct{}

	eventClient kubev1core.EventInterface

	exportMetric bool
}

func (this *KubernetesEventSource) GetNewEvents() *core.EventBatch {
	startTime := time.Now()
	defer func() {
		lastEventTimestamp.Set(float64(time.Now().Unix()))
		scrapEventsDuration.Observe(float64(time.Since(startTime)) / float64(time.Millisecond))
	}()
	result := core.EventBatch{
		Timestamp: time.Now(),
		Events:    []*kubeapi.Event{},
	}
	// Get all data from the buffer.
event_loop:
	for {
		select {
		case event := <-this.localEventsBuffer:
			result.Events = append(result.Events, event)
		default:
			break event_loop
		}
	}

	totalEventsNum.Add(float64(len(result.Events)))

	return &result
}

func (this *KubernetesEventSource) watch() {
	// Outer loop, for reconnections.
	for {
		events, err := this.eventClient.List(metav1.ListOptions{Limit: 1})
		if err != nil {
			klog.Errorf("Failed to load events: %v", err)
			time.Sleep(time.Second)
			continue
		}
		// Do not write old events.
		klog.V(9).Infof("kubernetes source watch event. list event first. raw events: %v", events)

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:

				klog.V(10).Infof("kubernetes source watch channel update. watch channel update. watchChanObject: %v", watchUpdate)

				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}
				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Event); ok {

					klog.V(9).Infof("kubernetes source watch event. watch channel update. event: %v", event)

					switch watchUpdate.Type {
					case kubewatch.Added, kubewatch.Modified:
						if this.exportMetric {
							metrics.RecordEvent(event)
						}
						select {
						case this.localEventsBuffer <- event:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func NewKubernetesSource(uri *url.URL, exportMetric bool) (*KubernetesEventSource, error) {
	kubeClient, err := kubernetes.GetKubernetesClient(uri)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client,because of %v", err)
		return nil, err
	}
	eventClient := kubeClient.CoreV1().Events(kubeapi.NamespaceAll)
	result := KubernetesEventSource{
		localEventsBuffer: make(chan *kubeapi.Event, LocalEventsBufferSize),
		stopChannel:       make(chan struct{}),
		eventClient:       eventClient,
		exportMetric:      exportMetric,
	}
	if exportMetric {
		metrics.InitMetrics()
	}
	go result.watch()
	return &result, nil
}
