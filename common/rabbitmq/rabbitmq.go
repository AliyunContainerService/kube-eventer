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

package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"k8s.io/klog"
	"net/url"
)

const (
	metricsTopic           = "heapster-metrics"
	eventsTopic            = "heapster-events"
)

const (
	TimeSeriesTopic = "timeseriestopic"
	EventsTopic     = "eventstopic"
)

type AmqpClient interface {
	Name() string
	Stop()
	ProduceAmqpMessage(msgData interface{}) error
}

type amqpSink struct {
	producer  *amqp.Channel
	dataTopic string
}

func (sink *amqpSink) ProduceAmqpMessage(msgData interface{}) error {
	return nil
}

func (sink *amqpSink) Name() string {
	return "Apache Amqp Sink"
}

func (sink *amqpSink) Stop() {
	err := sink.producer.Close()
	if err != nil {
		return
	}
}

func getTopic(opts map[string][]string, topicType string) (string, error) {
	var topic string
	switch topicType {
	case TimeSeriesTopic:
		topic = metricsTopic
	case EventsTopic:
		topic = eventsTopic
	default:
		return "", fmt.Errorf("Topic type '%s' is illegal.", topicType)
	}

	if len(opts[topicType]) > 0 {
		topic = opts[topicType][0]
	}

	return topic, nil
}

func getOptionsWithoutSecrets(values url.Values) string {
	var password []string
	if len(values["password"]) != 0 {
		password = values["password"]
		values["password"] = []string{"***"}
		defer func() { values["password"] = password }()
	}
	options := fmt.Sprintf("amqp sink option: %v", values)
	return options
}

func GetRabbitMQURL(values url.Values) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", values["username"], values["password"], values["host"], values["port"])
}

func NewAmqpClient(uri *url.URL, topicType string) (AmqpClient, error) {
	opts, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url's query string: %s", err)
	}
	klog.V(3).Info(getOptionsWithoutSecrets(opts))

	topic, err := getTopic(opts, topicType)
	if err != nil {
		return nil, err
	}

	amqp.Logger = GologAdapterLogger{}

	conn, err := amqp.Dial(GetRabbitMQURL(opts))
	if err != nil {
		return nil, err
	}

	klog.V(3).Infof("attempting to setup amqp sink")
	sinkProducer, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to setup Producer: - %v", err)
	}
	klog.V(3).Infof("amqp sink setup successfully")

	return &amqpSink{
		producer:  sinkProducer,
		dataTopic: topic,
	}, nil
}
