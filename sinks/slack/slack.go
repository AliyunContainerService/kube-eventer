// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slack

import (
	"fmt"
	apps_v1beta1 "k8s.io/api/apps/v1beta1"
	batch_v1 "k8s.io/api/batch/v1"
	api_v1 "k8s.io/api/core/v1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	"kube-eventer/util"
	"log"
	"os"

	"github.com/nlopes/slack"
)

var slackColors = map[string]string{
	"Normal":  "good",
	"Warning": "warning",
	"Danger":  "danger",
}

var slackErrMsg = `
%s
You need to set both slack token and channel for slack notify,
using "--token/-t" and "--channel/-c", or using environment variables:
export SLACK_TOKEN=slack_token
export SLACK_CHANNEL=slack_channel
Command line flags will override environment variables
`

var m = map[string]string{
	"created": "Normal",
	"deleted": "Danger",
	"updated": "Warning",
}

// Slack contains slack configuration
type Config struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
}

// Slack handler implements handler.Handler interface,
// Notify event to slack channel
type Slack struct {
	Token   string
	Channel string
}

type Event struct {
	Namespace string
	Kind      string
	Component string
	Host      string
	Reason    string
	Status    string
	Name      string
}

func (s *Slack) Init(c *Config) error {
	token := c.Token
	channel := c.Channel

	if token == "" {
		token = os.Getenv("SLACK_TOKEN")
	}

	if channel == "" {
		channel = os.Getenv("SLACK_CHANNEL")
	}

	s.Token = token
	s.Channel = channel

	return checkMissingSlackVars(s)
}

func checkMissingSlackVars(s *Slack) error {
	if s.Token == "" || s.Channel == "" {
		return fmt.Errorf(slackErrMsg, "Missing slack token or channel")
	}

	return nil
}

// New create new KubewatchEvent
func New(obj interface{}, action string) Event {
	var namespace, kind, component, host, reason, status, name string

	objectMeta := util.GetObjectMetaData(obj)
	namespace = objectMeta.Namespace
	name = objectMeta.Name
	reason = action
	status = m[action]

	switch object := obj.(type) {
	case *ext_v1beta1.DaemonSet:
		kind = "daemon set"
	case *apps_v1beta1.Deployment:
		kind = "deployment"
	case *batch_v1.Job:
		kind = "job"
	case *api_v1.Namespace:
		kind = "namespace"
	case *ext_v1beta1.Ingress:
		kind = "ingress"
	case *api_v1.PersistentVolume:
		kind = "persistent volume"
	case *api_v1.Pod:
		kind = "pod"
		host = object.Spec.NodeName
	case *api_v1.ReplicationController:
		kind = "replication controller"
	case *ext_v1beta1.ReplicaSet:
		kind = "replica set"
	case *api_v1.Service:
		kind = "service"
		component = string(object.Spec.Type)
	case *api_v1.Secret:
		kind = "secret"
	case *api_v1.ConfigMap:
		kind = "configmap"
	case Event:
		name = object.Name
		kind = object.Kind
		namespace = object.Namespace
	}

	kbEvent := Event{
		Namespace: namespace,
		Kind:      kind,
		Component: component,
		Host:      host,
		Reason:    reason,
		Status:    status,
		Name:      name,
	}
	return kbEvent
}

// ObjectCreated calls notifySlack on event creation
func (s *Slack) ObjectCreated(obj interface{}) {
	notifySlack(s, obj, "created")
}

// ObjectDeleted calls notifySlack on event creation
func (s *Slack) ObjectDeleted(obj interface{}) {
	notifySlack(s, obj, "deleted")
}

// ObjectUpdated calls notifySlack on event creation
func (s *Slack) ObjectUpdated(oldObj, newObj interface{}) {
	notifySlack(s, newObj, "updated")
}

// TestHandler tests the handler configurarion by sending test messages.
func (s *Slack) TestHandler() {
	api := slack.New(s.Token)
	attachment := slack.Attachment{
		Fields: []slack.AttachmentField{
			{
				Title: "kube-eventer",
				Value: "Testing Handler Configuration. This is a Test message.",
			},
		},
	}
	channelID, timestamp, err := api.PostMessage(s.Channel, "", slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func notifySlack(s *Slack, obj interface{}, action string) {
	e := New(obj, action)
	api := slack.New(s.Token)
	attachment := prepareSlackAttachment(e)
	channelID, timestamp, err := api.PostMessage(s.Channel, "", slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func prepareSlackAttachment(e Event) slack.Attachment {

	attachment := slack.Attachment{
		Fields: []slack.AttachmentField{
			{
				Title: "kube-eventer",
				Value: e.Message(),
			},
		},
	}

	if color, ok := slackColors[e.Status]; ok {
		attachment.Color = color
	}

	attachment.MarkdownIn = []string{"fields"}

	return attachment
}

// Message returns event message in standard format.
// included as a part of event packege to enhance code resuablity across handlers.
func (e *Event) Message() (msg string) {
	// using switch over if..else, since the format could vary based on the kind of the object in future.
	switch e.Kind {
	case "namespace":
		msg = fmt.Sprintf(
			"A namespace `%s` has been `%s`",
			e.Name,
			e.Reason,
		)
	default:
		msg = fmt.Sprintf(
			"A `%s` in namespace `%s` has been `%s`:\n`%s`",
			e.Kind,
			e.Namespace,
			e.Reason,
			e.Name,
		)
	}
	return msg
}
