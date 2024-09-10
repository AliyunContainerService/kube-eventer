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

package influxdb

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"k8s.io/client-go/pkg/version"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

const PingTimeout = time.Second * 5

type InfluxdbClient interface {
	Write(influxdb.BatchPoints) error
	Query(influxdb.Query) (*influxdb.Response, error)
	Ping(time.Duration) (time.Duration, string, error)
	Close() error
}

type InfluxdbConfig struct {
	User                  string
	Password              string
	Secure                bool
	Host                  string
	DbName                string
	WithFields            bool
	InsecureSsl           bool
	RetentionPolicy       string
	ClusterName           string
	DisableCounterMetrics bool
	Concurrency           int
}

func NewClient(c InfluxdbConfig) (InfluxdbClient, error) {
	url := &url.URL{
		Scheme: "http",
		Host:   c.Host,
	}
	if c.Secure {
		url.Scheme = "https"
	}

	iConfig := influxdb.HTTPConfig{
		Addr:               url.String(),
		Username:           c.User,
		Password:           c.Password,
		UserAgent:          fmt.Sprintf("%v/%v", "kube-eventer", version.Get().GitVersion),
		InsecureSkipVerify: c.InsecureSsl,
	}
	client, err := influxdb.NewHTTPClient(iConfig)
	if err != nil {
		return nil, err
	}
	if _, _, err := client.Ping(PingTimeout); err != nil {
		return nil, fmt.Errorf("failed to ping InfluxDB server at %q - %v", c.Host, err)
	}
	return client, nil
}

func BuildConfig(uri *url.URL) (*InfluxdbConfig, error) {
	config := InfluxdbConfig{
		User:                  "root",
		Password:              "root",
		Host:                  "localhost:8086",
		DbName:                "k8s",
		Secure:                false,
		WithFields:            false,
		InsecureSsl:           false,
		RetentionPolicy:       "0",
		ClusterName:           "default",
		DisableCounterMetrics: false,
		Concurrency:           1,
	}

	if len(uri.Host) > 0 {
		config.Host = uri.Host
		if uri.Scheme == "https" {
			config.Secure = true
		} else {
			config.Secure = false
		}
	}
	opts := uri.Query()
	if len(opts["user"]) >= 1 {
		config.User = opts["user"][0]
	}
	// TODO: use more secure way to pass the password.
	if len(opts["pw"]) >= 1 {
		config.Password = opts["pw"][0]
	}
	if len(opts["db"]) >= 1 {
		config.DbName = opts["db"][0]
	}
	if len(opts["retention"]) >= 1 {
		config.RetentionPolicy = opts["retention"][0]
	}
	if len(opts["withfields"]) >= 1 {
		val, err := strconv.ParseBool(opts["withfields"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `withfields` flag - %v", err)
		}
		config.WithFields = val
	}

	if len(opts["insecuressl"]) >= 1 {
		val, err := strconv.ParseBool(opts["insecuressl"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `insecuressl` flag - %v", err)
		}
		config.InsecureSsl = val
	}

	if len(opts["cluster_name"]) >= 1 {
		config.ClusterName = opts["cluster_name"][0]
	}

	if len(opts["disable_counter_metrics"]) >= 1 {
		val, err := strconv.ParseBool(opts["disable_counter_metrics"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `disable_counter_metrics` flag - %v", err)
		}
		config.DisableCounterMetrics = val
	}

	if len(opts["concurrency"]) >= 1 {
		concurrency, err := strconv.Atoi(opts["concurrency"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `concurrency` flag - %v", err)
		}

		if concurrency <= 0 {
			return nil, errors.New("`concurrency` flag can only be positive")
		}

		config.Concurrency = concurrency
	}

	return &config, nil
}
