package sls

import (
	"net/http"
	"time"

	"github.com/aliyun/aliyun-log-go-sdk/util"
)

// CreateNormalInterface create a normal client.
//
// Deprecated: use CreateNormalInterfaceV2 instead.
// If you keep using long-lived AccessKeyID and AccessKeySecret,
// use the example code below.
//
//	  provider := NewStaticCredentailsProvider(accessKeyID, accessKeySecret, securityToken)
//		client := CreateNormalInterfaceV2(endpoint, provider)
func CreateNormalInterface(endpoint, accessKeyID, accessKeySecret, securityToken string) ClientInterface {
	client := &Client{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		SecurityToken:   securityToken,

		credentialsProvider: NewStaticCredentialsProvider(
			accessKeyID,
			accessKeySecret,
			securityToken,
		),
	}
	client.setSignV4IfInAcdr(endpoint)
	return client
}

// CreateNormalInterfaceV2 create a normal client, with a CredentialsProvider.
//
// It is highly recommended to use a CredentialsProvider that provides dynamic
// expirable credentials for security.
//
// See [credentials_provider.go] for more details.
func CreateNormalInterfaceV2(endpoint string, credentialsProvider CredentialsProvider) ClientInterface {
	client := &Client{
		Endpoint:            endpoint,
		credentialsProvider: credentialsProvider,
	}
	client.setSignV4IfInAcdr(endpoint)
	return client
}

type UpdateTokenFunction = util.UpdateTokenFunction

// CreateTokenAutoUpdateClient create a TokenAutoUpdateClient,
// this client will auto fetch security token and retry when operation is `Unauthorized`
//
// Deprecated: Use CreateNormalInterfaceV2 and UpdateFuncProviderAdapter instead.
//
// Example:
//
//		provider := NewUpdateFuncProviderAdapter(updateStsTokenFunc)
//	  client := CreateNormalInterfaceV2(endpoint, provider)
//
// @note TokenAutoUpdateClient will destroy when shutdown channel is closed
func CreateTokenAutoUpdateClient(endpoint string, tokenUpdateFunc UpdateTokenFunction, shutdown <-chan struct{}) (client ClientInterface, err error) {
	accessKeyID, accessKeySecret, securityToken, expireTime, err := tokenUpdateFunc()
	if err != nil {
		return nil, err
	}
	tauc := &TokenAutoUpdateClient{
		logClient:              CreateNormalInterface(endpoint, accessKeyID, accessKeySecret, securityToken),
		shutdown:               shutdown,
		tokenUpdateFunc:        tokenUpdateFunc,
		maxTryTimes:            3,
		waitIntervalMin:        time.Duration(time.Second * 1),
		waitIntervalMax:        time.Duration(time.Second * 60),
		updateTokenIntervalMin: time.Duration(time.Second * 1),
		nextExpire:             expireTime,
	}
	go tauc.flushSTSToken()
	return tauc, nil
}

