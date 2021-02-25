package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/core"
	kube_api "k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/http"
	"sync"
	"time"
)

type BufferEventBatch map[string][]*kube_api.Event

func (d *DingTalkSink) ExportBufferEvents(batch *core.EventBatch) {

	var wg sync.WaitGroup
	var bufferEventBatch = BufferEventBatch{}
	defer func() {
		bufferEventBatch = BufferEventBatch{}
	}()
	// dump level is error event into buffer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, event := range batch.Events {
			// only handler Warning Buffer
			if event.Type == "Warning" {
				bufferEventBatch[event.InvolvedObject.Name] = append(bufferEventBatch[event.InvolvedObject.Name], event)
			}
		}
	}()

	//buffer windows
	klog.V(2).Info("dingding buffer windows is ", ArgDDbufferWindows)
	time.Sleep(ArgDDbufferWindows)
	klog.V(2).Info("NewEventBatch len:", len(bufferEventBatch))

	for _, bufferEvent := range bufferEventBatch {
		d.DingBuffer(bufferEvent)
		// add threshold
		time.Sleep(time.Millisecond * 50)
	}

	wg.Wait()
}

func (d *DingTalkSink) DingBuffer(bufferevent []*kube_api.Event) {

	msg := NewcreateMsgFromEvent(d, bufferevent)

	if msg == nil {
		klog.Warningf("failed to create msg from event,because of %v", bufferevent)
		return
	}

	msg_bytes, err := json.Marshal(msg)
	if err != nil {
		klog.Warningf("failed to marshal msg %v", msg)
		return
	}

	b := bytes.NewBuffer(msg_bytes)

	resp, err := http.Post(fmt.Sprintf("https://%s?access_token=%s", d.Endpoint, d.Token), CONTENT_TYPE_JSON, b)
	if err != nil {
		klog.Errorf("failed to send msg to dingtalk. error: %s", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp != nil && resp.StatusCode != http.StatusOK {
		klog.Errorf("failed to send msg to dingtalk, because the response code is %d", resp.StatusCode)
		return
	}
}

func NewcreateMsgFromEvent(d *DingTalkSink, bufferevent []*kube_api.Event) *DingTalkMsg {
	msg := &DingTalkMsg{}
	msg.MsgType = d.MsgType

	m := ""
	m2 := ""
	i := 0
	for _, event := range bufferevent {
		i = i + 1
		m = m + fmt.Sprintf("msg%d : ", i) + event.Message + "\n" + "  "
		m2 = m2 + "#### " + fmt.Sprintf("msg%d : ", i) + event.Message + "\n" + "  "
	}
	msgs := fmt.Sprintf("[%s]", m)
	msgs_markdown := fmt.Sprintf("[\n%s]", m2)

	switch msg.MsgType {
	//https://open-doc.dingtalk.com/microapp/serverapi2/ye8tup#-6
	case MARKDOWN_MSG_TYPE:
		markdownCreator := NewMarkdownMsgBuilder(d.ClusterID, d.Region, bufferevent[0], msgs_markdown)
		markdownCreator.AddNodeName(bufferevent[0].Source.Host)
		markdownCreator.AddLabels(d.Labels)
		msg.Markdown = DingTalkMarkdown{
			//title 加不加其实没所谓,最终不会显示
			Title: fmt.Sprintf("Kubernetes(ID:%s) Event", d.ClusterID),
			Text:  markdownCreator.Build(),
		}
		break

	default:
		//默认按文本模式推送
		template := MSG_TEMPLATE
		if len(d.Labels) > 0 {
			for _, label := range d.Labels {
				template = fmt.Sprintf(LABE_TEMPLATE, label) + template
			}
		}

		event := bufferevent[0]
		msg.Text = DingTalkText{
			Content: fmt.Sprintf(template, event.Type, event.InvolvedObject.Kind, event.Namespace, event.InvolvedObject.Name, event.Reason, event.LastTimestamp.Format(TIME_FORMAT), msgs),
		}
	}

	return msg
}
