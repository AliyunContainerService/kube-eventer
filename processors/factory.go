package processors

import (
	"fmt"

	"github.com/AliyunContainerService/kube-eventer/common/flags"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/processors/npd"
	"k8s.io/klog"
)

// ProcessorFactory is used create processors
type ProcessorFactory struct {
}

func (pf *ProcessorFactory) build(uri flags.Uri) (core.EventProcessor, error) {
	switch uri.Key {
	case "npd":
		processor, err := npd.NewProcessor(&uri.Val)
		return processor, err
	default:
		return nil, fmt.Errorf("Processor not recognized: %s", uri.Key)
	}
}

func (pf *ProcessorFactory) getDefaultProcessors() ([]core.EventProcessor, error) {
	result := []core.EventProcessor{}
	npdProcessor, err := npd.NewProcessor(nil)
	result = append(result, npdProcessor)
	return result, err
}

// BuildAll build all processors
func (pf *ProcessorFactory) BuildAll(uris flags.Uris) ([]core.EventProcessor, error) {
	if len(uris) == 0 {
		return pf.getDefaultProcessors()
	}
	result := []core.EventProcessor{}
	for _, uri := range uris {
		processor, err := pf.build(uri)
		if err != nil {
			klog.Errorf("Failed to create processor %s: %v", uri.Key, err)
			return nil, err
		}
		result = append(result, processor)
	}
	return result, nil
}

// NewProcessorFactory return a processor factory instance
func NewProcessorFactory() *ProcessorFactory {
	return &ProcessorFactory{}
}
