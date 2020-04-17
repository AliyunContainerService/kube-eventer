package webhook

import (
	"net/url"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/common/filters"
	"github.com/AliyunContainerService/kube-eventer/common/kubernetes"
	"k8s.io/klog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WebhookSinkName = "webhook"
	Warning         = "Warning"
	Normal          = "Normal"
)

var (
	defaultBodyTemplate = ``
)

type WebhookSink struct {
	filters                map[string]filters.Filter
	endpoint               string
	level                  string
	method                 string
	headers                []string
	bodyConfigMapName      string
	bodyConfigMapNamespace string
	bodyTemplate           string
}

func (ws *WebhookSink) Name() string {
	return WebhookSinkName
}

func (ws *WebhookSink) ExportEvents(batch *core.EventBatch) {
	
}

func (ws *WebhookSink) Stop() {
	// not implement
	return
}

func NewWebhookSink(uri *url.URL) (*WebhookSink, error) {
	s := &WebhookSink{
		level:   Warning,
		filters: make(map[string]filters.Filter),
	}
	if len(uri.Host) > 0 {
		s.endpoint = uri.Host + uri.Path
	}
	opts := uri.Query()

	if len(opts["method"]) >= 1 {
		s.method = opts["method"][0]
	}

	// set header of webhook
	s.headers = opts["header"]

	if len(opts["level"]) >= 1 {
		s.level = opts["level"][0]
	}
	s.filters["LevelFilter"] = filters.NewGenericFilter("Type", []string{s.level}, false)

	namespaces := filters.GetValues(opts["namespaces"])
	// kinds:https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds
	s.filters["NamespacesFilter"] = filters.NewGenericFilter("Namespace", namespaces, true)

	// such as node,pod,component and so on
	kinds := filters.GetValues(opts["kinds"])
	s.filters["KindsFilter"] = filters.NewGenericFilter("Kind", kinds, false)

	// reason filter
	reasons := opts["reason"]
	s.filters["ReasonsFilter"] = filters.NewGenericFilter("Reason", reasons, true)

	if len(opts["custom_body_configmap"]) >= 1 {
		s.bodyConfigMapName = opts["custom_body_configmap"][0]

		if len(opts["custom_body_configmap_namespace"]) >= 1 {
			s.bodyConfigMapNamespace = opts["custom_body_configmap_namespace"][0]
		} else {
			s.bodyConfigMapNamespace = "default"
		}

		client, err := kubernetes.GetKubernetesClient(nil)
		if err != nil {
			klog.Warning("Failed to get kubernetes client and use default bodyTemplate instead")
			s.bodyTemplate = defaultBodyTemplate
			return s, nil
		}
		configmap, err := client.CoreV1().ConfigMaps(s.bodyConfigMapNamespace).Get(s.bodyConfigMapName, metav1.GetOptions{})
		if err != nil {
			klog.Warning("Failed to get configMap %s in namespace %s and use default bodyTemplate instead,because of %v", s.bodyConfigMapName, s.bodyConfigMapNamespace, err)
			s.bodyTemplate = defaultBodyTemplate
			return s, nil
		}
		if content, ok := configmap.Data["body"]; !ok {
			klog.Warning("Failed to get configMap content and use default bodyTemplate instead,because of %v", content, err)
			s.bodyTemplate = defaultBodyTemplate
			return s, nil
		}else{
			s.bodyTemplate = content
		}
	}

	return s, nil
}
