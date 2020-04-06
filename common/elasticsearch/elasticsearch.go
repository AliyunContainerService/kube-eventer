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
package elasticsearch

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"k8s.io/klog"
)

const (
	ESIndex       = "heapster"
	ESClusterName = "default"
)

type UnsupportedVersion struct{}

func (UnsupportedVersion) Error() string {
	return "Unsupported ElasticSearch Client Version"
}

type elasticWrapper interface {
	IndexExists(indices ...string) (bool, error)
	CreateIndex(name string, mapping string) (bool, error)
	AddAlias(index string, alias string) (bool, error)
	HasAlias(index string, alias string) (bool, error)
	AddBulkReq(index, typeName string, data interface{}) error
	ErrorStats() int64
	FlushBulk() error
}

type ElasticConfig struct {
	Url         []string
	User        string
	Secret      string
	MaxRetries  *int
	HealthCheck *bool
	Timeout     *time.Duration
	HttpClient  *http.Client
	Sniff       *bool
}

type ElasticSearchService struct {
	EsClient     elasticWrapper
	baseIndex    string
	ClusterName  string
	UseNamespace bool
}

func (esSvc *ElasticSearchService) Index(date time.Time, namespace string) string {
	dateStr := date.Format("2006.01.02")
	if len(namespace) > 0 {
		return fmt.Sprintf("%s-%s-%s", esSvc.baseIndex, namespace, dateStr)
	}
	return fmt.Sprintf("%s-%s", esSvc.baseIndex, dateStr)
}

func (esSvc *ElasticSearchService) IndexAlias(typeName string) string {
	return fmt.Sprintf("%s-%s", esSvc.baseIndex, typeName)
}

func (esSvc *ElasticSearchService) FlushData() error {
	return esSvc.EsClient.FlushBulk()
}

func (esSvc *ElasticSearchService) ErrorStats() int64 {
	return esSvc.EsClient.ErrorStats()
}

// SaveDataIntoES save metrics and events to ES by using ES client
func (esSvc *ElasticSearchService) SaveData(date time.Time, typeName string, namespace string, sinkData []interface{}) error {
	if typeName == "" || len(sinkData) == 0 {
		return nil
	}

	indexName := esSvc.Index(date, namespace)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := esSvc.EsClient.IndexExists(indexName)
	if err != nil {
		return err
	}

	if !exists {
		// Create a new index.
		ack, err := esSvc.EsClient.CreateIndex(indexName, mapping)
		if err != nil {
			return err
		}

		if !ack {
			return errors.New("Failed to acknoledge index creation")
		}
	}

	aliasName := esSvc.IndexAlias(typeName)

	hasAlias, err := esSvc.EsClient.HasAlias(indexName, aliasName)
	if err != nil {
		return err
	}

	if !hasAlias {
		ack, err := esSvc.EsClient.AddAlias(indexName, esSvc.IndexAlias(typeName))
		if err != nil {
			return err
		}

		if !ack {
			return errors.New("Failed to acknoledge index alias creation")
		}
	}

	for _, data := range sinkData {
		esSvc.EsClient.AddBulkReq(indexName, typeName, data)
	}

	return nil
}

// CreateElasticSearchConfig creates an ElasticSearch configuration struct
// which contains an ElasticSearch client for later use
func CreateElasticSearchService(uri *url.URL) (*ElasticSearchService, error) {

	var esSvc ElasticSearchService
	opts, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse url's query string: %s", err)
	}

	version := 5
	if len(opts["ver"]) > 0 {
		version, err = strconv.Atoi(opts["ver"][0])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse URL's version value into an int: %v", err)
		}
	}

	esSvc.ClusterName = ESClusterName
	if len(opts["cluster_name"]) > 0 {
		esSvc.ClusterName = opts["cluster_name"][0]
	}

	// set the index for es,the default value is "heapster"
	esSvc.baseIndex = ESIndex
	if len(opts["index"]) > 0 {
		esSvc.baseIndex = opts["index"][0]
	}

	if len(opts["use_namespace"]) > 0 {
		esSvc.UseNamespace = true
	}
	var config ElasticConfig

	// Set the URL endpoints of the ES's nodes. Notice that when sniffing is
	// enabled, these URLs are used to initially sniff the cluster on startup.
	if len(opts["nodes"]) > 0 {
		config.Url = opts["nodes"]
	} else if uri.Scheme != "" && uri.Host != "" {
		config.Url = []string{uri.Scheme + "://" + uri.Host}
	} else {
		return nil, errors.New("There is no node assigned for connecting ES cluster")
	}

	// If the ES cluster needs authentication, the username and secret
	// should be set in sink config.Else, set the Authenticate flag to false
	if len(opts["esUserName"]) > 0 && len(opts["esUserSecret"]) > 0 {
		config.User = opts["esUserName"][0]
		config.Secret = opts["esUserSecret"][0]
	}

	if len(opts["maxRetries"]) > 0 {
		maxRetries, err := strconv.Atoi(opts["maxRetries"][0])
		if err != nil {
			return nil, errors.New("Failed to parse URL's maxRetries value into an int")
		}
		config.MaxRetries = &maxRetries
	}

	if len(opts["healthCheck"]) > 0 {
		healthCheck, err := strconv.ParseBool(opts["healthCheck"][0])
		if err != nil {
			return nil, errors.New("Failed to parse URL's healthCheck value into a bool")
		}
		config.HealthCheck = &healthCheck
	}

	if len(opts["startupHealthcheckTimeout"]) > 0 {
		timeout, err := time.ParseDuration(opts["startupHealthcheckTimeout"][0] + "s")
		if err != nil {
			return nil, fmt.Errorf("Failed to parse URL's startupHealthcheckTimeout: %s", err.Error())
		}
		config.Timeout = &timeout
	}

	if useSigV4(opts) {
		klog.Info("Configuring with AWS credentials..")

		awsClient, err := createAWSClient()
		sniffDisable := false
		if err != nil {
			return nil, err
		}
		config.HttpClient = awsClient
		config.Sniff = &sniffDisable
	} else {
		if len(opts["sniff"]) > 0 {
			sniff, err := strconv.ParseBool(opts["sniff"][0])
			if err != nil {
				return nil, errors.New("Failed to parse URL's sniff value into a bool")
			}
			config.Sniff = &sniff
		}
	}

	bulkWorkers := 5
	if len(opts["bulkWorkers"]) > 0 {
		bulkWorkers, err = strconv.Atoi(opts["bulkWorkers"][0])
		if err != nil {
			return nil, errors.New("Failed to parse URL's bulkWorkers value into an int")
		}
	}

	pipeline := ""
	if len(opts["pipeline"]) > 0 {
		pipeline = opts["pipeline"][0]
	}

	switch version {
	case 2:
		esSvc.EsClient, err = NewEsClient2(config, bulkWorkers)
	case 5:
		esSvc.EsClient, err = NewEsClient5(config, bulkWorkers, pipeline)
	case 6:
		esSvc.EsClient, err = NewEsClient6(config, bulkWorkers, pipeline)
	case 7:
		esSvc.EsClient, err = NewEsClient7(config, bulkWorkers, pipeline)
	default:
		return nil, UnsupportedVersion{}
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to create ElasticSearch client: %v", err)
	}

	klog.V(2).Infof("ElasticSearch sink configure successfully")

	return &esSvc, nil
}
