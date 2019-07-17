### dingtalk sink

*This sink supports dingtalk bot*.
To use the dingtalk sink add the following flag:

	--sink=dingtalk:<DINGTALK_WEBHOOK_URL>&label=<your_cluster_id>&level=<Normal or Warning, Warning default>


The following options are available:
* `label` - Custom labels on alerting message.(such as clusterId)
* `level` - Level of event (default: Warning. Options: Warning and Normal)
* `namespaces` - Namespaces to filter (defualt: all namespaces,use commas to separate multi namespaces)
* `kinds` - Kinds to filter (default: all kinds,use commas to separate multi kinds. Options: Node,Pod and so on.)

For example:

    --sink=dingtalk:https://oapi.dingtalk.com/robot/send?access_token=a5c19f3e02feba7bd5dfc22bfb04afa212359acfe86fd80eb159187097b7d014&label=c550367cdf1e84dfabab013b277cc6bc2&level=Normal


#### Markdown dingtalk alert

**WARNING:ONLY SUPPORT ALIYUN PLATFORM**

Default alert mode is text.
You can also use `Markdown` alert mode by setting following flag:

    --sink=dingtalk:<DINGTALK_WEBHOOK_URL>&label=<your_cluster_id>&level=<Normal or Warning, Warning default>&msg_type=markdown&cluster_id=<cluster_id>&region=<region>

msg_type , cluster_id , cluster_id ,those params are all required.

You can find `cluster_id` on [ALIYUN Kubernetes website](https://cs.console.aliyun.com/#/k8s/cluster/list)

For example:

    --sink=dingtalk:https://oapi.dingtalk.com/robot/send?access_token=a5c19f3e02feba7bd5dfc22bfb04afa212359acfe86fd80eb159187097b7d014&label=c550367cdf1e84dfabab013b277cc6bc2&level=Normal&msg_type=markdown&cluster_id=a5c19f3e02feba7bd5dfc22bfb04afa212359acfe86fd80eb159187097b7d014&region=cn-shenzhen