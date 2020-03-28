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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"

	"github.com/AliyunContainerService/kube-eventer/core"
)

const Name = "WebhookSink"

const (
	retryMaxTimesParamName     = "sinkRetryMaxTimes"
	retryPeriodParamName       = "sinkRetryPeriod"
	retryJitterFactorParamName = "sinkRetryJitterFactor"
)

var (
	RetryMaxTimes     = 3
	RetryPeriod       = time.Second * 1
	RetryJitterFactor = 1.0
)

// Webhook sink usage
// --sink=webhook:https://webhook.example.com/path/to/recive/data/?sinkRetryMaxTimes=3&sinkRetryPeriod=100ms&sinkRetryJitterFactor=1.0
type WebhookSink struct {
	url               *url.URL
	retryMaxTimes     int
	retryPeriod       time.Duration
	retryJitterFactor float64

	client  *http.Client
	stopC   chan struct{}
	stopped bool
	mu      sync.Mutex
}

func NewWebhookSink(webhookURL string) (*WebhookSink, error) {
	u, err := url.Parse(webhookURL)
	if err != nil {
		return nil, err
	}

	w := &WebhookSink{
		url:               u,
		retryMaxTimes:     RetryMaxTimes,
		retryPeriod:       RetryPeriod,
		retryJitterFactor: RetryJitterFactor,
		client:            &http.Client{},
		stopC:             make(chan struct{}),
		stopped:           false,
		mu:                sync.Mutex{},
	}
	return w, nil
}

func (w *WebhookSink) ExportEvents(batch *core.EventBatch) {
	select {
	case <-w.stopC:
		klog.Warning("events will be dropped because of the sink has been stopped")
		return
	default:
	}
	if batch == nil {
		return
	}

	data, err := json.Marshal(batch)
	if err != nil {
		klog.Errorf("marshal the events to json data failed: %s", err)
		return
	}

	if err := w.sendWithRetry(data); err != nil {
		klog.Errorf("send data to webhook url %s failed: %s", w.urlWithoutParams(), err)
	}
}

func (w *WebhookSink) Name() string {
	return Name
}

func (w *WebhookSink) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped {
		return
	}
	close(w.stopC)
	w.stopped = true
}

func (w *WebhookSink) sendWithRetry(data []byte) error {
	var err error
	var retryTimes int
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wait.JitterUntil(func() {
		err = w.send(data)
		if err != nil {
			retryTimes++
		} else {
			cancel()
		}
		if retryTimes > w.retryMaxTimes {
			cancel()
		}
	}, w.retryPeriod, w.retryJitterFactor, true, ctx.Done())

	return err
}

func (w *WebhookSink) send(data []byte) error {
	req, err := http.NewRequest(http.MethodPost, w.url.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"send data failed: the status code of response is %d", resp.StatusCode)
	}
	return nil
}

func (w *WebhookSink) urlWithoutParams() string {
	return fmt.Sprintf("%s://%s%s", w.url.Scheme, w.url.Host, w.url.Path)
}

func CreateWebhookSink(uri *url.URL) (*WebhookSink, error) {
	var err error
	retryMaxTimes := RetryMaxTimes
	retryMaxTimesParam := uri.Query().Get(retryMaxTimesParamName)
	if retryMaxTimesParam != "" {
		retryMaxTimes, err = strconv.Atoi(retryMaxTimesParam)
		if err != nil {
			return nil, fmt.Errorf("invalid value of %s(%s): %s", retryMaxTimesParamName, retryMaxTimesParam, err)
		}
	}
	retryPeriod := RetryPeriod
	retryPeriodParam := uri.Query().Get(retryPeriodParamName)
	if retryPeriodParam != "" {
		retryPeriod, err = time.ParseDuration(retryPeriodParam)
		if err != nil {
			return nil, fmt.Errorf("invalid value of %s(%s): %s", retryPeriodParamName, retryPeriodParam, err)
		}
	}
	retryJitterFactor := RetryJitterFactor
	retryJitterFactorParam := uri.Query().Get(retryJitterFactorParamName)
	if retryJitterFactorParam != "" {
		retryJitterFactor, err = strconv.ParseFloat(retryJitterFactorParam, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value of %s(%s): %s", retryJitterFactorParamName, retryJitterFactorParam, err)
		}
	}

	w, err := NewWebhookSink(uri.String())
	if err != nil {
		return nil, err
	}
	w.retryMaxTimes = retryMaxTimes
	w.retryPeriod = retryPeriod
	w.retryJitterFactor = retryJitterFactor

	return w, nil
}
