package metrichub

const (
	// EventTimeLayout .000 固定给三位毫秒计数，.999 会去掉尾部的0
	EventTimeLayout = "20060102T150405.000-0700" // Example of timeStr: 20200304T000252.190+0800
	Product         = "k8s"
)

type SystemEvent struct {
	Product    string `json:"product"`
	EventType  string `json:"eventType"`
	Name       string `json:"name"`
	EventTime  string `json:"eventTime"`            // 本次报警时间，记录时间, 20211108T163717.313+0800
	GroupId    string `json:"groupId,omitempty"`    // 应用分组Id
	Resource   string `json:"resource,omitempty"`   //
	ResourceId string `json:"resourceId,omitempty"` // dimensions
	Level      string `json:"level"`                // CRITICAL\WARN\INFO，如果不知道填什么就填INFO
	Status     string `json:"status"`
	UserId     string `json:"userId,omitempty"` // userId
	Tags       string `json:"tags,omitempty"`   // metric=acs_ecs/cpu_usage,metric=acs_ecs/mem_usage
	Content    string `json:"content"`          // NewSystemEventContent
	RegionId   string `json:"regionId"`
	Time       string `json:"time,omitempty"`
}
