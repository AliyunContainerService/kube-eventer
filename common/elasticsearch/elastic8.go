package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"k8s.io/klog"

	elastic7 "github.com/olivere/elastic/v7"
	"github.com/pborman/uuid"
)

type Elastic8Wrapper struct {
	client        *elastic7.Client
	pipeline      string
	bulkProcessor *elastic7.BulkProcessor
}

func NewEsClient8(config ElasticConfig, bulkWorkers int, pipeline string) (*Elastic8Wrapper, error) {
	var startupFns []elastic7.ClientOptionFunc

	if len(config.Url) > 0 {
		startupFns = append(startupFns, elastic7.SetURL(config.Url...))
	}

	if config.User != "" && config.Secret != "" {
		startupFns = append(startupFns, elastic7.SetBasicAuth(config.User, config.Secret))
	}

	if config.MaxRetries != nil {
		startupFns = append(startupFns, elastic7.SetMaxRetries(*config.MaxRetries))
	}

	if config.HealthCheck != nil {
		startupFns = append(startupFns, elastic7.SetHealthcheck(*config.HealthCheck))
	}

	if config.HealthCheck != nil {
		startupFns = append(startupFns, elastic7.SetHealthcheck(*config.HealthCheck))
	}

	if config.Timeout != nil {
		startupFns = append(startupFns, elastic7.SetHealthcheckTimeoutStartup(*config.Timeout))
	}

	if config.HttpClient != nil {
		startupFns = append(startupFns, elastic7.SetHttpClient(config.HttpClient))
	}

	if config.Sniff != nil {
		startupFns = append(startupFns, elastic7.SetSniff(*config.Sniff))
	}

	client, err := elastic7.NewClient(startupFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to an ElasticSearch Client: %v", err)
	}
	bps, err := client.BulkProcessor().
		Name("ElasticSearchWorker").
		Workers(bulkWorkers).
		After(bulkAfterCBV8).
		Stats(true).
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(10 * time.Second). // commit every 10s
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to an ElasticSearch Bulk Processor: %v", err)
	}

	return &Elastic8Wrapper{client: client, bulkProcessor: bps, pipeline: pipeline}, nil
}

func (es *Elastic8Wrapper) IndexExists(indices ...string) (bool, error) {
	return es.client.IndexExists(indices...).Do(context.Background())
}

func (es *Elastic8Wrapper) CreateIndex(name string, mapping string) (bool, error) {
	res, err := es.client.CreateIndex(name).Do(context.Background())
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic8Wrapper) getAliases(index string) (*elastic7.AliasesResult, error) {
	return es.client.Aliases().Index(index).Do(context.Background())
}

func (es *Elastic8Wrapper) AddAlias(index string, alias string) (bool, error) {
	res, err := es.client.Alias().Add(index, alias).Do(context.Background())
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic8Wrapper) HasAlias(indexName string, aliasName string) (bool, error) {
	aliases, err := es.getAliases(indexName)
	if err != nil {
		return false, err
	}
	return aliases.Indices[indexName].HasAlias(aliasName), nil
}

func (es *Elastic8Wrapper) ErrorStats() int64 {
	if es.bulkProcessor != nil {
		return es.bulkProcessor.Stats().Failed
	}
	return 0
}

func (es *Elastic8Wrapper) AddBulkReq(index, typeName string, data interface{}) error {
	req := elastic7.NewBulkIndexRequest().
		Index(index).
		// Type(typeName).
		Id(uuid.NewUUID().String()).
		Doc(data)
	if es.pipeline != "" {
		req.Pipeline(es.pipeline)
	}

	es.bulkProcessor.Add(req)
	return nil
}

func (es *Elastic8Wrapper) FlushBulk() error {
	return es.bulkProcessor.Flush()
}

func bulkAfterCBV8(_ int64, _ []elastic7.BulkableRequest, response *elastic7.BulkResponse, err error) {
	if err != nil {
		klog.Warningf("Failed to execute bulk operation to ElasticSearch: %v", err)
	}

	if response != nil && response.Errors {
		for _, list := range response.Items {
			for name, itm := range list {
				if itm.Error != nil {
					klog.V(3).Infof("Failed to execute bulk operation to ElasticSearch on %s: %v", name, itm.Error)
				}
			}
		}
	}
}
