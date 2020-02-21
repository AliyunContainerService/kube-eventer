package dingtalk

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	MARKDOWN_MSG_TYPE      = "markdown"
	MARKDOWN_TEMPLATE      = "Level: %s \n\nKind: %s \n\nNamespace: %s \n\nName: %s \n\nReason: %s \n\nTimestamp: %s \n\nMessage: %s"
	MARKDOWN_LINK_TEMPLATE = "[%s](%s)"
	MARKDOWN_TEXT_BOLD     = "**%s**"
	MARKDOWN_NEW_LINE      = "\n\n"

	URL_ALIYUN_K8S_CONSULE = "https://cs.console.aliyun.com/#/k8s"
	//阿里云 kubernetes 管理控制台, Deployment,StatefulSet,DaemonSet 有同样的URL规律
	URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE = URL_ALIYUN_K8S_CONSULE + "/%s/detail/%s/%s/%s/%s/pods"
	URL_ALIYUN_POD_TEMPLATE             = URL_ALIYUN_K8S_CONSULE + "/pod/%s/%s/%s/container"
	URL_ALIYUN_CROBJOB_TEMPLATE         = URL_ALIYUN_K8S_CONSULE + "/cronjob/detail/%s/%s/%s/%s/jobs"
	URL_ALIYUN_SVC_TEMPLATE             = URL_ALIYUN_K8S_CONSULE + "/service/detail/%s/%s/%s/%s"
	URL_ALIYUN_NAMESPACE_TEMPLATE       = URL_ALIYUN_K8S_CONSULE + "/namespace"
	URL_ALIYUN_ECS_TEMPLATE             = "https://ecs.console.aliyun.com/#/server/%s/detail?regionId=%s"
)

type MarkdownMsgBuilder struct {
	Labels     []string
	Region     string
	ClusterID  string
	OutputText string
}

func NewMarkdownMsgBuilder(clusterID, region string, event *v1.Event) *MarkdownMsgBuilder {

	m := MarkdownMsgBuilder{
		Region:    region,
		ClusterID: clusterID,
	}

	level := fmt.Sprintf(MARKDOWN_TEXT_BOLD, event.Type)
	kind := fmt.Sprintf(MARKDOWN_TEXT_BOLD, event.InvolvedObject.Kind)
	namespace := fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Namespace, URL_ALIYUN_NAMESPACE_TEMPLATE)
	name := ""

	switch event.InvolvedObject.Kind {
	case "Deployment":
		deployName := removeDotContent(event.Name)
		podsURL := fmt.Sprintf(URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE, "deployment", m.Region, m.ClusterID, event.Namespace, deployName)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsURL)
		break
	case "Pod":
		podName := removeDotContent(event.Name)
		podsURL := fmt.Sprintf(URL_ALIYUN_POD_TEMPLATE, m.ClusterID, event.Namespace, podName)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsURL)
		break
	case "StatefulSet":
		ssName := removeDotContent(event.Name)
		podsURL := fmt.Sprintf(URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE, "statefulset", m.Region, m.ClusterID, event.Namespace, ssName)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsURL)
		break
	case "DaemonSet":
		dsName := removeDotContent(event.Name)
		podsURL := fmt.Sprintf(URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE, "daemonset", m.Region, m.ClusterID, event.Namespace, dsName)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsURL)
		break
	case "CronJob":
		jobName := removeDotContent(event.Name)
		jobURL := fmt.Sprintf(URL_ALIYUN_CROBJOB_TEMPLATE, m.Region, m.ClusterID, event.Namespace, jobName)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, jobURL)
		break
	case "Service":
		serviceName := removeDotContent(event.Name)
		svcURL := fmt.Sprintf(URL_ALIYUN_SVC_TEMPLATE, m.Region, m.ClusterID, event.Namespace, serviceName)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, svcURL)
		break
		//fixme:覆盖所有 event.InvolvedObject.Kind
	default:
		name = event.Name
		break
	}
	reason := fmt.Sprintf(MARKDOWN_TEXT_BOLD, event.Reason)
	timestamp := fmt.Sprintf(MARKDOWN_TEXT_BOLD, event.LastTimestamp.String())
	message := fmt.Sprintf(MARKDOWN_TEXT_BOLD, event.Message)
	m.OutputText = fmt.Sprintf(MARKDOWN_TEMPLATE, level, kind, namespace, name, reason, timestamp, message)
	return &m

}

// removeDotContent 每个 Event 由 <resource>.<UnixNano> 组成,需要去掉.后面的部分,得到 <resource>
func removeDotContent(s string) string {
	if dotPosition := strings.Index(s, "."); dotPosition > -1 {
		s = s[:dotPosition]
	}
	return s
}

func (m *MarkdownMsgBuilder) AddLabels(labels []string) {
	if labels != nil && len(labels) > 0 {
		for i := len(labels) - 1; i >= 0; i-- {
			if label := strings.TrimSpace(labels[i]); len(label) > 0 {
				m.OutputText = fmt.Sprintf("label[%d]: **%s**"+MARKDOWN_NEW_LINE, i, labels[i]) + m.OutputText
			}
		}
	}
}

func (m *MarkdownMsgBuilder) AddNodeName(nodeName string) {
	if len(nodeName) < 1 {
		return
	}
	ecsInfo := strings.Split(nodeName, ".")
	var nodeInfo string
	if len(ecsInfo) > 1 {
		ecsURL := fmt.Sprintf(URL_ALIYUN_ECS_TEMPLATE, ecsInfo[1], ecsInfo[0])
		nodeInfo = fmt.Sprintf("Node: "+MARKDOWN_LINK_TEMPLATE+" "+MARKDOWN_NEW_LINE, nodeName, ecsURL)
	} else {
		nodeInfo = fmt.Sprintf("Node: "+MARKDOWN_TEXT_BOLD+" "+MARKDOWN_NEW_LINE, nodeName)
	}
	m.OutputText = nodeInfo + m.OutputText
}

func (m *MarkdownMsgBuilder) Build() string {
	return m.OutputText
}