// ClientInterface for all log's open api
type ClientInterface interface {
	// SetUserAgent set userAgent for sls client
	SetUserAgent(userAgent string)
	// SetHTTPClient set a custom http client, all request will send to sls by this client
	SetHTTPClient(client *http.Client)
	// SetRetryTimeout set retry timeout, client will retry util retry timeout
	SetRetryTimeout(timeout time.Duration)
	// #################### Client Operations #####################
	// ResetAccessKeyToken reset client's access key token
	ResetAccessKeyToken(accessKeyID, accessKeySecret, securityToken string)
	// SetRegion Set region for signature v4
	SetRegion(region string)
	// SetAuthVersion Set signature version
	SetAuthVersion(version AuthVersionType)
	// Close the client
	Close() error

	// #################### Project Operations #####################
	// CreateProject create a new loghub project.
	CreateProject(name, description string) (*LogProject, error)
	// CreateProject create a new loghub project, with dataRedundancyType option.
	CreateProjectV2(name, description, dataRedundancyType string) (*LogProject, error)
	GetProject(name string) (*LogProject, error)
	// UpdateProject create a new loghub project.
	UpdateProject(name, description string) (*LogProject, error)
	// ListProject list all projects in specific region
	// the region is related with the client's endpoint
	ListProject() (projectNames []string, err error)
	// ListProjectV2 list all projects in specific region
	// the region is related with the client's endpoint
	// ref https://www.alibabacloud.com/help/doc-detail/74955.htm
	ListProjectV2(offset, size int) (projects []LogProject, count, total int, err error)
	// CheckProjectExist check project exist or not
	CheckProjectExist(name string) (bool, error)
	// DeleteProject ...
	DeleteProject(name string) error

	// #################### Logstore Operations #####################
	// ListLogStore returns all logstore names of project p.
	ListLogStore(project string) ([]string, error)
	// GetLogStore returns logstore according by logstore name.
	GetLogStore(project string, logstore string) (*LogStore, error)
	// CreateLogStore creates a new logstore in SLS
	// where name is logstore name,
	// and ttl is time-to-live(in day) of logs,
	// and shardCnt is the number of shards,
	// and autoSplit is auto split,
	// and maxSplitShard is the max number of shard.
	CreateLogStore(project string, logstore string, ttl, shardCnt int, autoSplit bool, maxSplitShard int) error
	// CreateLogStoreV2 creates a new logstore in SLS
	CreateLogStoreV2(project string, logstore *LogStore) error
	// DeleteLogStore deletes a logstore according by logstore name.
	DeleteLogStore(project string, logstore string) (err error)
	// UpdateLogStore updates a logstore according by logstore name,
	// obviously we can't modify the logstore name itself.
	UpdateLogStore(project string, logstore string, ttl, shardCnt int) (err error)
	// UpdateLogStoreV2 updates a logstore according by logstore name,
	// obviously we can't modify the logstore name itself.
	UpdateLogStoreV2(project string, logstore *LogStore) error
	// CheckLogstoreExist check logstore exist or not
	CheckLogstoreExist(project string, logstore string) (bool, error)
	// GetLogStoreMeteringMode get the metering mode of logstore, eg. ChargeByFunction / ChargeByDataIngest
	GetLogStoreMeteringMode(project string, logstore string) (*GetMeteringModeResponse, error)
	// GetLogStoreMeteringMode update the metering mode of logstore, eg. ChargeByFunction / ChargeByDataIngest
	//
	// Warning: this method may affect your billings, for more details ref: https://www.aliyun.com/price/detail/sls
	UpdateLogStoreMeteringMode(project string, logstore string, meteringMode string) error

	// #################### MetricStore Operations #####################
	// CreateMetricStore creates a new metric store in SLS.
	CreateMetricStore(project string, metricStore *LogStore) error
	// UpdateMetricStore updates a metric store.
	UpdateMetricStore(project string, metricStore *LogStore) error
	// DeleteMetricStore deletes a metric store.
	DeleteMetricStore(project, name string) error
	// GetMetricStore return a metric store.
	GetMetricStore(project, name string) (*LogStore, error)

	// #################### EventStore Operations #####################
	// CreateEventStore creates a new event store in SLS.
	CreateEventStore(project string, eventStore *LogStore) error
	// UpdateEventStore updates a event store.
	UpdateEventStore(project string, eventStore *LogStore) error
	// DeleteEventStore deletes a event store.
	DeleteEventStore(project, name string) error
	// GetEventStore return a event store.
	GetEventStore(project, name string) (*LogStore, error)
	// ListEventStore returns all eventStore names of project p.
	ListEventStore(project string, offset, size int) ([]string, error)

	// #################### StoreView Operations #####################
	// CreateStoreView creates a new storeView.
	CreateStoreView(project string, storeView *StoreView) error
	// UpdateStoreView updates a storeView.
	UpdateStoreView(project string, storeView *StoreView) error
	// DeleteStoreView deletes a storeView.
	DeleteStoreView(project string, storeViewName string) error
	// GetStoreView returns storeView.
	GetStoreView(project string, storeViewName string) (*StoreView, error)
	// ListStoreViews returns all storeView names of a project.
	ListStoreViews(project string, req *ListStoreViewsRequest) (*ListStoreViewsResponse, error)
	// GetStoreViewIndex returns all index config of logstores in the storeView, only support storeType logstore.
	GetStoreViewIndex(project string, storeViewName string) (*GetStoreViewIndexResponse, error)

	// #################### Logtail Operations #####################
	// ListMachineGroup returns machine group name list and the total number of machine groups.
	// The offset starts from 0 and the size is the max number of machine groups could be returned.
	ListMachineGroup(project string, offset, size int) (m []string, total int, err error)
	// ListMachines list all machines in machineGroupName
	ListMachines(project, machineGroupName string) (ms []*Machine, total int, err error)
	ListMachinesV2(project, machineGroupName string, offset, size int) (ms []*Machine, total int, err error)
	// CheckMachineGroupExist check machine group exist or not
	CheckMachineGroupExist(project string, machineGroup string) (bool, error)
	// GetMachineGroup retruns machine group according by machine group name.
	GetMachineGroup(project string, machineGroup string) (m *MachineGroup, err error)
	// CreateMachineGroup creates a new machine group in SLS.
	CreateMachineGroup(project string, m *MachineGroup) error
	// UpdateMachineGroup updates a machine group.
	UpdateMachineGroup(project string, m *MachineGroup) (err error)
	// DeleteMachineGroup deletes machine group according machine group name.
	DeleteMachineGroup(project string, machineGroup string) (err error)
	// ListConfig returns config names list and the total number of configs.
	// The offset starts from 0 and the size is the max number of configs could be returned.
	ListConfig(project string, offset, size int) (cfgNames []string, total int, err error)
	// CheckConfigExist check config exist or not
	CheckConfigExist(project string, config string) (ok bool, err error)
	// GetConfig returns config according by config name.
	GetConfig(project string, config string) (logConfig *LogConfig, err error)
	// GetConfigString returns config according by config name.
	GetConfigString(name string, config string) (c string, err error)
	// UpdateConfig updates a config.
	UpdateConfig(project string, config *LogConfig) (err error)
	// UpdateConfigString updates a config.
	UpdateConfigString(project string, configName, configDetail string) (err error)
	// CreateConfig creates a new config in SLS.
	CreateConfig(project string, config *LogConfig) (err error)
	// CreateConfigString creates a new config in SLS.
	CreateConfigString(project string, config string) (err error)
	// DeleteConfig deletes a config according by config name.
	DeleteConfig(project string, config string) (err error)
	// GetAppliedMachineGroups returns applied machine group names list according config name.
	GetAppliedMachineGroups(project string, confName string) (groupNames []string, err error)
	// GetAppliedConfigs returns applied config names list according machine group name groupName.
	GetAppliedConfigs(project string, groupName string) (confNames []string, err error)
	// ApplyConfigToMachineGroup applies config to machine group.
	ApplyConfigToMachineGroup(project string, confName, groupName string) (err error)
	// RemoveConfigFromMachineGroup removes config from machine group.
	RemoveConfigFromMachineGroup(project string, confName, groupName string) (err error)

	// #################### ETL Operations #####################
	CreateETL(project string, etljob ETL) error
	UpdateETL(project string, etljob ETL) error
	GetETL(project string, etlName string) (ETLJob *ETL, err error)
	ListETL(project string, offset int, size int) (*ListETLResponse, error)
	DeleteETL(project string, etlName string) error
	StartETL(project, name string) error
	StopETL(project, name string) error
	RestartETL(project string, etljob ETL) error

	CreateEtlMeta(project string, etlMeta *EtlMeta) (err error)
	UpdateEtlMeta(project string, etlMeta *EtlMeta) (err error)
	DeleteEtlMeta(project string, etlMetaName, etlMetaKey string) (err error)
	listEtlMeta(project string, etlMetaName, etlMetaKey, etlMetaTag string, offset, size int) (total int, count int, etlMeta []*EtlMeta, err error)
	GetEtlMeta(project string, etlMetaName, etlMetaKey string) (etlMeta *EtlMeta, err error)
	ListEtlMeta(project string, etlMetaName string, offset, size int) (total int, count int, etlMetaList []*EtlMeta, err error)
	ListEtlMetaWithTag(project string, etlMetaName, etlMetaTag string, offset, size int) (total int, count int, etlMetaList []*EtlMeta, err error)
	ListEtlMetaName(project string, offset, size int) (total int, count int, etlMetaNameList []string, err error)

	// #################### Shard Operations #####################
	// ListShards returns shard id list of this logstore.
	ListShards(project, logstore string) (shards []*Shard, err error)
	// SplitShard https://help.aliyun.com/document_detail/29021.html,
	SplitShard(project, logstore string, shardID int, splitKey string) (shards []*Shard, err error)
	// SplitNumShard https://help.aliyun.com/document_detail/29021.html,
	SplitNumShard(project, logstore string, shardID, shardsNum int) (shards []*Shard, err error)
	// MergeShards https://help.aliyun.com/document_detail/29022.html
	MergeShards(project, logstore string, shardID int) (shards []*Shard, err error)

	// #################### Log Operations #####################
	PutLogsWithMetricStoreURL(project, logstore string, lg *LogGroup) (err error)
	// PutLogs put logs into logstore.
	// The callers should transform user logs into LogGroup.
	PutLogs(project, logstore string, lg *LogGroup) (err error)
	// PostLogStoreLogs put logs into Shard logstore by hashKey.
	// The callers should transform user logs into LogGroup.
	PostLogStoreLogs(project, logstore string, lg *LogGroup, hashKey *string) (err error)
	PostLogStoreLogsV2(project, logstore string, req *PostLogStoreLogsRequest) (err error)
	// PostRawLogWithCompressType put logs into logstore with specific compress type and hashKey.
	PostRawLogWithCompressType(project, logstore string, rawLogData []byte, compressType int, hashKey *string) (err error)
	// PutLogsWithCompressType put logs into logstore with specific compress type.
	// The callers should transform user logs into LogGroup.
	PutLogsWithCompressType(project, logstore string, lg *LogGroup, compressType int) (err error)
	// PutRawLogWithCompressType put raw log data to log service, no marshal
	PutRawLogWithCompressType(project, logstore string, rawLogData []byte, compressType int) (err error)
	// GetCursor gets log cursor of one shard specified by shardId.
	// The from can be in three form: a) unix timestamp in seccond, b) "begin", c) "end".
	// For more detail please read: https://help.aliyun.com/document_detail/29024.html
	GetCursor(project, logstore string, shardID int, from string) (cursor string, err error)
	// GetCursorTime gets the server time based on the cursor.
	// For more detail please read: https://help.aliyun.com/document_detail/113274.html
	GetCursorTime(project, logstore string, shardID int, cursor string) (cursorTime time.Time, err error)
	// GetLogsBytes gets logs binary data from shard specified by shardId according cursor and endCursor.
	// The logGroupMaxCount is the max number of logGroup could be returned.
	// The nextCursor is the next curosr can be used to read logs at next time.
	GetLogsBytes(project, logstore string, shardID int, cursor, endCursor string,
		logGroupMaxCount int) (out []byte, nextCursor string, err error)
	// Deprecated: Use GetLogsBytesWithQuery instead.
	GetLogsBytesV2(plr *PullLogRequest) (out []byte, nextCursor string, err error)
	GetLogsBytesWithQuery(plr *PullLogRequest) (out []byte, plm *PullLogMeta, err error)
	// PullLogs gets logs from shard specified by shardId according cursor and endCursor.
	// The logGroupMaxCount is the max number of logGroup could be returned.
	// The nextCursor is the next cursor can be used to read logs at next time.
	// @note if you want to pull logs continuous, set endCursor = ""
	PullLogs(project, logstore string, shardID int, cursor, endCursor string,
		logGroupMaxCount int) (gl *LogGroupList, nextCursor string, err error)
	// Deprecated: Use PullLogsWithQuery instead.
	PullLogsV2(plr *PullLogRequest) (gl *LogGroupList, nextCursor string, err error)
	PullLogsWithQuery(plr *PullLogRequest) (gl *LogGroupList, plm *PullLogMeta, err error)
	// GetHistograms query logs with [from, to) time range
	GetHistograms(project, logstore string, topic string, from int64, to int64, queryExp string) (*GetHistogramsResponse, error)
	// GetLogs query logs with [from, to) time range
	GetLogs(project, logstore string, topic string, from int64, to int64, queryExp string,
		maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error)
	GetLogLines(project, logstore string, topic string, from int64, to int64, queryExp string,
		maxLineNum int64, offset int64, reverse bool) (*GetLogLinesResponse, error)
	// GetLogsByNano query logs with [fromInNs, toInNs) nano time range
	GetLogsByNano(project, logstore string, topic string, fromInNs int64, toInNs int64, queryExp string,
		maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error)
	GetLogLinesByNano(project, logstore string, topic string, fromInNs int64, toInNs int64, queryExp string,
		maxLineNum int64, offset int64, reverse bool) (*GetLogLinesResponse, error)

	GetLogsV2(project, logstore string, req *GetLogRequest) (*GetLogsResponse, error)
	GetLogLinesV2(project, logstore string, req *GetLogRequest) (*GetLogLinesResponse, error)
	GetLogsV3(project, logstore string, req *GetLogRequest) (*GetLogsV3Response, error)

	// GetHistogramsToCompleted query logs with [from, to) time range to completed
	GetHistogramsToCompleted(project, logstore string, topic string, from int64, to int64, queryExp string) (*GetHistogramsResponse, error)
	// GetLogsToCompleted query logs with [from, to) time range to completed
	GetLogsToCompleted(project, logstore string, topic string, from int64, to int64, queryExp string, maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error)
	// GetLogsToCompletedV2 query logs with [from, to) time range to completed
	GetLogsToCompletedV2(project, logstore string, req *GetLogRequest) (*GetLogsResponse, error)
	// GetLogsToCompletedV3 query logs with [from, to) time range to completed
	GetLogsToCompletedV3(project, logstore string, req *GetLogRequest) (*GetLogsV3Response, error)

	// #################### Index Operations #####################
	// CreateIndex ...
	CreateIndex(project, logstore string, index Index) error
	// CreateIndexString ...
	CreateIndexString(project, logstore string, indexStr string) error
	// UpdateIndex ...
	UpdateIndex(project, logstore string, index Index) error
	// UpdateIndexString ...
	UpdateIndexString(project, logstore string, indexStr string) error
	// DeleteIndex ...
	DeleteIndex(project, logstore string) error
	// GetIndex ...
	GetIndex(project, logstore string) (*Index, error)
	// GetIndexString ...
	GetIndexString(project, logstore string) (string, error)

	// #################### Chart&Dashboard Operations #####################
	ListDashboard(project string, dashboardName string, offset, size int) (dashboardList []string, count, total int, err error)
	ListDashboardV2(project string, dashboardName string, offset, size int) (dashboardList []string, dashboardItems []ResponseDashboardItem, count, total int, err error)
	GetDashboard(project, name string) (dashboard *Dashboard, err error)
	GetDashboardString(project, name string) (dashboard string, err error)
	DeleteDashboard(project, name string) error
	UpdateDashboard(project string, dashboard Dashboard) error
	UpdateDashboardString(project string, dashboardName, dashboardStr string) error
	CreateDashboard(project string, dashboard Dashboard) error
	CreateDashboardString(project string, dashboardStr string) error
	GetChart(project, dashboardName, chartName string) (chart *Chart, err error)
	DeleteChart(project, dashboardName, chartName string) error
	UpdateChart(project, dashboardName string, chart Chart) error
	CreateChart(project, dashboardName string, chart Chart) error

	// #################### SavedSearch&Alert Operations #####################
	CreateSavedSearch(project string, savedSearch *SavedSearch) error
	UpdateSavedSearch(project string, savedSearch *SavedSearch) error
	DeleteSavedSearch(project string, savedSearchName string) error
	GetSavedSearch(project string, savedSearchName string) (*SavedSearch, error)
	ListSavedSearch(project string, savedSearchName string, offset, size int) (savedSearches []string, total int, count int, err error)
	ListSavedSearchV2(project string, savedSearchName string, offset, size int) (savedSearches []string, savedsearchItems []ResponseSavedSearchItem, total int, count int, err error)
	CreateAlert(project string, alert *Alert) error
	UpdateAlert(project string, alert *Alert) error
	DeleteAlert(project string, alertName string) error
	GetAlert(project string, alertName string) (*Alert, error)
	DisableAlert(project string, alertName string) error
	EnableAlert(project string, alertName string) error
	ListAlert(project, alertName, dashboard string, offset, size int) (alerts []*Alert, total int, count int, err error)
	CreateAlertString(project string, alert string) error
	UpdateAlertString(project string, alertName, alert string) error
	GetAlertString(project string, alertName string) (string, error)

	// #################### Consumer Operations #####################
	CreateConsumerGroup(project, logstore string, cg ConsumerGroup) (err error)
	UpdateConsumerGroup(project, logstore string, cg ConsumerGroup) (err error)
	DeleteConsumerGroup(project, logstore string, cgName string) (err error)
	ListConsumerGroup(project, logstore string) (cgList []*ConsumerGroup, err error)
	HeartBeat(project, logstore string, cgName, consumer string, heartBeatShardIDs []int) (shardIDs []int, err error)
	UpdateCheckpoint(project, logstore string, cgName string, consumer string, shardID int, checkpoint string, forceSuccess bool) (err error)
	GetCheckpoint(project, logstore string, cgName string) (checkPointList []*ConsumerGroupCheckPoint, err error)

	// ####################### Resource Tags API ######################
	// TagResources tag specific resource
	TagResources(project string, tags *ResourceTags) error
	// UnTagResources untag specific resource
	UnTagResources(project string, tags *ResourceUnTags) error
	// ListTagResources list rag resources
	ListTagResources(project string,
		resourceType string,
		resourceIDs []string,
		tags []ResourceFilterTag,
		nextToken string) (respTags []*ResourceTagResponse, respNextToken string, err error)
	// TagResourcesSystemTags tag specific resource
	TagResourcesSystemTags(project string, tags *ResourceSystemTags) error
	// UnTagResourcesSystemTags untag specific resource
	UnTagResourcesSystemTags(project string, tags *ResourceUnSystemTags) error
	// ListSystemTagResources list system tag resources
	ListSystemTagResources(project string,
		resourceType string,
		resourceIDs []string,
		tags []ResourceFilterTag,
		tagOwnerUid string,
		category string,
		scope string,
		nextToken string) (respTags []*ResourceTagResponse, respNextToken string, err error)
	CreateScheduledSQL(project string, scheduledsql *ScheduledSQL) error
	DeleteScheduledSQL(project string, name string) error
	UpdateScheduledSQL(project string, scheduledsql *ScheduledSQL) error
	GetScheduledSQL(project string, name string) (*ScheduledSQL, error)
	ListScheduledSQL(project, name, displayName string, offset, size int) ([]*ScheduledSQL, int, int, error)
	GetScheduledSQLJobInstance(projectName, jobName, instanceId string, result bool) (instance *ScheduledSQLJobInstance, err error)
	ModifyScheduledSQLJobInstanceState(projectName, jobName, instanceId string, state ScheduledSQLState) error
	ListScheduledSQLJobInstances(projectName, jobName string, status *InstanceStatus) (instances []*ScheduledSQLJobInstance, total, count int64, err error)

	// #################### Resource Operations #####################
	ListResource(resourceType string, resourceName string, offset, size int) (resourceList []*Resource, count, total int, err error)
	GetResource(name string) (resource *Resource, err error)
	GetResourceString(name string) (resource string, err error)
	DeleteResource(name string) error
	UpdateResource(resource *Resource) error
	UpdateResourceString(resourceName, resourceStr string) error
	CreateResource(resource *Resource) error
	CreateResourceString(resourceStr string) error

	// #################### Resource Record Operations #####################
	ListResourceRecord(resourceName string, offset, size int) (recordList []*ResourceRecord, count, total int, err error)
	GetResourceRecord(resourceName, recordId string) (record *ResourceRecord, err error)
	GetResourceRecordString(resourceName, name string) (record string, err error)
	DeleteResourceRecord(resourceName, recordId string) error
	UpdateResourceRecord(resourceName string, record *ResourceRecord) error
	UpdateResourceRecordString(resourceName, recordStr string) error
	CreateResourceRecord(resourceName string, record *ResourceRecord) error
	CreateResourceRecordString(resourceName, recordStr string) error

	// #################### Ingestion #####################
	CreateIngestion(project string, ingestion *Ingestion) error
	UpdateIngestion(project string, ingestion *Ingestion) error
	GetIngestion(project string, name string) (*Ingestion, error)
	ListIngestion(project, logstore, name, displayName string, offset, size int) (ingestions []*Ingestion, total, count int, error error)
	DeleteIngestion(project string, name string) error

	// #################### Export #####################
	CreateExport(project string, export *Export) error
	UpdateExport(project string, export *Export) error
	GetExport(project, name string) (*Export, error)
	ListExport(project, logstore, name, displayName string, offset, size int) (exports []*Export, total, count int, error error)
	DeleteExport(project string, name string) error
	RestartExport(project string, export *Export) error

	// UpdateProjectPolicy updates project's policy.
	UpdateProjectPolicy(project string, policy string) error
	// DeleteProjectPolicy deletes project's policy.
	DeleteProjectPolicy(project string) error
	// GetProjectPolicy return project's policy.
	GetProjectPolicy(project string) (string, error)

	// #################### AlertPub Msg  #####################
	PublishAlertEvent(project string, alertResult []byte) error
}
