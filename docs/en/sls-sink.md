### sls sink

*This sink supports sls (Log Service of Alibaba Cloud)*.
To use the sls sink add the following flag:

	--sink=sls:<SLS_ENDPOINTL>&logStore=[your_logstore]&project=[your_project]&topic=[topic_for_log]


The following options are available:
* `project` - Project of SLS instance. 
* `logStore` - logStore of SLS instance project. 
* `topic` - topic for every log sent to SLS. 


For example:

    --sink=sls:https://sls.aliyuncs.com?project=my_sls_project&logStore=my_sls_project_logStore&topic=k8s-cluster-dev