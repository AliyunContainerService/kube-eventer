### InfluxDB sink

*This sink supports InfluxDB versions v0.9 and above*.
To use the InfluxDB sink add the following flag:

	--sink=influxdb:<INFLUXDB_URL>[?<INFLUXDB_OPTIONS>]

The following options are available:
* `user` - InfluxDB username (default: `root`)
* `pw` - InfluxDB password (default: `root`)
* `db` - InfluxDB Database name (default: `k8s`)
* `insecuressl` - Ignore SSL certificate validity (default: `false`)
* `withfields` - Use InfluxDB fields (default: `false`)
* `cluster_name` - Cluster name for different Kubernetes clusters. (default: `default`)

For example:

    --sink=influxdb:http://monitoring-influxdb:80/