### wechat sink

*This sink supports work wechat*.
To use the wechat sink add the following flag:

	--sink=wechat:?corp_id=<your_corp_id>&corp_secret=<your_corp_secret>&agent_id=<your_agent_id>&to_user=<to_user>&label=<your_cluster_id>&level=<Normal or Warning, Warning default>


The following options are available:
* `corp_id` - Your wechat CorpID
* `corp_secret` - Your wechat CorpSecret
* `agent_id` - Your wechat AgentID
* `to_user`  - send to user  (defualt: @all)
* `label` - Custom labels on alerting message.(such as clusterId)
* `level` - Level of event (default: Warning. Options: Warning and Normal)
* `namespaces` - Namespaces to filter (defualt: all namespaces,use commas to separate multi namespaces)
* `kinds` - Kinds to filter (default: all kinds,use commas to separate multi kinds. Options: Node,Pod and so on.)

For example:
    --sink=wechat:?corp_id=a5c19f3e02feba7bd5dfc22bfb&corp_secret=a212359acfe86fd80eb1591870&agent_id=1000012&to_user=zhangshan,xiaowang&level=Normal
or
    --sink=wechat:?corp_id=a5c19f3e02feba7bd5dfc22bfb&corp_secret=a212359acfe86fd80eb1591870&agent_id=1000012&to_user=&level=Normal
