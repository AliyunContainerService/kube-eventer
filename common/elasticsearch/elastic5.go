package elasticsearch

import (
	"context"
	"fmt"
	"k8s.io/klog"
	"time"

	"github.com/pborman/uuid"
	elastic5 "gopkg.in/olivere/elastic.v5"
)

type Elastic5Wrapper struct {
	client        *elastic5.Client
	pipeline      string
	bulkProcessor *elastic5.BulkProcessor
}

func NewEsClient5(config ElasticConfig, bulkWorkers int, pipeline string) (*Elastic5Wrapper, error) {
	var startupFns []elastic5.ClientOptionFunc

	if len(config.Url) > 0 {
		startupFns = append(startupFns, elastic5.SetURL(config.Url...))
	}

	if config.User != "" && config.Secret != "" {
		startupFns = append(startupFns, elastic5.SetBasicAuth(config.User, config.Secret))
	}

	if config.MaxRetries != nil {
		startupFns = append(startupFns, elastic5.SetMaxRetries(*config.MaxRetries))
	}

	if config.HealthCheck != nil {
		startupFns = append(startupFns, elastic5.SetHealthcheck(*config.HealthCheck))
	}

	if config.HealthCheck != nil {
		startupFns = append(startupFns, elastic5.SetHealthcheck(*config.HealthCheck))
	}

	if config.Timeout != nil {
		startupFns = append(startupFns, elastic5.SetHealthcheckTimeoutStartup(*config.Timeout))
	}

	if config.HttpClient != nil {
		startupFns = append(startupFns, elastic5.SetHttpClient(config.HttpClient))
	}

	if config.Sniff != nil {
		startupFns = append(startupFns, elastic5.SetSniff(*config.Sniff))
	}

	client, err := elastic5.NewClient(startupFns...)
	if err != nil {
		return nil, fmt.Errorf("Failed to an ElasticSearch Client: %v", err)
	}
	bps, err := client.BulkProcessor().
		Name("ElasticSearchWorker").
		Workers(bulkWorkers).
		After(bulkAfterCBV5).
		Stats(true).
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(10 * time.Second). // commit every 10s
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to an ElasticSearch Bulk Processor: %v", err)
	}

	return &Elastic5Wrapper{client: client, bulkProcessor: bps, pipeline: pipeline}, nil
}

func (es *Elastic5Wrapper) IndexExists(indices ...string) (bool, error) {
	return es.client.IndexExists(indices...).Do(context.Background())
}

func (es *Elastic5Wrapper) CreateIndex(name string, mapping string) (bool, error) {
	res, err := es.client.CreateIndex(name).BodyString(mapping).Do(context.Background())
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic5Wrapper) getAliases(index string) (*elastic5.AliasesResult, error) {
	return es.client.Aliases().Index(index).Do(context.Background())
}

func (es *Elastic5Wrapper) AddAlias(index string, alias string) (bool, error) {
	res, err := es.client.Alias().Add(index, alias).Do(context.Background())
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic5Wrapper) HasAlias(indexName string, aliasName string) (bool, error) {
	aliases, err := es.getAliases(indexName)
	if err != nil {
		return false, err
	}
	return aliases.Indices[indexName].HasAlias(aliasName), nil
}

func (es *Elastic5Wrapper) ErrorStats() int64 {
	if es.bulkProcessor != nil {
		return es.bulkProcessor.Stats().Failed
	}
	return 0
}

func (es *Elastic5Wrapper) AddBulkReq(index, typeName string, data interface{}) error {
	req := elastic5.NewBulkIndexRequest().
		Index(index).
		Type(typeName).
		Id(uuid.NewUUID().String()).
		Doc(data)
	if es.pipeline != "" {
		req.Pipeline(es.pipeline)
	}

	es.bulkProcessor.Add(req)
	return nil
}

func (es *Elastic5Wrapper) FlushBulk() error {
	return es.bulkProcessor.Flush()
}

func bulkAfterCBV5(_ int64, _ []elastic5.BulkableRequest, response *elastic5.BulkResponse, err error) {
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
