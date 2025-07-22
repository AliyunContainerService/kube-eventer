package producer

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	uberatomic "go.uber.org/atomic"
)

type CallBack interface {
	Success(result *Result)
	Fail(result *Result)
}

type IoWorker struct {
	taskCount              int64
	client                 sls.ClientInterface
	retryQueue             *RetryQueue
	retryQueueShutDownFlag *uberatomic.Bool
	logger                 log.Logger
	maxIoWorker            chan int64
	noRetryStatusCodeMap   map[int]*string
	producer               *Producer
}

func initIoWorker(client sls.ClientInterface, retryQueue *RetryQueue, logger log.Logger, maxIoWorkerCount int64, errorStatusMap map[int]*string, producer *Producer) *IoWorker {
	return &IoWorker{
		client:                 client,
		retryQueue:             retryQueue,
		taskCount:              0,
		retryQueueShutDownFlag: uberatomic.NewBool(false),
		logger:                 logger,
		maxIoWorker:            make(chan int64, maxIoWorkerCount),
		noRetryStatusCodeMap:   errorStatusMap,
		producer:               producer,
	}
}

func (ioWorker *IoWorker) sendToServer(producerBatch *ProducerBatch) {
	level.Debug(ioWorker.logger).Log("msg", "ioworker send data to server")
	beginMs := GetTimeMs(time.Now().UnixNano())
	var err error
	if producerBatch.isUseMetricStoreUrl() {
		// not use compress type now
		err = ioWorker.client.PutLogsWithMetricStoreURL(producerBatch.getProject(), producerBatch.getLogstore(), producerBatch.logGroup)
	} else {
		req := &sls.PostLogStoreLogsRequest{
			LogGroup:     producerBatch.logGroup,
			HashKey:      producerBatch.getShardHash(),
			CompressType: ioWorker.producer.producerConfig.CompressType,
		}
		err = ioWorker.client.PostLogStoreLogsV2(producerBatch.getProject(), producerBatch.getLogstore(), req)
	}
	if err == nil {
		level.Debug(ioWorker.logger).Log("msg", "sendToServer suecssed,Execute successful callback function")
		if producerBatch.attemptCount < producerBatch.maxReservedAttempts {
			nowMs := GetTimeMs(time.Now().UnixNano())
			attempt := createAttempt(true, "", "", "", nowMs, nowMs-beginMs)
			producerBatch.result.attemptList = append(producerBatch.result.attemptList, attempt)
		}
		producerBatch.result.successful = true
		// After successful delivery, producer removes the batch size sent out
		atomic.AddInt64(&ioWorker.producer.producerLogGroupSize, -producerBatch.totalDataSize)
		if len(producerBatch.callBackList) > 0 {
			for _, callBack := range producerBatch.callBackList {
				callBack.Success(producerBatch.result)
			}
		}
	} else {
		if ioWorker.retryQueueShutDownFlag.Load() {
			if len(producerBatch.callBackList) > 0 {
				for _, callBack := range producerBatch.callBackList {
					ioWorker.addErrorMessageToBatchAttempt(producerBatch, err, false, beginMs)
					callBack.Fail(producerBatch.result)
				}
			}
			return
		}
		level.Info(ioWorker.logger).Log("msg", "sendToServer failed", "error", err)
		if slsError, ok := err.(*sls.Error); ok {
			if _, ok := ioWorker.noRetryStatusCodeMap[int(slsError.HTTPCode)]; ok {
				ioWorker.addErrorMessageToBatchAttempt(producerBatch, err, false, beginMs)
				ioWorker.excuteFailedCallback(producerBatch)
				return
			}
		}
		if producerBatch.attemptCount < producerBatch.maxRetryTimes {
			ioWorker.addErrorMessageToBatchAttempt(producerBatch, err, true, beginMs)
			retryWaitTime := producerBatch.baseRetryBackoffMs * int64(math.Pow(2, float64(producerBatch.attemptCount)-1))
			if retryWaitTime < producerBatch.maxRetryIntervalInMs {
				producerBatch.nextRetryMs = GetTimeMs(time.Now().UnixNano()) + retryWaitTime
			} else {
				producerBatch.nextRetryMs = GetTimeMs(time.Now().UnixNano()) + producerBatch.maxRetryIntervalInMs
			}
			level.Debug(ioWorker.logger).Log("msg", "Submit to the retry queue after meeting the retry criteria。")
			ioWorker.retryQueue.sendToRetryQueue(producerBatch, ioWorker.logger)
		} else {
			ioWorker.excuteFailedCallback(producerBatch)
		}
	}
}

func (ioWorker *IoWorker) addErrorMessageToBatchAttempt(producerBatch *ProducerBatch, err error, retryInfo bool, beginMs int64) {
	if producerBatch.attemptCount < producerBatch.maxReservedAttempts {
		slsError := err.(*sls.Error)
		if retryInfo {
			level.Info(ioWorker.logger).Log("msg", "sendToServer failed,start retrying", "retry times", producerBatch.attemptCount, "requestId", slsError.RequestID, "error code", slsError.Code, "error message", slsError.Message)
		}
		nowMs := GetTimeMs(time.Now().UnixNano())
		attempt := createAttempt(false, slsError.RequestID, slsError.Code, slsError.Message, nowMs, nowMs-beginMs)
		producerBatch.result.attemptList = append(producerBatch.result.attemptList, attempt)
	}
	producerBatch.result.successful = false
	producerBatch.attemptCount += 1
}

func (ioWorker *IoWorker) closeSendTask(ioWorkerWaitGroup *sync.WaitGroup) {
	<-ioWorker.maxIoWorker
	atomic.AddInt64(&ioWorker.taskCount, -1)
	ioWorkerWaitGroup.Done()
}

func (ioWorker *IoWorker) startSendTask(ioWorkerWaitGroup *sync.WaitGroup) {
	atomic.AddInt64(&ioWorker.taskCount, 1)
	ioWorker.maxIoWorker <- 1
	ioWorkerWaitGroup.Add(1)
}

func (ioWorker *IoWorker) excuteFailedCallback(producerBatch *ProducerBatch) {
	level.Info(ioWorker.logger).Log("msg", "sendToServer failed,Execute failed callback function")
	atomic.AddInt64(&ioWorker.producer.producerLogGroupSize, -producerBatch.totalDataSize)
	if len(producerBatch.callBackList) > 0 {
		for _, callBack := range producerBatch.callBackList {
			callBack.Fail(producerBatch.result)
		}
	}
}
