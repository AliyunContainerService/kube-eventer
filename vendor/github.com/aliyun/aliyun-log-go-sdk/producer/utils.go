package producer

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gogo/protobuf/proto"
)

func GenerateLog(logTime uint32, addLogMap map[string]string) *sls.Log {

	content := []*sls.LogContent{}
	for key, value := range addLogMap {
		content = append(content, &sls.LogContent{
			Key:   proto.String(key),
			Value: proto.String(value),
		})
	}
	return &sls.Log{
		Time:     proto.Uint32(logTime),
		Contents: content,
	}
}

func GetTimeMs(t int64) int64 {
	return t / 1000 / 1000
}

func GetLogSizeCalculate(log *sls.Log) int {
	sizeInBytes := 4
	logContent := log.GetContents()
	count := len(logContent)
	for i := 0; i < count; i++ {
		sizeInBytes += len(*logContent[i].Value)
		sizeInBytes += len(*logContent[i].Key)
	}

	return sizeInBytes

}

func GetLogListSize(logList []*sls.Log) int {
	sizeInBytes := 0
	for _, log := range logList {
		sizeInBytes += GetLogSizeCalculate(log)
	}
	return sizeInBytes
}
