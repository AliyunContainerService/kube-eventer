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

package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"k8s.io/klog"
	"net/url"
	"time"
)

type PulsarClient interface {
	Name() string
	Stop()
	ProducePulsarMessage(msgData interface{}) error
}

type pulsarSink struct {
	topic    string
	producer pulsar.Producer
}

func (p *pulsarSink) Name() string {
	return "Apache Pulsar Sink"
}

func (p *pulsarSink) Stop() {
	p.producer.Close()

}

func (p *pulsarSink) ProducePulsarMessage(msgData interface{}) error {
	start := time.Now()
	msgJson, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("failed to transform the items to json : %s", err)
	}
	send, err := p.producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload:    msgJson,
		Properties: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to produce message to Pulsar: %s", err)
	}
	end := time.Now()
	klog.V(4).Infof("Exported %d data to Pulsar in %s, messageID: %s", len(msgJson), end.Sub(start), send.String())
	return nil
}

func NewPulsarClient(uri *url.URL) (PulsarClient, error) {
	opts, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url's query string: %s", err)
	}
	klog.V(3).Infof("Pulsar opts: %v", opts)

	var (
		serviceURL   []string
		token, topic string
		client       pulsar.Client
	)
	if len(opts["serviceurl"]) < 1 {
		return nil, fmt.Errorf("there is no broker assigned for connecting Pulsar")
	}
	serviceURL = append(serviceURL, opts["serviceurl"]...)

	if len(opts["eventstopic"]) != 1 {
		return nil, fmt.Errorf("there is no topic assigned for connecting Pulsar")
	}
	topic = opts["eventstopic"][0]

	if len(opts["token"]) > 0 {
		token = opts["token"][0]
	}

	if len(token) == 0 {
		client, err = pulsar.NewClient(pulsar.ClientOptions{
			URL: serviceURL[0],
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Pulsar client: %v", err)
		}
	} else {
		client, err = pulsar.NewClient(pulsar.ClientOptions{
			URL:            serviceURL[0],
			Authentication: pulsar.NewAuthenticationToken(token),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Pulsar client: %v", err)
		}
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Name:  "kube-eventer",
		Topic: topic,
	})

	return &pulsarSink{
		producer: producer,
	}, nil
}
