package webhook

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/AliyunContainerService/kube-eventer/common/filters"
	"github.com/AliyunContainerService/kube-eventer/common/kubernetes"
	"github.com/AliyunContainerService/kube-eventer/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	SinkName = "webHook"
	Warning  = "Warning"
	Normal   = "Normal"
)

var (
	// body template of event
	defaultBodyTemplate = `
{
	"EventType": "{{ .Type }}",
	"EventKind": "{{ .InvolvedObject.Kind }}",
	"EventReason": "{{ .Reason }}",
	"EventTime": "{{ .LastTimestamp }}",
	"EventMessage": "{{ .Message }}"
}`
)

type WebHookSink struct {
	filters                map[string]filters.Filter
	headerMap              map[string]string
	endpoint               string
	method                 string
	bodyTemplate           string
	bodyConfigMapName      string
	bodyConfigMapNamespace string
}

func (ws *WebHookSink) Name() string {
	return SinkName
}

func (ws *WebHookSink) ExportEvents(batch *core.EventBatch) {
	for _, event := range batch.Events {
		err := ws.Send(event)
		if err != nil {
			klog.Warningf("Failed to send event to WebHook sink,because of %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// send msg to generic webHook
func (ws *WebHookSink) Send(event *v1.Event) (err error) {
	for _, v := range ws.filters {
		if !v.Filter(event) {
			return
		}
	}

	body, err := ws.RenderBodyTemplate(event)
	if err != nil {
		klog.Errorf("Failed to RenderBodyTemplate,because of %v", err)
		return err
	}

	bodyBuffer := bytes.NewBuffer([]byte(body))
	req, err := http.NewRequest(ws.method, ws.endpoint, bodyBuffer)

	// append header to http request
	if ws.headerMap != nil && len(ws.headerMap) != 0 {
		for k, v := range ws.headerMap {
			req.Header.Set(k, v)
		}
	}

	if err != nil {
		klog.Errorf("Failed to create request,because of %v", err)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		klog.Errorf("Failed to send event to sink,because of %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp != nil && resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = fmt.Errorf("failed to send msg to sink, because the response code is %d, body is : %v", resp.StatusCode, string(body))
		klog.Errorln(err)
		return err
	}
	return nil
}

func (ws *WebHookSink) RenderBodyTemplate(event *v1.Event) (body string, err error) {
	var tpl bytes.Buffer
	tp, err := template.New("body").Parse(ws.bodyTemplate)
	if err != nil {
		klog.Errorf("Failed to parse template,because of %v", err)
		return "", err
	}
	event.Message = strings.Replace(event.Message, `"`, ``, -1)
	if err := tp.Execute(&tpl, event); err != nil {
		klog.Errorf("Failed to renderTemplate,because of %v", err)
		return "", err
	}
	return tpl.String(), nil
}

func (ws *WebHookSink) Stop() {
	// not implement
	return
}

func getLevels(level string) []string {
	switch level {
	case Normal:
		return []string{Normal, Warning}
	case Warning:
		return []string{Warning}
	}
	return []string{Warning}
}

// init WebHookSink with url params
func NewWebHookSink(uri *url.URL) (*WebHookSink, error) {
	s := &WebHookSink{
		// default http method
		method:       http.MethodGet,
		bodyTemplate: defaultBodyTemplate,
		filters:      make(map[string]filters.Filter),
	}

	if len(uri.Host) > 0 {
		s.endpoint = uri.String()
	} else {
		klog.Errorf("uri host's length is 0 and pls check your uri: %v", uri)
		return nil, fmt.Errorf("uri host is not valid.url: %v", uri)
	}

	opts := uri.Query()

	if len(opts["method"]) >= 1 {
		s.method = opts["method"][0]
	}

	// set header of webHook
	s.headerMap = parseHeaders(opts["header"])

	level := Warning
	if len(opts["level"]) >= 1 {
		level = opts["level"][0]
		s.filters["LevelFilter"] = filters.NewGenericFilter("Type", getLevels(level), false)
	}

	if len(opts["namespaces"]) >= 1 {
		// namespace filter doesn't support regexp
		namespaces := filters.GetValues(opts["namespaces"])
		s.filters["NamespacesFilter"] = filters.NewGenericFilter("Namespace", namespaces, false)
	}

	if len(opts["kinds"]) >= 1 {
		// such as node,pod,component and so on
		// kinds:https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds
		kinds := filters.GetValues(opts["kinds"])
		s.filters["KindsFilter"] = filters.NewGenericFilter("Kind", kinds, false)
	}

	if len(opts["reason"]) >= 1 {
		// reason filter support regexp.
		reasons := filters.GetValues(opts["reason"])
		s.filters["ReasonsFilter"] = filters.NewGenericFilter("Reason", reasons, true)
	}

	if len(opts["custom_body_configmap"]) >= 1 {
		s.bodyConfigMapName = opts["custom_body_configmap"][0]

		if len(opts["custom_body_configmap_namespace"]) >= 1 {
			s.bodyConfigMapNamespace = opts["custom_body_configmap_namespace"][0]
		} else {
			s.bodyConfigMapNamespace = "default"
		}

		client, err := kubernetes.GetKubernetesClient(nil)
		if err != nil {
			klog.Warningf("Failed to get kubernetes client and use default bodyTemplate instead")
			s.bodyTemplate = defaultBodyTemplate
			return s, nil
		}
		configmap, err := client.CoreV1().ConfigMaps(s.bodyConfigMapNamespace).Get(s.bodyConfigMapName, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Failed to get configMap %s in namespace %s and use default bodyTemplate instead,because of %v", s.bodyConfigMapName, s.bodyConfigMapNamespace, err)
			s.bodyTemplate = defaultBodyTemplate
			return s, nil
		}
		if content, ok := configmap.Data["content"]; !ok {
			klog.Warningf("Failed to get configMap content and use default bodyTemplate instead,because of %v", err)
			s.bodyTemplate = defaultBodyTemplate
			return s, nil
		} else {
			s.bodyTemplate = content
		}
	}

	return s, nil
}

func parseHeaders(headers []string) map[string]string {
	headerMap := make(map[string]string)
	for _, h := range headers {
		if arr := strings.Split(h, "="); len(arr) == 2 {
			headerMap[arr[0]] = arr[1]
		}
	}
	return headerMap
}
