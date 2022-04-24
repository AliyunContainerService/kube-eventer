### webhook sink

*This sink supports generic webhook and you can use this sink integrated with chatbot(DingTalk,Slack,BearChat and so on) and webhook services.
To use the webhook sink add the following flag:

	--sink=webhook:<WEBHOOK_URL>&level=<Normal or Warning, Warning default>

The following options are available:
* `level` - Level of event (optional. default: Warning. Options: Warning and Normal)
* `namespaces` - Namespaces to filter (optional. default: all namespaces,use commas to separate multi namespaces, Regexp pattern support)
* `kinds` - Kinds to filter (optional. default: all kinds,use commas to separate multi kinds. Options: Node,Pod and so on.)
* `reason` - Reason to filter (optional. default: empty, Regexp pattern support). You can use multi reason fields in query.
* `method` - Method to send request (optional. default: GET)
* `header` - Header in request (optional. default: empty). You can use multi header field in query.
* `custom_body_configmap` - The configmap name of request body template. You can use Template to customize request body. (optional.)
* `custom_body_configmap_namespace` -  The configmap namespace of request body template. (optional.)

For example:

   	--sink=webhook:https://oapi.dingtalk.com/robot/send?access_token=a5c19f3e02feba7bd5dfc22bfb04afa212359acfe86fd80eb159187097b7d014&level=Normal&namespaces=a,b&kinds=c,d&header=contentType=customContentType&header=customHeaderKey=customHeaderValue 

### custom_body_configmap pattern 
The default request body template is below.     
```$xslt
{
	"EventType": "{{ .Type }}",
	"EventKind": "{{ .InvolvedObject.Kind }}",
	"EventReason": "{{ .Reason }}",
	"EventTime": "{{ .LastTimestamp }}",
	"EventMessage": "{{ .Message }}"
}
```
`kube-eventer` will render template with event to sink. The event struct is below.   
```$xslt
type Event struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// The object that this event is about.
	InvolvedObject ObjectReference `json:"involvedObject" protobuf:"bytes,2,opt,name=involvedObject"`

	// This should be a short, machine understandable string that gives the reason
	// for the transition into the object's current status.
	// TODO: provide exact specification for format.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// A human-readable description of the status of this operation.
	// TODO: decide on maximum length.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`

	// The component reporting this event. Should be a short machine understandable string.
	// +optional
	Source EventSource `json:"source,omitempty" protobuf:"bytes,5,opt,name=source"`

	// The time at which the event was first recorded. (Time of server receipt is in TypeMeta.)
	// +optional
	FirstTimestamp metav1.Time `json:"firstTimestamp,omitempty" protobuf:"bytes,6,opt,name=firstTimestamp"`

	// The time at which the most recent occurrence of this event was recorded.
	// +optional
	LastTimestamp metav1.Time `json:"lastTimestamp,omitempty" protobuf:"bytes,7,opt,name=lastTimestamp"`

	// The number of times this event has occurred.
	// +optional
	Count int32 `json:"count,omitempty" protobuf:"varint,8,opt,name=count"`

	// Type of this event (Normal, Warning), new types could be added in the future
	// +optional
	Type string `json:"type,omitempty" protobuf:"bytes,9,opt,name=type"`

	// Time when this Event was first observed.
	// +optional
	EventTime metav1.MicroTime `json:"eventTime,omitempty" protobuf:"bytes,10,opt,name=eventTime"`

	// Data about the Event series this event represents or nil if it's a singleton Event.
	// +optional
	Series *EventSeries `json:"series,omitempty" protobuf:"bytes,11,opt,name=series"`

	// What action was taken/failed regarding to the Regarding object.
	// +optional
	Action string `json:"action,omitempty" protobuf:"bytes,12,opt,name=action"`

	// Optional secondary object for more complex actions.
	// +optional
	Related *ObjectReference `json:"related,omitempty" protobuf:"bytes,13,opt,name=related"`

	// Name of the controller that emitted this Event, e.g. `kubernetes.io/kubelet`.
	// +optional
	ReportingController string `json:"reportingComponent" protobuf:"bytes,14,opt,name=reportingComponent"`

	// ID of the controller instance, e.g. `kubelet-xyzf`.
	// +optional
	ReportingInstance string `json:"reportingInstance" protobuf:"bytes,15,opt,name=reportingInstance"`
}
```
If you want to change the body struct with custom struct. You need to use `custom_body_configmap` and `custom_body_configmap_namespace`.    
The configMap must have a field called `content` and then put custom body template as value of `content`. For example.

```$xslt
apiVersion: v1
data:
  content: >-
    {"EventType": "{{ .Type }}","EventKind": "{{ .InvolvedObject.Kind }}","EventReason": "{{
    .Reason }}","EventTime": "{{ .LastTimestamp }}","EventMessage": "{{ .Message
    }}"}
kind: ConfigMap
metadata: 
  name: custom-webhook-body 
  namespace: kube-system 
