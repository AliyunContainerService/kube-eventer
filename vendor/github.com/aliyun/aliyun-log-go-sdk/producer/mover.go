package producer

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.uber.org/atomic"
)

type Mover struct {
	moverShutDownFlag *atomic.Bool
	retryQueue        *RetryQueue
	ioWorker          *IoWorker
	logAccumulator    *LogAccumulator
	logger            log.Logger
	threadPool        *IoThreadPool
}

func initMover(logAccumulator *LogAccumulator, retryQueue *RetryQueue, ioWorker *IoWorker, logger log.Logger, threadPool *IoThreadPool) *Mover {
	mover := &Mover{
		moverShutDownFlag: atomic.NewBool(false),
		retryQueue:        retryQueue,
		ioWorker:          ioWorker,
		logAccumulator:    logAccumulator,
		logger:            logger,
		threadPool:        threadPool,
	}
	return mover

}

func (mover *Mover) sendToServer(key string, batch *ProducerBatch, config *ProducerConfig) {
	if value, ok := mover.logAccumulator.logGroupData[key]; !ok {
		return
	} else if GetTimeMs(time.Now().UnixNano())-value.createTimeMs < config.LingerMs {
		return
	}
	mover.threadPool.addTask(batch)
	delete(mover.logAccumulator.logGroupData, key)
}

func (mover *Mover) run(moverWaitGroup *sync.WaitGroup, config *ProducerConfig) {
	defer moverWaitGroup.Done()
	for !mover.moverShutDownFlag.Load() {
		sleepMs := config.LingerMs

		nowTimeMs := GetTimeMs(time.Now().UnixNano())
		mover.logAccumulator.lock.Lock()
		mapCount := len(mover.logAccumulator.logGroupData)
		for key, batch := range mover.logAccumulator.logGroupData {
			timeInterval := batch.createTimeMs + config.LingerMs - nowTimeMs
			if timeInterval <= 0 {
				level.Debug(mover.logger).Log("msg", "mover groutine execute sent producerBatch to IoWorker")
				mover.sendToServer(key, batch, config)
			} else {
				if sleepMs > timeInterval {
					sleepMs = timeInterval
				}
			}
		}
		mover.logAccumulator.lock.Unlock()

		if mapCount == 0 {
			level.Debug(mover.logger).Log("msg", "No data time in map waiting for user configured RemainMs parameter values")
			sleepMs = config.LingerMs
		}

		retryProducerBatchList := mover.retryQueue.getRetryBatch(mover.moverShutDownFlag.Load())
		if retryProducerBatchList == nil {
			// If there is nothing to send in the retry queue, just wait for the minimum time that was given to me last time.
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		} else {
			count := len(retryProducerBatchList)
			for i := 0; i < count; i++ {
				mover.threadPool.addTask(retryProducerBatchList[i])
			}
		}

	}
	mover.logAccumulator.lock.Lock()
	for _, batch := range mover.logAccumulator.logGroupData {
		mover.threadPool.addTask(batch)
	}
	mover.logAccumulator.logGroupData = make(map[string]*ProducerBatch)
	mover.logAccumulator.lock.Unlock()

	producerBatchList := mover.retryQueue.getRetryBatch(mover.moverShutDownFlag.Load())
	count := len(producerBatchList)
	for i := 0; i < count; i++ {
		mover.threadPool.addTask(producerBatchList[i])
	}
	level.Info(mover.logger).Log("msg", "mover thread closure complete")
}
