package rabbitmq

import (
	"k8s.io/klog"
)

type GologAdapterLogger struct {
}

func (l GologAdapterLogger) Printf(format string, v ...interface{}) {
	klog.Infof(format, v...)
}
