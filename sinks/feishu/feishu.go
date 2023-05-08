package feishu

import (
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/chyroc/lark"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	FEISHU_SINK             = "FeishuSink"
	FEISHU_BOT_ENDPOINT     = "https://open.feishu.cn/open-apis/bot/v2/hook/"
	WARNING             int = 2
	NORMAL              int = 1
	DEFAULT_MSG_TYPE        = "text"
	CONTENT_TYPE_JSON       = "application/json"
	LABEL_TEMPLATE          = "%s\n"
	TIME_FORMAT             = "2006-01-02 15:04:05"
)

type FeishuSink struct {
	Endpoint   string
	Namespaces []string
	Kinds      []string
	Level      int
	Labels     []string
	ClusterID  string
	Region     string
}

type SendBotMessageReq struct {
	MsgType string                   `json:"msg_type"`
	Card    *lark.MessageContentCard `json:"card"`
}

func (f *FeishuSink) Name() string {
	return FEISHU_SINK
}

func (f *FeishuSink) Stop() {

}

func (f *FeishuSink) ExportEvents(batch *core.EventBatch) {
	for _, event := range batch.Events {
		if f.isEventLevelDangerous(event.Type) {
			f.Send(event)
			// add threshold
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (f *FeishuSink) Send(event *v1.Event) {
	if f.Namespaces != nil {
		skip := true
		for _, namespace := range f.Namespaces {
			if namespace == event.Namespace {
				skip = false
				break
			}
		}
		if skip {
			return
		}
	}
	if f.Kinds != nil {
		skip := true
		for _, kind := range f.Kinds {
			if kind == event.InvolvedObject.Kind {
				skip = false
				break
			}
		}
		if skip {
			return
		}
	}
	card := f.NewFeishuMsgBuilder(event)
	if card == nil {
		klog.Warningf("failed to create msg from event,because of %v", event)
		return
	}
	req := SendBotMessageReq{
		MsgType: "interactive",
		Card:    card,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		klog.Warningf("failed to marshal feishu card %v", card)
		return
	}
	u := fmt.Sprintf("https://%s", f.Endpoint)
	resp, err := http.Post(u, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		klog.Errorf("failed to send msg to feishu. error: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp != nil && resp.StatusCode != http.StatusOK {
		klog.Errorf("failed to send msg to feishu, because the response code is %d", resp.StatusCode)
		return
	}

}

func (f *FeishuSink) isEventLevelDangerous(level string) bool {
	score := getLevel(level)
	if score >= f.Level {
		return true
	}
	return false
}

func NewFeishuSink(uri *url.URL) (*FeishuSink, error) {
	f := &FeishuSink{
		Level: WARNING,
	}
	if len(uri.Host) > 0 {
		f.Endpoint = uri.Host + uri.Path
	}
	opts := uri.Query()

	if len(opts["level"]) >= 1 {
		f.Level = getLevel(opts["level"][0])
	}
	//add extra labels
	if len(opts["label"]) >= 1 {
		f.Labels = opts["label"]
	}

	if clusterID := opts["cluster_id"]; len(clusterID) >= 1 {
		f.ClusterID = clusterID[0]
	}

	if region := opts["region"]; len(region) >= 1 {
		f.Region = region[0]
	}

	f.Namespaces = getValues(opts["namespaces"])
	// kinds:https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds
	// such as node,pod,component and so on
	f.Kinds = getValues(opts["kinds"])

	return f, nil
}

func getValues(o []string) []string {
	if len(o) >= 1 {
		if len(o[0]) == 0 {
			return nil
		}
		return strings.Split(o[0], ",")
	}
	return nil
}

func getLevel(level string) int {
	score := 0
	switch level {
	case v1.EventTypeWarning:
		score += 2
	case v1.EventTypeNormal:
		score += 1
	default:
		//score will remain 0
	}
	return score
}

func (f *FeishuSink) NewFeishuMsgBuilder(event *v1.Event) (msg *lark.MessageContentCard) {
	msg = &lark.MessageContentCard{
		Header: &lark.MessageContentCardHeader{
			Title: &lark.MessageContentCardObjectText{
				Tag:     "plain_text",
				Content: fmt.Sprintf("‼️ Kubernetes(集群: %s)事件通知", f.ClusterID),
			},
			Template: "red",
		},
		Config: &lark.MessageContentCardConfig{
			EnableForward: true,
		},
		I18NModules: &lark.MessageContentCardI18NModule{
			ZhCn: []lark.MessageContentCardModule{
				lark.MessageContentCardModuleDIV{
					Text: &lark.MessageContentCardObjectText{
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**事件等级: %s**`, event.Type),
					},
				},
				lark.MessageContentCardModuleDIV{
					Text: &lark.MessageContentCardObjectText{
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**事件类型: %s**`, event.InvolvedObject.Kind),
					},
				},
				lark.MessageContentCardModuleDIV{
					Text: &lark.MessageContentCardObjectText{
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**Namespace: %s**`, event.Namespace),
					},
				},
				lark.MessageContentCardModuleDIV{
					Text: &lark.MessageContentCardObjectText{
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**Reason: %s**`, event.Reason),
					},
				},
				lark.MessageContentCardModuleDIV{
					Text: &lark.MessageContentCardObjectText{
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**时间: %s**`, event.LastTimestamp.String()),
					},
				},
				lark.MessageContentCardModuleDIV{
					Text: &lark.MessageContentCardObjectText{
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**详细信息: %s**`, event.Message),
					},
				},
			},
		},
	}
	if len(event.Source.Host) < 1 {
		msg.I18NModules.ZhCn = append(msg.I18NModules.ZhCn, lark.MessageContentCardModuleDIV{
			Text: &lark.MessageContentCardObjectText{
				Tag:     "lark_md",
				Content: fmt.Sprintf(`**NodeName: %s**`, event.Source.Host),
			},
		})
	}
	return
}
