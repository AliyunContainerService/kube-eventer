### sls sink

*This sink supports sls (Log Service of Alibaba Cloud)*.
To use the sls sink add the following flag:

	--sink=sls:<SLS_ENDPOINTL>&logStore=[your_logstore]&project=[your_project]


The following options are available:
* `project` - Project of SLS instance. 
* `logStore` - logStore of SLS instance project. 


For example:

    --sink=sls:https://sls.aliyuncs.com?project=my_sls_project&logStore=my_sls_project_logStore