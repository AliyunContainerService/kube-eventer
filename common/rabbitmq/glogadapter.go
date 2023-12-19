package rabbitmq

import (
	"k8s.io/klog"
	_ "k8s.io/klog"
)

type Logging interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println( v ...interface{})
}

type GologAdapterLogger struct {
}

func (l GologAdapterLogger) Print(v ...interface{}) {
	klog.Info(v...)
}

func (l GologAdapterLogger) Printf(format string, v ...interface{}) {
	klog.Infof(format, v...)
}

func (l GologAdapterLogger) Println(v ...interface{}) {
	klog.Infoln(v...)
}
