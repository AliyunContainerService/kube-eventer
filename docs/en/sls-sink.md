### sls sink

*This sink supports sls (Log Service of Alibaba Cloud)*.
To use the sls sink add the following flag:

	--sink=sls:<SLS_ENDPOINT>?logStore=[your_logstore]&project=[your_project]&topic=[topic_for_log]&regionId=[your_region_id]&label=<key,value>


The following options are available:
* `project` - Project of SLS instance. 
* `logStore` - logStore of SLS instance project. 
* `topic` - topic for every log sent to SLS. 
* `regionId` - **Strongly recommended for self-built (non-Alibaba Cloud managed) clusters.** The Alibaba Cloud region ID where your SLS instance resides (e.g. `cn-hangzhou`). If not set via this parameter or the `RegionId` environment variable, kube-eventer will attempt to auto-detect the region from the instance metadata service (`http://100.100.100.200/latest/meta-data/region-id`), which is unavailable in self-built clusters and will cause a startup error. Can also be configured via the `RegionId` environment variable.
* `label` - Custom labels on alerting message.(such as clusterId), format is label=ClusterId,test_clusterId&label=RegionId,test_regionId&label=UserId,test_uid
* `accessKeyId` - optional param. aliyun access key to sink to sls. 
* `accessKeySecret` - optional param. aliyun access key secret to sink to sls.
* `internal` - optional param. if true, it will sink to sls through aliyun internal network connection. 

> **Note for self-built clusters:** If `regionId` is not specified and the metadata service is unreachable, kube-eventer will fail at startup with an error like:
> `can not create sls client because of Get "http://100.100.100.200/latest/meta-data/region-id": dial tcp 100.100.100.200:80: i/o timeout`
>
> Always set `regionId` explicitly when running outside of Alibaba Cloud managed environments.

For example:

    --sink=sls:https://sls.aliyuncs.com?project=my_sls_project&logStore=my_sls_project_logStore&topic=k8s-cluster-dev&regionId=cn-hangzhou&label=Key1,Value1&label=Key2,Value2

For VPC/internal network access:

    --sink=sls:https://sls.aliyuncs.com?project=my_sls_project&logStore=my_sls_project_logStore&topic=k8s-cluster-dev&regionId=cn-hangzhou&internal=true&label=Key1,Value1

#### How to configure regionId via environment variable.

You can also set `regionId` through the `RegionId` environment variable in the kube-eventer deployment template:

    env:
     - name: RegionId
       value: "cn-hangzhou"

> Note: The environment variable name is `RegionId` (case-sensitive).

#### How to config aliyun access key.

*If you run kube-eventer manually, you will need to config aliyun access key to get the permission to sink your data to sls. when running on Aliyun Container Service, you don't need to config access key manually.*

You can config kube-eventer global aliyun access key through kube-eventer deployment's template env params:

    
    env:
     - name: AccessKeyId
       value: "xxx"
     - name: AccessKeySecret
       value: "xxx"
     

 You can also config aliyun access key with sls sink optional params.