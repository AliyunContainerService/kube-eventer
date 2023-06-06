### sls sink

*This sink supports sls (Log Service of Alibaba Cloud)*.
To use the sls sink add the following flag:

	--sink=sls:<SLS_ENDPOINTL>?logStore=[your_logstore]&project=[your_project]&topic=[topic_for_log]&label=<key,value>


The following options are available:
* `project` - Project of SLS instance. 
* `logStore` - logStore of SLS instance project. 
* `topic` - topic for every log sent to SLS. 
* `label` - Custom labels on alerting message.(such as clusterId), format is label=ClusterId,test_clusterId&label=RegionId,test_regionId&label=UserId,test_uid
* `accessKeyId` - optional param. aliyun access key to sink to sls. 
* `accessKeySecret` - optional param. aliyun access key secret to sink to sls.
* `internal` - optional param. if true, it will sink to sls through aliyun internal network connection. 

For example:

    --sink=sls:https://sls.aliyuncs.com?project=my_sls_project&logStore=my_sls_project_logStore&topic=k8s-cluster-dev&label=Key1,Value1&label=Key2,Value2
    
#### How to config aliyun access key.

*If you run kube-eventer manually, you will need to config aliyun access key to get the permission to sink your data to sls. when running on Aliyun Container Service, you don't need to config access key manually.*

You can config kube-eventer global aliyun access key through kube-eventer deployment's template env params:

    
    env:
     - AccessKeyId: "xxx"
     - AccessKeySecret: "xxx"
     
 
 You can also config aliyun access key with sls sink optional params.