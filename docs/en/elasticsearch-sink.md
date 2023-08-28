### Elasticsearch

To use the Elasticsearch sink add the following flag:
```
    --sink=elasticsearch:<ES_SERVER_URL>[?<OPTIONS>]
```
Normally an Elasticsearch cluster has multiple nodes or a proxy, so these need
to be configured for the Elasticsearch sink. To do this, you can set
`ES_SERVER_URL` to a dummy value, and use the `?nodes=` query value for each
additional node in the cluster. For example:
```
  --sink=elasticsearch:?nodes=http://foo.com:9200&nodes=http://bar.com:9200
```
(*) Notice that using the `?nodes` notation will override the `ES_SERVER_URL`

If you run your ElasticSearch cluster behind a loadbalancer (or otherwise do
not want to specify multiple nodes) then you can do the following:

(*) Be sure to add your version tag in your sink;
```
  --sink=elasticsearch:http://elasticsearch.example.com:9200?sniff=false&ver=6
```
Besides this, the following options can be set in query string:

(*) Note that the keys are case sensitive

* `index` - the index for metrics and events. The default is `heapster`, you can define index with uri param: index=xxx
* `esUserName` - the username if authentication is enabled
* `esUserSecret` - the password if authentication is enabled
* `maxRetries` - the number of retries that the Elastic client will perform
  for a single request after before giving up and return an error. It is `0`
  by default, so retry is disabled by default.
* `healthCheck` - specifies if healthCheck are enabled by default. It is enabled
  by default. To disable, provide a negative boolean value like `0` or `false`.
* `sniff` - specifies if the sniffer is enabled by default. It is enabled
  by default. To disable, provide a negative boolean value like `0` or `false`.
* `startupHealthcheckTimeout` - the time in seconds the healthCheck waits for
  a response from Elasticsearch on startup, i.e. when creating a client. The
  default value is `1`.
* `ver` - ElasticSearch cluster version, can be either `2`, `5`, `6` or `7`. The default is `5`
* `bulkWorkers` - number of workers for bulk processing. Default value is `5`.
* `cluster_name` - cluster name for different Kubernetes clusters. Default value is `default`.
* `pipeline` - (optional; >ES5) Ingest Pipeline to process the documents. The default is disabled(empty value)