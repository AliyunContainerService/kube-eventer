package dingtalk

import (
	"fmt"
	"k8s.io/api/core/v1"
)

const (
	MARKDOWN_MSG_TYPE      = "markdown"
	MARKDOWN_TEMPLATE      = "Level:%s \n\nKind:%s \n\nNamespace:%s \n\nName:%s \n\nReason:%s \n\nTimestamp:%s \n\nMessage:%s"
	MARKDOWN_LINK_TEMPLATE = "[%s](%s)"
	MARKDOWN_TEXT_BOLD     = "**%s**"

	URL_ALIYUN_K8S_CONSULE = "https://cs.console.aliyun.com/#/k8s"
	//阿里云 kubernetes 管理控制台, Deployment,StatefulSet,DaemonSet 有同样的URL规律
	URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE = URL_ALIYUN_K8S_CONSULE + "/%s/detail/%s/%s/%s/%s/pods"
	URL_ALIYUN_POD_TEMPLATE             = URL_ALIYUN_K8S_CONSULE + "/pod/%s/%s/%s/container"
	URL_ALIYUN_CROBJOB_TEMPLATE         = URL_ALIYUN_K8S_CONSULE + "/cronjob/detail/%s/%s/%s/%s/jobs"
	URL_ALIYUN_NAMESPACE_TEMPLATE       = URL_ALIYUN_K8S_CONSULE + "/namespace"
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
		podsUrl := fmt.Sprintf(URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE, "deployment", m.Region, m.ClusterID, event.Namespace, event.Name)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsUrl)
		break
	case "Pod":
		podsUrl := fmt.Sprintf(URL_ALIYUN_POD_TEMPLATE, m.ClusterID, event.Namespace, event.Name)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsUrl)
		break
	case "StatefulSet":
		podsUrl := fmt.Sprintf(URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE, "statefulset", m.Region, m.ClusterID, event.Namespace, event.Name)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsUrl)
		break
	case "DaemonSet":
		podsUrl := fmt.Sprintf(URL_ALIYUN_RESOURCE_DETAIL_TEMPLATE, "daemonset", m.Region, m.ClusterID, event.Namespace, event.Name)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, podsUrl)
		break
	case "CronJob":
		jobUrl := fmt.Sprintf(URL_ALIYUN_CROBJOB_TEMPLATE, m.Region, m.ClusterID, event.Namespace, event.Name)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, jobUrl)
		break
	case "Service":
		svcUrl := fmt.Sprintf("%s/service/detail/%s/%s/%s/%s", URL_ALIYUN_K8S_CONSULE, m.Region, m.ClusterID, event.Namespace, event.Name)
		name = fmt.Sprintf(MARKDOWN_LINK_TEMPLATE, event.Name, svcUrl)
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

func (m *MarkdownMsgBuilder) AddLabels(labels []string) {
	if len(labels) > 0 {
		for i := len(labels) - 1; i >= 0; i-- {
			m.OutputText = fmt.Sprintf("label[%d]:**%s**\n\n", i, labels[i]) + m.OutputText
		}

	}
}

func (m *MarkdownMsgBuilder) Build() string {
	return m.OutputText
}
