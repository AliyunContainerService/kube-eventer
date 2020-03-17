package elasticsearch

import (
	"fmt"
	"k8s.io/klog"
	"time"

	"github.com/pborman/uuid"
	elastic2 "gopkg.in/olivere/elastic.v3"
)

type Elastic2Wrapper struct {
	client        *elastic2.Client
	bulkProcessor *elastic2.BulkProcessor
}

func NewEsClient2(config ElasticConfig, bulkWorkers int) (*Elastic2Wrapper, error) {
	var startupFns []elastic2.ClientOptionFunc

	if len(config.Url) > 0 {
		startupFns = append(startupFns, elastic2.SetURL(config.Url...))
	}

	if config.User != "" && config.Secret != "" {
		startupFns = append(startupFns, elastic2.SetBasicAuth(config.User, config.Secret))
	}

	if config.MaxRetries != nil {
		startupFns = append(startupFns, elastic2.SetMaxRetries(*config.MaxRetries))
	}

	if config.HealthCheck != nil {
		startupFns = append(startupFns, elastic2.SetHealthcheck(*config.HealthCheck))
	}

	if config.HealthCheck != nil {
		startupFns = append(startupFns, elastic2.SetHealthcheck(*config.HealthCheck))
	}

	if config.Timeout != nil {
		startupFns = append(startupFns, elastic2.SetHealthcheckTimeoutStartup(*config.Timeout))
	}

	if config.HttpClient != nil {
		startupFns = append(startupFns, elastic2.SetHttpClient(config.HttpClient))
	}

	if config.Sniff != nil {
		startupFns = append(startupFns, elastic2.SetSniff(*config.Sniff))
	}

	client, err := elastic2.NewClient(startupFns...)
	if err != nil {
		return nil, fmt.Errorf("Failed to an ElasticSearch Client: %v", err)
	}
	bps, err := client.BulkProcessor().
		Name("ElasticSearchWorker").
		Workers(bulkWorkers).
		After(bulkAfterCBV2).
		Stats(true).
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(10 * time.Second). // commit every 10s
		Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to an ElasticSearch Bulk Processor: %v", err)
	}

	return &Elastic2Wrapper{client: client, bulkProcessor: bps}, nil
}

func (es *Elastic2Wrapper) IndexExists(indices ...string) (bool, error) {
	return es.client.IndexExists(indices...).Do()
}

func (es *Elastic2Wrapper) CreateIndex(name string, mapping string) (bool, error) {
	result, err := es.client.CreateIndex(name).BodyString(mapping).Do()
	if err != nil {
		return false, err
	}
	return result.Acknowledged, err
}

func (es *Elastic2Wrapper) getAliases(index string) (*elastic2.AliasesResult, error) {
	return es.client.Aliases().Index(index).Do()
}

func (es *Elastic2Wrapper) AddAlias(index string, alias string) (bool, error) {
	res, err := es.client.Alias().Add(index, alias).Do()
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic2Wrapper) HasAlias(indexName string, aliasName string) (bool, error) {
	aliases, err := es.getAliases(indexName)
	if err != nil {
		return false, err
	}
	return aliases.Indices[indexName].HasAlias(aliasName), nil
}

func (es *Elastic2Wrapper) ErrorStats() int64 {
	if es.bulkProcessor != nil {
		return es.bulkProcessor.Stats().Failed
	}
	return 0
}

func (es *Elastic2Wrapper) AddBulkReq(index, typeName string, data interface{}) error {
	es.bulkProcessor.Add(elastic2.NewBulkIndexRequest().
		Index(index).
		Type(typeName).
		Id(uuid.NewUUID().String()).
		Doc(data))
	return nil
}

func (es *Elastic2Wrapper) FlushBulk() error {
	return es.bulkProcessor.Flush()
}

func bulkAfterCBV2(_ int64, _ []elastic2.BulkableRequest, response *elastic2.BulkResponse, err error) {
	if err != nil {
		klog.Warningf("Failed to execute bulk operation to ElasticSearch: %v", err)
	}

	if response.Errors {
		for _, list := range response.Items {
			for name, itm := range list {
				if itm.Error != nil {
					klog.V(3).Infof("Failed to execute bulk operation to ElasticSearch on %s: %v", name, itm.Error)
				}
			}
		}
	}
}