```

### Typical Scenarios
#### Dingtalk 
Params 
```
--sink=webhook:https://oapi.dingtalk.com/robot/send?access_token=token&level=Normal&kinds=Pod&header=Content-Type=application/json&custom_body_configmap=custom-body&custom_body_configmap_namespace=kube-system&method=POST
```
configmap Body
```
{	"msgtype": "text",
	"text": {"content":"EventType:{{ .Type }}\nEventKind:{{ .InvolvedObject.Kind }}\nEventReason:{{ .Reason }}\nEventTime:{{ .LastTimestamp }}\nEventMessage:{{ .Message }}"},
	"markdown": {"title":"","text":""}
}
```
#### wechat 
Params 
```
--sink=webhook:http://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=633a31f6-7f9c-4bc4-97a0-0ec1eefa5898&level=Normal&kinds=Pod&header=Content-Type=application/json&custom_body_configmap=custom-body&custom_body_configmap_namespace=kube-system&method=POST
```
configmap Body 
```
{"msgtype": "text","text": {"content": "EventType:{{ .Type }}\nEventKind:{{ .InvolvedObject.Kind }}\nEventReason:{{ .Reason }}\nEventTime:{{ .LastTimestamp }}\nEventMessage:{{ .Message }}"}}
```
#### slack 
Params 
```
--sink=webhook:https://hooks.slack.com/services/d/B00000000/XXX?&level=Normal&kinds=Pod&header=Content-Type=application/json&custom_body_configmap=custom-body&custom_body_configmap_namespace=kube-system&method=POST
```
configmap Body 
```
{"channel": "testing",
"username": "Eventer",
"text":"EventType:{{ .Type }}\nEventKind:{{ .InvolvedObject.Kind }}\nEventReason:{{ .Reason }}\nEventTime:{{ .LastTimestamp }}\nEventMessage:{{ .Message }}"}
```

configmap example

```yaml
apiVersion: v1
data:
  content: '{
    "channel": "testing",
    "icon_emoji": ":k8s:",
    "username": "eventer",
    "attachments": [
        {
            "color": "warning",
            "text": "*Type*: `{{.Type}}`\n*Namespace*: `{{.InvolvedObject.Namespace}}`\n*Object*: `{{ .InvolvedObject.Kind }}/{{ .InvolvedObject.Name }}`\n*Reason*: `{{ .Reason }}`\n*Meaasge*: `{{ .Message }}`\n*Time*: `{{ .LastTimestamp }}`"
        }
    ]
  }'
kind: ConfigMap
metadata:
  name: custom-body
  namespace: kube-system
```

#### bear chat 
Params 
```
--sink=webhook:https://hook.bearychat.com/=bwIsS/incoming/xxxxxxxxxxxxxxxxxxxxxx?&level=Normal&kinds=Pod&header=Content-Type=application/json&custom_body_configmap=custom-body&custom_body_configmap_namespace=kube-system&method=POST
```
configmap Body 
```
"text":"EventType:{{ .Type }}\nEventKind:{{ .InvolvedObject.Kind }}\nEventReason:{{ .Reason }}\nEventTime:{{ .LastTimestamp }}\nEventMessage:{{ .Message }}"
```

#### feishu
Params
```
--sink=webhook:https://open.feishu.cn/open-apis/bot/hook/xxxxxxxxxxxxxxxxxxxxxxxxxx?level=Warning&method=POST&header=Content-Type=application/json&custom_body_configmap=custom-body&custom_body_configmap_namespace=kube-system
```
configmap example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-body
  namespace: kube-system
data:
  content: '{
   "title": "Kube-eventer",
   "text":  "EventType:  {{ .Type }}\nEventKind:  {{ .InvolvedObject.Kind }}\nEventReason:  {{ .Reason }}\nEventTime:  {{ .LastTimestamp }}\nEventMessage:  {{ .Message }}"
   }'

```

#### feishu v2
Params
```
--sink=webhook:https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxxxxxxxxxxx?level=Warning&method=POST&header=Content-Type=application/json&custom_body_configmap=custom-body&custom_body_configmap_namespace=kube-system
```

configmap example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-body
  namespace: kube-system
data:
  content: '{
   "msg_type": "interactive",
   "card": {
      "config": {
         "wide_screen_mode": true,
         "enable_forward": true
      },
      "header": {
         "title": {
            "tag": "plain_text",
            "content": "Kube-eventer"
         },
         "template": "Red"
      },
      "elements": [
         {
            "tag": "div",
            "text": {
               "tag": "lark_md",
               "content":  "**EventType:**  {{ .Type }}\n**EventKind:**  {{ .InvolvedObject.Kind }}\n**EventReason:**  {{ .Reason }}\n**EventTime:**  {{ .LastTimestamp }}\n**EventMessage:**  {{ .Message }}"
            }
        	}
      	]
   		}
		}'

```
