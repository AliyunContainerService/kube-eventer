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
	"net/url"
	"time"

	kubeconfig "github.com/AliyunContainerService/kube-eventer/common/kubernetes"
	"github.com/AliyunContainerService/kube-eventer/core"

	"github.com/prometheus/client_golang/prometheus"
	kubeapi "k8s.io/api/core/v1"
	"k8s.io/api/events/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	kubeclient "k8s.io/client-go/kubernetes"
	clientCore "k8s.io/client-go/kubernetes/typed/core/v1"
	clientEvents "k8s.io/client-go/kubernetes/typed/events/v1beta1"
	"k8s.io/klog"
)

const (
	// LocalEventsBufferSize number of object pointers.
	// big enough so it won't be hit anytime soon with reasonable GetNewEvents frequency
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

// EventSource implements core.EventSource interface.
type EventSource struct {
	// Large local buffer, periodically read.
	localEventsBuffer chan *kubeapi.Event

	stopChannel chan struct{}

	eventClient    clientCore.EventInterface
	eventNewClient clientEvents.EventInterface
}

// GetNewEvents returns all the received events and flushes the buffer
func (k8ssrc *EventSource) GetNewEvents() *core.EventBatch {
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
		case event := <-k8ssrc.localEventsBuffer:
			result.Events = append(result.Events, event)
		default:
			break event_loop
		}
	}

	totalEventsNum.Add(float64(len(result.Events)))

	return &result
}

func (k8ssrc *EventSource) watch() {
	// Outer loop, for reconnections.
	var watcher kubewatch.Interface
	var err error
	var meta metav1.ListInterface
	var event *kubeapi.Event
	var newEvent *v1beta1.Event
	useNewAPI := k8ssrc.eventNewClient != nil

	for {
		if useNewAPI {
			meta, err = k8ssrc.eventNewClient.List(metav1.ListOptions{})
		} else {
			meta, err = k8ssrc.eventClient.List(metav1.ListOptions{})
		}
		if err != nil {
			klog.Errorf("Failed to load events: %v", err)
			time.Sleep(time.Second)
			continue
		}
		// Do not write old events.
		if useNewAPI {
			watcher, err = k8ssrc.eventNewClient.Watch(metav1.ListOptions{
				Watch:           true,
				ResourceVersion: meta.GetResourceVersion(),
			})
		} else {
			watcher, err = k8ssrc.eventClient.Watch(
				metav1.ListOptions{
					Watch:           true,
					ResourceVersion: meta.GetResourceVersion()})
		}
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
				if useNewAPI {
					newEvent, ok = watchUpdate.Object.(*v1beta1.Event)
					event = convertFromV1beta1(newEvent)
				} else {
					event, ok = watchUpdate.Object.(*kubeapi.Event)
				}
				if ok {
					switch watchUpdate.Type {
					case kubewatch.Added, kubewatch.Modified:
						select {
						case k8ssrc.localEventsBuffer <- event:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-k8ssrc.stopChannel:
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

// NewKubernetesSource connects to the cluster, sets up a watch an returns a running source of events
func NewKubernetesSource(uri *url.URL, useEventsAPI bool) (*EventSource, error) {
	var eventClient clientCore.EventInterface
	var eventNewClient clientEvents.EventInterface
	kubeConfig, err := kubeconfig.GetKubeClientConfig(uri)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubeclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	if useEventsAPI {
		eventNewClient = kubeClient.EventsV1beta1().Events(kubeapi.NamespaceAll)
	} else {
		eventClient = kubeClient.CoreV1().Events(kubeapi.NamespaceAll)
	}
	result := EventSource{
		localEventsBuffer: make(chan *kubeapi.Event, LocalEventsBufferSize),
		stopChannel:       make(chan struct{}),
		eventClient:       eventClient,
		eventNewClient:    eventNewClient,
	}
	go result.watch()
	return &result, nil
}

func convertFromV1beta1(event *v1beta1.Event) *kubeapi.Event {
	var series *kubeapi.EventSeries
	if event.Series != nil {
		series = &kubeapi.EventSeries{
			Count:            event.Series.Count,
			LastObservedTime: event.Series.LastObservedTime,
			State:            kubeapi.EventSeriesState(event.Series.State),
		}
	} else {
		series = nil
	}
	return &kubeapi.Event{
		TypeMeta:            event.TypeMeta,
		ObjectMeta:          event.ObjectMeta,
		InvolvedObject:      event.Regarding,
		Reason:              event.Reason,
		Message:             event.Note,
		Source:              event.DeprecatedSource,
		FirstTimestamp:      event.DeprecatedFirstTimestamp,
		LastTimestamp:       event.DeprecatedLastTimestamp,
		Count:               event.DeprecatedCount,
		Type:                event.Type,
		EventTime:           event.EventTime,
		Series:              series,
		Action:              event.Action,
		Related:             event.Related,
		ReportingController: event.ReportingController,
		ReportingInstance:   event.ReportingInstance,
	}
}
