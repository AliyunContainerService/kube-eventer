package producer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gogo/protobuf/proto"
)

type ProducerBatch struct {
	totalDataSize        int64
	lock                 sync.RWMutex
	logGroup             *sls.LogGroup
	logGroupSize         int
	logGroupCount        int
	attemptCount         int
	baseRetryBackoffMs   int64
	nextRetryMs          int64
	maxRetryIntervalInMs int64
	callBackList         []CallBack
	createTimeMs         int64
	maxRetryTimes        int
	project              string
	logstore             string
	shardHash            *string
	result               *Result
	maxReservedAttempts  int
	useMetricStoreUrl    bool
}

func generatePackId(source string) string {
	srcData := source + time.Now().String()
	return ToMd5(srcData)[0:16]
}

func initProducerBatch(logData interface{}, callBackFunc CallBack, project, logstore, logTopic, logSource, shardHash string, config *ProducerConfig) *ProducerBatch {
	logs := []*sls.Log{}

	if log, ok := logData.(*sls.Log); ok {
		logs = append(logs, log)
	} else if logList, ok := logData.([]*sls.Log); ok {
		logs = append(logs, logList...)
	}

	logGroup := &sls.LogGroup{
		Logs:    logs,
		LogTags: config.LogTags,
		Topic:   proto.String(logTopic),
		Source:  proto.String(logSource),
	}
	if config.GeneratePackId {
		config.packLock.Lock()
		if config.packPrefix == "" {
			config.packPrefix = strings.ToUpper(generatePackId(logSource)) + "-"
		}
		packStr := config.packPrefix + fmt.Sprintf("%X", config.packNumber)
		logGroup.LogTags = append(logGroup.LogTags, &sls.LogTag{
			Key:   proto.String("__pack_id__"),
			Value: proto.String(packStr),
		})
		config.packNumber++
		config.packLock.Unlock()
	}
	currentTimeMs := GetTimeMs(time.Now().UnixNano())
	producerBatch := &ProducerBatch{
		logGroup:             logGroup,
		attemptCount:         0,
		maxRetryIntervalInMs: config.MaxRetryBackoffMs,
		callBackList:         []CallBack{},
		createTimeMs:         currentTimeMs,
		maxRetryTimes:        config.Retries,
		baseRetryBackoffMs:   config.BaseRetryBackoffMs,
		project:              project,
		logstore:             logstore,
		result:               initResult(),
		maxReservedAttempts:  config.MaxReservedAttempts,
		useMetricStoreUrl:    config.UseMetricStoreURL,
	}
	if shardHash == "" {
		producerBatch.shardHash = nil
	} else {
		producerBatch.shardHash = &shardHash
	}
	producerBatch.totalDataSize = int64(producerBatch.logGroup.Size())

	if callBackFunc != nil {
		producerBatch.callBackList = append(producerBatch.callBackList, callBackFunc)
	}
	return producerBatch
}

func (producerBatch *ProducerBatch) getProject() string {
	defer producerBatch.lock.RUnlock()
	producerBatch.lock.RLock()
	return producerBatch.project
}

func (producerBatch *ProducerBatch) getLogstore() string {
	defer producerBatch.lock.RUnlock()
	producerBatch.lock.RLock()
	return producerBatch.logstore
}

func (producerBatch *ProducerBatch) getShardHash() *string {
	defer producerBatch.lock.RUnlock()
	producerBatch.lock.RLock()
	return producerBatch.shardHash
}

func (producerBatch *ProducerBatch) getLogGroupCount() int {
	defer producerBatch.lock.RUnlock()
	producerBatch.lock.RLock()
	return len(producerBatch.logGroup.GetLogs())
}

func (producerBatch *ProducerBatch) isUseMetricStoreUrl() bool {
	defer producerBatch.lock.RUnlock()
	producerBatch.lock.RLock()
	return producerBatch.useMetricStoreUrl
}

func (producerBatch *ProducerBatch) addLogToLogGroup(log interface{}) {
	defer producerBatch.lock.Unlock()
	producerBatch.lock.Lock()
	if mlog, ok := log.(*sls.Log); ok {
		producerBatch.logGroup.Logs = append(producerBatch.logGroup.Logs, mlog)
	} else if logList, ok := log.([]*sls.Log); ok {
		producerBatch.logGroup.Logs = append(producerBatch.logGroup.Logs, logList...)
	}
}

func (producerBacth *ProducerBatch) addProducerBatchCallBack(callBack CallBack) {
	defer producerBacth.lock.Unlock()
	producerBacth.lock.Lock()
	producerBacth.callBackList = append(producerBacth.callBackList, callBack)
}
