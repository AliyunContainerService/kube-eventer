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

package webhook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/stretchr/testify/assert"
	kube_api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SampleWebhookServerForTesting struct {
	events []*core.EventBatch
	codes  []int

	server *httptest.Server
	mux    sync.Mutex
}

func newSampleWebhookServer() *SampleWebhookServerForTesting {
	s := &SampleWebhookServerForTesting{
		events: []*core.EventBatch{},
		codes:  []int{},
		mux:    sync.Mutex{},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		e := &core.EventBatch{}
		json.Unmarshal(data, e)
		s.pushEvent(e)

		w.WriteHeader(s.getCode())
	})

	s.server = httptest.NewServer(handler)
	return s
}

func (s *SampleWebhookServerForTesting) pushEvent(e *core.EventBatch) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.events = append(s.events, e)
}

func (s *SampleWebhookServerForTesting) pushCodes(cs ...int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.codes = append(s.codes, cs...)
}

func (s *SampleWebhookServerForTesting) getCode() int {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.codes) == 0 {
		return 200
	}
	c := s.codes[0]
	s.codes = append([]int{}, s.codes[1:len(s.codes)]...)
	return c
}

func TestWebhookSink_Name(t *testing.T) {
	w, _ := NewWebhookSink("https://webhook.example.com")
	assert.Equal(t, w.Name(), Name)
}

func TestNewWebhookSink_error(t *testing.T) {
	w, err := NewWebhookSink("://webhook.example.com")
	assert.Error(t, err)
	assert.Nil(t, w)
}

func TestWebhookSink_Stop(t *testing.T) {
	w, _ := NewWebhookSink("https://webhook.example.com")
	assert.False(t, w.stopped)

	w.Stop()
	assert.True(t, w.stopped)
	w.Stop()
	assert.True(t, w.stopped)
}

func TestWebhookSink_ExportEvents_no_retry(t *testing.T) {
	s := newSampleWebhookServer()
	defer s.server.Close()
	u, _ := url.Parse(s.server.URL)

	w, _ := CreateWebhookSink(u)
	w.retryMaxTimes = 3
	e := &core.EventBatch{
		Timestamp: time.Now(),
		Events: []*kube_api.Event{
			&kube_api.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "event_1",
				},
			},
			&kube_api.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "event_2",
				},
			},
		},
	}
	w.ExportEvents(e)

	jsonValueEqual(t, []*core.EventBatch{e}, s.events)
}

func TestWebhookSink_ExportEvents_retry_once(t *testing.T) {
	s := newSampleWebhookServer()
	defer s.server.Close()
	u, _ := url.Parse(s.server.URL)

	w, _ := CreateWebhookSink(u)
	w.retryMaxTimes = 3
	w.retryPeriod = time.Millisecond
	e := &core.EventBatch{
		Timestamp: time.Now(),
		Events: []*kube_api.Event{
			&kube_api.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "event_1",
				},
			},
		},
	}
	s.pushCodes(400)
	w.ExportEvents(e)

	jsonValueEqual(t, []*core.EventBatch{e, e}, s.events)
}

func TestWebhookSink_ExportEvents_retry_max(t *testing.T) {
	s := newSampleWebhookServer()
	defer s.server.Close()
	u, _ := url.Parse(s.server.URL)

	w, _ := CreateWebhookSink(u)
	w.retryMaxTimes = 3
	w.retryPeriod = time.Millisecond
	e := &core.EventBatch{
		Timestamp: time.Now(),
		Events: []*kube_api.Event{
			&kube_api.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "event_1",
				},
			},
		},
	}
	s.pushCodes(http.StatusBadRequest, http.StatusBadGateway, http.StatusForbidden,
		http.StatusConflict)
	w.ExportEvents(e)

	jsonValueEqual(t, []*core.EventBatch{e, e, e, e}, s.events)
}

func TestWebhookSink_ExportEvents_after_stop(t *testing.T) {
	s := newSampleWebhookServer()
	defer s.server.Close()
	u, _ := url.Parse(s.server.URL)

	w, _ := CreateWebhookSink(u)
	w.Stop()
	e := &core.EventBatch{
		Timestamp: time.Now(),
		Events: []*kube_api.Event{
			&kube_api.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "event_1",
				},
			},
		},
	}
	w.ExportEvents(e)

	assert.Equal(t, 0, len(s.events))
}

func jsonValueEqual(t *testing.T, expected interface{}, actual interface{}) {
	a, _ := json.MarshalIndent(expected, "", " ")
	b, _ := json.MarshalIndent(actual, "", " ")
	assert.JSONEq(t, string(a), string(b))
}

func TestCreateWebhookSink_param_error(t *testing.T) {
	tests := []struct {
		name   string
		uri    string
		errmsg string
	}{
		{
			name:   "no error",
			uri:    "https://webhook.example.com/",
			errmsg: "",
		},
		{
			name:   "invalid sinkRetryMaxTimes",
			uri:    "https://webhook.example.com/?sinkRetryMaxTimes=abc",
			errmsg: "invalid value of sinkRetryMaxTimes(abc)",
		},
		{
			name:   "invalid sinkRetryPeriod",
			uri:    "https://webhook.example.com/?sinkRetryPeriod=abc",
			errmsg: "invalid value of sinkRetryPeriod(abc)",
		},
		{
			name:   "invalid sinkRetryJitterFactor",
			uri:    "https://webhook.example.com/?sinkRetryJitterFactor=abc",
			errmsg: "invalid value of sinkRetryJitterFactor(abc)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse(tt.uri)
			got, err := CreateWebhookSink(u)
			if tt.errmsg == "" {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errmsg)
			}
		})
	}
}

func TestCreateWebhookSink_with_param(t *testing.T) {
	u, _ := url.Parse("https://webhook.example.com/?sinkRetryMaxTimes=5&sinkRetryPeriod=10ms&sinkRetryJitterFactor=1.2")
	w, err := CreateWebhookSink(u)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	assert.Equal(t, w.retryMaxTimes, 5)
	assert.Equal(t, w.retryPeriod, 10*time.Millisecond)
	assert.Equal(t, w.retryJitterFactor, 1.2)
	assert.Equal(t, w.urlWithoutParams(), "https://webhook.example.com/")
}

func TestCreateWebhookSink_without_param(t *testing.T) {
	u, _ := url.Parse("https://webhook.example.com/?a=b&c=d&e=2")
	w, err := CreateWebhookSink(u)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	assert.Equal(t, w.retryMaxTimes, RetryMaxTimes)
	assert.Equal(t, w.retryPeriod, RetryPeriod)
	assert.Equal(t, w.retryJitterFactor, RetryJitterFactor)
}
