### Kafka sink

To use the kafka sink add the following flag:

    --sink=kafka:<?<OPTIONS>>

Normally, kafka server has multi brokers, so brokers' list need be configured for producer.
So, we provide kafka brokers' list and topics about timeseries & topic in url's query string.
Options can be set in query string, like this:

* `brokers` - Kafka's brokers' list.
* `eventstopic` - Kafka's topic for events. Default value : `heapster-events`.
* `compression` - Kafka's compression for both topics. Must be `gzip` or `none` or `snappy` or `lz4`. Default value : none.
* `user` - Kafka's SASL PLAIN username. Must be set with `password` option.
* `password` - Kafka's SASL PLAIN password. Must be set with `user` option.
* `cacert` - Kafka's SSL Certificate Authority file path.
* `cert` - Kafka's SSL Client Certificate file path (In case of Two-way SSL). Must be set with `key` option.
* `key` - Kafka's SSL Client Private Key file path (In case of Two-way SSL). Must be set with `cert` option.
* `insecuressl` - Kafka's Ignore SSL certificate validity. Default value : `false`.

For example,

    --sink=kafka:?brokers=localhost:9092&brokers=localhost:9093&timeseriestopic=testseries
    or
    --sink=kafka:?brokers=localhost:9092&brokers=localhost:9093&eventstopic=testtopic