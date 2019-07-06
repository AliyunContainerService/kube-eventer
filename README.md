## kube-eventer    

<p align="center">
	<img src="docs/logo/kube-eventer.png" width="150px" />   
  <p align="center">
    kube-eventer emit kubernetes events to sinks
  </p>
</p>

### Overview 

kube-eventer is an event emitter that sends kubernetes events to sinks(.e.g, dingtalk,sls,kafka and so on). The core design concept of kubernetes is state machine. So there will be `Normal` events when transfer to desired state and `Warning` events occur when to unexpected state. kube-eventer can help to diagnose, analysis and alarm problems.


### Sink Configure 

Supported Sink:

| Sink Name                    | Description                       |
| ---------------------------- | :-------------------------------- |
| <a href="docs/en/dingtalk-sink.md">dingtalk</a>      | sink to dingtalk bot              |
| <a href="docs/en/sls-sink.md">sls</a>           | sink to alibaba cloud sls service |
| <a href="docs/en/elasticsearch-sink.md">elasticsearch</a> | sink to elasticsearch             |
| <a href="docs/en/honeycomb-sink.md">honeycomb</a>     | sink to honeycomb                 |
| <a href="docs/en/influxdb-sink.md">influxdb</a>      | sink to influxdb                  |
| <a href="docs/en/kafka-sink.md">kafka</a>         | sink to kafka                     |
| <a href="docs/en/log-sink.md">log</a>               | sink to standard output           |
|                              |                                   |

### Contributing 

Please check <a href="docs/en/CONTRIBUTING.md" target="_blank">CONTRIBUTING.md</a>


### License 
This software is released under the Apache 2.0 license.