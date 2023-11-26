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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"net/url"
	"strconv"
	"time"
)

const (
	brokerDialTimeout      = 10 * time.Second
	brokerDialRetryLimit   = 1
	brokerDialRetryWait    = 0
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

func CustomDialer(brokerDialTimeout time.Duration, brokerDialRetryLimit int, brokerDialRetryWait time.Duration) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		var conn net.Conn
		var err error
		for i := 0; i <= brokerDialRetryLimit; i++ {
			conn, err = net.DialTimeout(network, addr, brokerDialTimeout)
			if err == nil {
				return conn, nil
			}
			if i < brokerDialRetryLimit {
				time.Sleep(brokerDialRetryWait)
			}
		}
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
}

func getTlsConfiguration(opts url.Values) (*tls.Config, bool, error) {
	if len(opts["cacert"]) == 0 &&
		(len(opts["cert"]) == 0 || len(opts["key"]) == 0) {
		return nil, false, nil
	}

	t := &tls.Config{}
	if len(opts["cacert"]) != 0 {
		caFile := opts["cacert"][0]
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, false, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		t.RootCAs = caCertPool
	}

	if len(opts["cert"]) != 0 && len(opts["key"]) != 0 {
		certFile := opts["cert"][0]
		keyFile := opts["key"][0]
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, false, err
		}
		t.Certificates = []tls.Certificate{cert}
	}

	if len(opts["insecuressl"]) != 0 {
		insecuressl := opts["insecuressl"][0]
		insecure, err := strconv.ParseBool(insecuressl)
		if err != nil {
			return nil, false, err
		}
		t.InsecureSkipVerify = insecure
	}

	return t, true, nil
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

	TLSClientConfig, TLSClientConfigEnable, err := getTlsConfiguration(opts)

	var config = amqp.Config{}

	if TLSClientConfigEnable {
		config.TLSClientConfig = TLSClientConfig
	}

	config.Dial = CustomDialer(brokerDialTimeout, brokerDialRetryLimit, brokerDialRetryWait)

	conn, err := amqp.DialConfig(GetRabbitMQURL(opts), config)
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
