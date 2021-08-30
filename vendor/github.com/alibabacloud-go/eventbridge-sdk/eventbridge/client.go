// This file is auto-generated, don't edit it. Thanks.
/**
 *
 */
package eventbridge

import (
	eventbridgeutil "github.com/alibabacloud-go/eventbridge-util/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

/**
 * Model for initing client
 */
type Config struct {
	// accesskey id
	AccessKeyId *string `json:"accessKeyId,omitempty" xml:"accessKeyId,omitempty"`
	// accesskey secret
	AccessKeySecret *string `json:"accessKeySecret,omitempty" xml:"accessKeySecret,omitempty"`
	// security token
	SecurityToken *string `json:"securityToken,omitempty" xml:"securityToken,omitempty"`
	// http protocol
	Protocol *string `json:"protocol,omitempty" xml:"protocol,omitempty"`
	// region id
	RegionId *string `json:"regionId,omitempty" xml:"regionId,omitempty" pattern:"^[a-zA-Z0-9_-]+$"`
	// read timeout
	ReadTimeout *int `json:"readTimeout,omitempty" xml:"readTimeout,omitempty"`
	// connect timeout
	ConnectTimeout *int `json:"connectTimeout,omitempty" xml:"connectTimeout,omitempty"`
	// http proxy
	HttpProxy *string `json:"httpProxy,omitempty" xml:"httpProxy,omitempty"`
	// https proxy
	HttpsProxy *string `json:"httpsProxy,omitempty" xml:"httpsProxy,omitempty"`
	// credential
	Credential credential.Credential `json:"credential,omitempty" xml:"credential,omitempty"`
	// endpoint
	Endpoint *string `json:"endpoint,omitempty" xml:"endpoint,omitempty"`
	// proxy white list
	NoProxy *string `json:"noProxy,omitempty" xml:"noProxy,omitempty"`
	// max idle conns
	MaxIdleConns *int `json:"maxIdleConns,omitempty" xml:"maxIdleConns,omitempty"`
}

func (s Config) String() string {
	return tea.Prettify(s)
}

func (s Config) GoString() string {
	return s.String()
}

func (s *Config) SetAccessKeyId(v string) *Config {
	s.AccessKeyId = &v
	return s
}

func (s *Config) SetAccessKeySecret(v string) *Config {
	s.AccessKeySecret = &v
	return s
}

func (s *Config) SetSecurityToken(v string) *Config {
	s.SecurityToken = &v
	return s
}

func (s *Config) SetProtocol(v string) *Config {
	s.Protocol = &v
	return s
}

func (s *Config) SetRegionId(v string) *Config {
	s.RegionId = &v
	return s
}

func (s *Config) SetReadTimeout(v int) *Config {
	s.ReadTimeout = &v
	return s
}

func (s *Config) SetConnectTimeout(v int) *Config {
	s.ConnectTimeout = &v
	return s
}

func (s *Config) SetHttpProxy(v string) *Config {
	s.HttpProxy = &v
	return s
}

func (s *Config) SetHttpsProxy(v string) *Config {
	s.HttpsProxy = &v
	return s
}

func (s *Config) SetCredential(v credential.Credential) *Config {
	s.Credential = v
	return s
}

func (s *Config) SetEndpoint(v string) *Config {
	s.Endpoint = &v
	return s
}

func (s *Config) SetNoProxy(v string) *Config {
	s.NoProxy = &v
	return s
}

func (s *Config) SetMaxIdleConns(v int) *Config {
	s.MaxIdleConns = &v
	return s
}

/**
 * The detail of put event result
 */
type PutEventsResponseEntry struct {
	EventId      *string `json:"EventId,omitempty" xml:"EventId,omitempty"`
	TraceId      *string `json:"TraceId,omitempty" xml:"TraceId,omitempty"`
	ErrorCode    *string `json:"ErrorCode,omitempty" xml:"ErrorCode,omitempty"`
	ErrorMessage *string `json:"ErrorMessage,omitempty" xml:"ErrorMessage,omitempty"`
}

func (s PutEventsResponseEntry) String() string {
	return tea.Prettify(s)
}

func (s PutEventsResponseEntry) GoString() string {
	return s.String()
}

func (s *PutEventsResponseEntry) SetEventId(v string) *PutEventsResponseEntry {
	s.EventId = &v
	return s
}

func (s *PutEventsResponseEntry) SetTraceId(v string) *PutEventsResponseEntry {
	s.TraceId = &v
	return s
}

func (s *PutEventsResponseEntry) SetErrorCode(v string) *PutEventsResponseEntry {
	s.ErrorCode = &v
	return s
}

func (s *PutEventsResponseEntry) SetErrorMessage(v string) *PutEventsResponseEntry {
	s.ErrorMessage = &v
	return s
}

/**
 * Cloud Event Stanard Froamt
 */
type CloudEvent struct {
	Id              *string                `json:"id,omitempty" xml:"id,omitempty" require:"true"`
	Source          *string                `json:"source,omitempty" xml:"source,omitempty" require:"true" maxLength:"128"`
	Specversion     *string                `json:"specversion,omitempty" xml:"specversion,omitempty"`
	Type            *string                `json:"type,omitempty" xml:"type,omitempty" require:"true" maxLength:"64"`
	Datacontenttype *string                `json:"datacontenttype,omitempty" xml:"datacontenttype,omitempty"`
	Dataschema      *string                `json:"dataschema,omitempty" xml:"dataschema,omitempty"`
	Subject         *string                `json:"subject,omitempty" xml:"subject,omitempty"`
	Time            *string                `json:"time,omitempty" xml:"time,omitempty" maxLength:"64"`
	Extensions      map[string]interface{} `json:"extensions,omitempty" xml:"extensions,omitempty" require:"true"`
	Data            []byte                 `json:"data,omitempty" xml:"data,omitempty"`
}

func (s CloudEvent) String() string {
	return tea.Prettify(s)
}

func (s CloudEvent) GoString() string {
	return s.String()
}

func (s *CloudEvent) SetId(v string) *CloudEvent {
	s.Id = &v
	return s
}

func (s *CloudEvent) SetSource(v string) *CloudEvent {
	s.Source = &v
	return s
}

func (s *CloudEvent) SetSpecversion(v string) *CloudEvent {
	s.Specversion = &v
	return s
}

func (s *CloudEvent) SetType(v string) *CloudEvent {
	s.Type = &v
	return s
}

func (s *CloudEvent) SetDatacontenttype(v string) *CloudEvent {
	s.Datacontenttype = &v
	return s
}

func (s *CloudEvent) SetDataschema(v string) *CloudEvent {
	s.Dataschema = &v
	return s
}

func (s *CloudEvent) SetSubject(v string) *CloudEvent {
	s.Subject = &v
	return s
}

func (s *CloudEvent) SetTime(v string) *CloudEvent {
	s.Time = &v
	return s
}

func (s *CloudEvent) SetExtensions(v map[string]interface{}) *CloudEvent {
	s.Extensions = v
	return s
}

func (s *CloudEvent) SetData(v []byte) *CloudEvent {
	s.Data = v
	return s
}

/**
 * put event response
 */
type PutEventsResponse struct {
	RequestId              *string                   `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string                   `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	FailedEntryCount       *int                      `json:"FailedEntryCount,omitempty" xml:"FailedEntryCount,omitempty"`
	EntryList              []*PutEventsResponseEntry `json:"EntryList,omitempty" xml:"EntryList,omitempty" type:"Repeated"`
}

func (s PutEventsResponse) String() string {
	return tea.Prettify(s)
}

func (s PutEventsResponse) GoString() string {
	return s.String()
}

func (s *PutEventsResponse) SetRequestId(v string) *PutEventsResponse {
	s.RequestId = &v
	return s
}

func (s *PutEventsResponse) SetResourceOwnerAccountId(v string) *PutEventsResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *PutEventsResponse) SetFailedEntryCount(v int) *PutEventsResponse {
	s.FailedEntryCount = &v
	return s
}

func (s *PutEventsResponse) SetEntryList(v []*PutEventsResponseEntry) *PutEventsResponse {
	s.EntryList = v
	return s
}

/**
 * The request of create EventBus
 */
type CreateEventBusRequest struct {
	EventBusName *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true" maxLength:"127" minLength:"1"`
	Description  *string            `json:"Description,omitempty" xml:"Description,omitempty"`
	Tags         map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s CreateEventBusRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateEventBusRequest) GoString() string {
	return s.String()
}

func (s *CreateEventBusRequest) SetEventBusName(v string) *CreateEventBusRequest {
	s.EventBusName = &v
	return s
}

func (s *CreateEventBusRequest) SetDescription(v string) *CreateEventBusRequest {
	s.Description = &v
	return s
}

func (s *CreateEventBusRequest) SetTags(v map[string]*string) *CreateEventBusRequest {
	s.Tags = v
	return s
}

/**
 * The response of create EventBus
 */
type CreateEventBusResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	EventBusARN            *string `json:"EventBusARN,omitempty" xml:"EventBusARN,omitempty"`
}

func (s CreateEventBusResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateEventBusResponse) GoString() string {
	return s.String()
}

func (s *CreateEventBusResponse) SetRequestId(v string) *CreateEventBusResponse {
	s.RequestId = &v
	return s
}

func (s *CreateEventBusResponse) SetResourceOwnerAccountId(v string) *CreateEventBusResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *CreateEventBusResponse) SetEventBusARN(v string) *CreateEventBusResponse {
	s.EventBusARN = &v
	return s
}

/**
 * The request of delete the EventBus
 */
type DeleteEventBusRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
}

func (s DeleteEventBusRequest) String() string {
	return tea.Prettify(s)
}

func (s DeleteEventBusRequest) GoString() string {
	return s.String()
}

func (s *DeleteEventBusRequest) SetEventBusName(v string) *DeleteEventBusRequest {
	s.EventBusName = &v
	return s
}

/**
 * The response of delete the EventBus
 */
type DeleteEventBusResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
}

func (s DeleteEventBusResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteEventBusResponse) GoString() string {
	return s.String()
}

func (s *DeleteEventBusResponse) SetRequestId(v string) *DeleteEventBusResponse {
	s.RequestId = &v
	return s
}

func (s *DeleteEventBusResponse) SetResourceOwnerAccountId(v string) *DeleteEventBusResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

/**
 * The request of get the detail of EventBus
 */
type GetEventBusRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
}

func (s GetEventBusRequest) String() string {
	return tea.Prettify(s)
}

func (s GetEventBusRequest) GoString() string {
	return s.String()
}

func (s *GetEventBusRequest) SetEventBusName(v string) *GetEventBusRequest {
	s.EventBusName = &v
	return s
}

/**
 * The response of get the detail of EventBus
 */
type GetEventBusResponse struct {
	RequestId              *string            `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string            `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	EventBusARN            *string            `json:"EventBusARN,omitempty" xml:"EventBusARN,omitempty" require:"true"`
	EventBusName           *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	Description            *string            `json:"Description,omitempty" xml:"Description,omitempty" require:"true"`
	CreateTimestamp        *int64             `json:"CreateTimestamp,omitempty" xml:"CreateTimestamp,omitempty" require:"true"`
	Tags                   map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s GetEventBusResponse) String() string {
	return tea.Prettify(s)
}

func (s GetEventBusResponse) GoString() string {
	return s.String()
}

func (s *GetEventBusResponse) SetRequestId(v string) *GetEventBusResponse {
	s.RequestId = &v
	return s
}

func (s *GetEventBusResponse) SetResourceOwnerAccountId(v string) *GetEventBusResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *GetEventBusResponse) SetEventBusARN(v string) *GetEventBusResponse {
	s.EventBusARN = &v
	return s
}

func (s *GetEventBusResponse) SetEventBusName(v string) *GetEventBusResponse {
	s.EventBusName = &v
	return s
}

func (s *GetEventBusResponse) SetDescription(v string) *GetEventBusResponse {
	s.Description = &v
	return s
}

func (s *GetEventBusResponse) SetCreateTimestamp(v int64) *GetEventBusResponse {
	s.CreateTimestamp = &v
	return s
}

func (s *GetEventBusResponse) SetTags(v map[string]*string) *GetEventBusResponse {
	s.Tags = v
	return s
}

/**
 * The request of list all the EventBus which meet the search criteria
 */
type ListEventBusesRequest struct {
	NamePrefix *string `json:"NamePrefix,omitempty" xml:"NamePrefix,omitempty"`
	Limit      *int    `json:"Limit,omitempty" xml:"Limit,omitempty"`
	NextToken  *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
}

func (s ListEventBusesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListEventBusesRequest) GoString() string {
	return s.String()
}

func (s *ListEventBusesRequest) SetNamePrefix(v string) *ListEventBusesRequest {
	s.NamePrefix = &v
	return s
}

func (s *ListEventBusesRequest) SetLimit(v int) *ListEventBusesRequest {
	s.Limit = &v
	return s
}

func (s *ListEventBusesRequest) SetNextToken(v string) *ListEventBusesRequest {
	s.NextToken = &v
	return s
}

/**
 * The detail of EventBuses
 */
type EventBusEntry struct {
	EventBusName    *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	EventBusARN     *string            `json:"EventBusARN,omitempty" xml:"EventBusARN,omitempty" require:"true"`
	Description     *string            `json:"Description,omitempty" xml:"Description,omitempty" require:"true"`
	CreateTimestamp *int64             `json:"CreateTimestamp,omitempty" xml:"CreateTimestamp,omitempty" require:"true"`
	Tags            map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s EventBusEntry) String() string {
	return tea.Prettify(s)
}

func (s EventBusEntry) GoString() string {
	return s.String()
}

func (s *EventBusEntry) SetEventBusName(v string) *EventBusEntry {
	s.EventBusName = &v
	return s
}

func (s *EventBusEntry) SetEventBusARN(v string) *EventBusEntry {
	s.EventBusARN = &v
	return s
}

func (s *EventBusEntry) SetDescription(v string) *EventBusEntry {
	s.Description = &v
	return s
}

func (s *EventBusEntry) SetCreateTimestamp(v int64) *EventBusEntry {
	s.CreateTimestamp = &v
	return s
}

func (s *EventBusEntry) SetTags(v map[string]*string) *EventBusEntry {
	s.Tags = v
	return s
}

/**
 * The response of search EventBus
 */
type ListEventBusesResponse struct {
	RequestId              *string          `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string          `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	EventBuses             []*EventBusEntry `json:"EventBuses,omitempty" xml:"EventBuses,omitempty" require:"true" type:"Repeated"`
	NextToken              *string          `json:"NextToken,omitempty" xml:"NextToken,omitempty" require:"true"`
	Total                  *int             `json:"Total,omitempty" xml:"Total,omitempty" require:"true"`
}

func (s ListEventBusesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListEventBusesResponse) GoString() string {
	return s.String()
}

func (s *ListEventBusesResponse) SetRequestId(v string) *ListEventBusesResponse {
	s.RequestId = &v
	return s
}

func (s *ListEventBusesResponse) SetResourceOwnerAccountId(v string) *ListEventBusesResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *ListEventBusesResponse) SetEventBuses(v []*EventBusEntry) *ListEventBusesResponse {
	s.EventBuses = v
	return s
}

func (s *ListEventBusesResponse) SetNextToken(v string) *ListEventBusesResponse {
	s.NextToken = &v
	return s
}

func (s *ListEventBusesResponse) SetTotal(v int) *ListEventBusesResponse {
	s.Total = &v
	return s
}

/**
 * The request of create an EventBus rule on Aliyun
 */
type CreateRuleRequest struct {
	EventBusName  *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true" maxLength:"127" minLength:"1"`
	Description   *string            `json:"Description,omitempty" xml:"Description,omitempty"`
	RuleName      *string            `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	Status        *string            `json:"Status,omitempty" xml:"Status,omitempty"`
	FilterPattern *string            `json:"FilterPattern,omitempty" xml:"FilterPattern,omitempty"`
	Targets       []*TargetEntry     `json:"Targets,omitempty" xml:"Targets,omitempty" require:"true" type:"Repeated"`
	Tags          map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s CreateRuleRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateRuleRequest) GoString() string {
	return s.String()
}

func (s *CreateRuleRequest) SetEventBusName(v string) *CreateRuleRequest {
	s.EventBusName = &v
	return s
}

func (s *CreateRuleRequest) SetDescription(v string) *CreateRuleRequest {
	s.Description = &v
	return s
}

func (s *CreateRuleRequest) SetRuleName(v string) *CreateRuleRequest {
	s.RuleName = &v
	return s
}

func (s *CreateRuleRequest) SetStatus(v string) *CreateRuleRequest {
	s.Status = &v
	return s
}

func (s *CreateRuleRequest) SetFilterPattern(v string) *CreateRuleRequest {
	s.FilterPattern = &v
	return s
}

func (s *CreateRuleRequest) SetTargets(v []*TargetEntry) *CreateRuleRequest {
	s.Targets = v
	return s
}

func (s *CreateRuleRequest) SetTags(v map[string]*string) *CreateRuleRequest {
	s.Tags = v
	return s
}

/**
 * The detail of TargetEntry
 */
type TargetEntry struct {
	Id                *string          `json:"Id,omitempty" xml:"Id,omitempty" require:"true"`
	Type              *string          `json:"Type,omitempty" xml:"Type,omitempty" require:"true"`
	Endpoint          *string          `json:"Endpoint,omitempty" xml:"Endpoint,omitempty" require:"true"`
	PushRetryStrategy *string          `json:"PushRetryStrategy,omitempty" xml:"PushRetryStrategy,omitempty"`
	ParamList         []*EBTargetParam `json:"ParamList,omitempty" xml:"ParamList,omitempty" type:"Repeated"`
}

func (s TargetEntry) String() string {
	return tea.Prettify(s)
}

func (s TargetEntry) GoString() string {
	return s.String()
}

func (s *TargetEntry) SetId(v string) *TargetEntry {
	s.Id = &v
	return s
}

func (s *TargetEntry) SetType(v string) *TargetEntry {
	s.Type = &v
	return s
}

func (s *TargetEntry) SetEndpoint(v string) *TargetEntry {
	s.Endpoint = &v
	return s
}

func (s *TargetEntry) SetPushRetryStrategy(v string) *TargetEntry {
	s.PushRetryStrategy = &v
	return s
}

func (s *TargetEntry) SetParamList(v []*EBTargetParam) *TargetEntry {
	s.ParamList = v
	return s
}

/**
 * The param of EBTargetParam
 */
type EBTargetParam struct {
	ResourceKey *string `json:"ResourceKey,omitempty" xml:"ResourceKey,omitempty" require:"true"`
	Form        *string `json:"Form,omitempty" xml:"Form,omitempty" require:"true"`
	Value       *string `json:"Value,omitempty" xml:"Value,omitempty"`
	Template    *string `json:"Template,omitempty" xml:"Template,omitempty"`
}

func (s EBTargetParam) String() string {
	return tea.Prettify(s)
}

func (s EBTargetParam) GoString() string {
	return s.String()
}

func (s *EBTargetParam) SetResourceKey(v string) *EBTargetParam {
	s.ResourceKey = &v
	return s
}

func (s *EBTargetParam) SetForm(v string) *EBTargetParam {
	s.Form = &v
	return s
}

func (s *EBTargetParam) SetValue(v string) *EBTargetParam {
	s.Value = &v
	return s
}

func (s *EBTargetParam) SetTemplate(v string) *EBTargetParam {
	s.Template = &v
	return s
}

/**
 * The response of create EventBus Rule
 */
type CreateRuleResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	RuleARN                *string `json:"RuleARN,omitempty" xml:"RuleARN,omitempty" require:"true"`
}

func (s CreateRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateRuleResponse) GoString() string {
	return s.String()
}

func (s *CreateRuleResponse) SetRequestId(v string) *CreateRuleResponse {
	s.RequestId = &v
	return s
}

func (s *CreateRuleResponse) SetResourceOwnerAccountId(v string) *CreateRuleResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *CreateRuleResponse) SetRuleARN(v string) *CreateRuleResponse {
	s.RuleARN = &v
	return s
}

/**
 * The request of delete the rule
 */
type DeleteRuleRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
}

func (s DeleteRuleRequest) String() string {
	return tea.Prettify(s)
}

func (s DeleteRuleRequest) GoString() string {
	return s.String()
}

func (s *DeleteRuleRequest) SetEventBusName(v string) *DeleteRuleRequest {
	s.EventBusName = &v
	return s
}

func (s *DeleteRuleRequest) SetRuleName(v string) *DeleteRuleRequest {
	s.RuleName = &v
	return s
}

/**
 * The response of delete the rule
 */
type DeleteRuleResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
}

func (s DeleteRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteRuleResponse) GoString() string {
	return s.String()
}

func (s *DeleteRuleResponse) SetRequestId(v string) *DeleteRuleResponse {
	s.RequestId = &v
	return s
}

func (s *DeleteRuleResponse) SetResourceOwnerAccountId(v string) *DeleteRuleResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

/**
 * The request of disable the EventBus rule
 */
type DisableRuleRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
}

func (s DisableRuleRequest) String() string {
	return tea.Prettify(s)
}

func (s DisableRuleRequest) GoString() string {
	return s.String()
}

func (s *DisableRuleRequest) SetEventBusName(v string) *DisableRuleRequest {
	s.EventBusName = &v
	return s
}

func (s *DisableRuleRequest) SetRuleName(v string) *DisableRuleRequest {
	s.RuleName = &v
	return s
}

/**
 * The response of disable the EventBus rule
 */
type DisableRuleResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
}

func (s DisableRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s DisableRuleResponse) GoString() string {
	return s.String()
}

func (s *DisableRuleResponse) SetRequestId(v string) *DisableRuleResponse {
	s.RequestId = &v
	return s
}

func (s *DisableRuleResponse) SetResourceOwnerAccountId(v string) *DisableRuleResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

/**
 * The request of enable the EventBus rule
 */
type EnableRuleRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
}

func (s EnableRuleRequest) String() string {
	return tea.Prettify(s)
}

func (s EnableRuleRequest) GoString() string {
	return s.String()
}

func (s *EnableRuleRequest) SetEventBusName(v string) *EnableRuleRequest {
	s.EventBusName = &v
	return s
}

func (s *EnableRuleRequest) SetRuleName(v string) *EnableRuleRequest {
	s.RuleName = &v
	return s
}

/**
 * The response of enable the EventBus rule
 */
type EnableRuleResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
}

func (s EnableRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s EnableRuleResponse) GoString() string {
	return s.String()
}

func (s *EnableRuleResponse) SetRequestId(v string) *EnableRuleResponse {
	s.RequestId = &v
	return s
}

func (s *EnableRuleResponse) SetResourceOwnerAccountId(v string) *EnableRuleResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

/**
 * The request of Get EventBus
 */
type GetRuleRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
}

func (s GetRuleRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRuleRequest) GoString() string {
	return s.String()
}

func (s *GetRuleRequest) SetEventBusName(v string) *GetRuleRequest {
	s.EventBusName = &v
	return s
}

func (s *GetRuleRequest) SetRuleName(v string) *GetRuleRequest {
	s.RuleName = &v
	return s
}

/**
 * The response of Get EventBus
 */
type GetRuleResponse struct {
	RequestId              *string            `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string            `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	EventBusName           *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleARN                *string            `json:"RuleARN,omitempty" xml:"RuleARN,omitempty" require:"true"`
	RuleName               *string            `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	Description            *string            `json:"Description,omitempty" xml:"Description,omitempty" require:"true"`
	Status                 *string            `json:"Status,omitempty" xml:"Status,omitempty" require:"true"`
	FilterPattern          *string            `json:"FilterPattern,omitempty" xml:"FilterPattern,omitempty" require:"true"`
	Targets                []*TargetEntry     `json:"Targets,omitempty" xml:"Targets,omitempty" require:"true" type:"Repeated"`
	Ctime                  *int64             `json:"Ctime,omitempty" xml:"Ctime,omitempty" require:"true"`
	Mtime                  *int64             `json:"Mtime,omitempty" xml:"Mtime,omitempty" require:"true"`
	Tags                   map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s GetRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRuleResponse) GoString() string {
	return s.String()
}

func (s *GetRuleResponse) SetRequestId(v string) *GetRuleResponse {
	s.RequestId = &v
	return s
}

func (s *GetRuleResponse) SetResourceOwnerAccountId(v string) *GetRuleResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *GetRuleResponse) SetEventBusName(v string) *GetRuleResponse {
	s.EventBusName = &v
	return s
}

func (s *GetRuleResponse) SetRuleARN(v string) *GetRuleResponse {
	s.RuleARN = &v
	return s
}

func (s *GetRuleResponse) SetRuleName(v string) *GetRuleResponse {
	s.RuleName = &v
	return s
}

func (s *GetRuleResponse) SetDescription(v string) *GetRuleResponse {
	s.Description = &v
	return s
}

func (s *GetRuleResponse) SetStatus(v string) *GetRuleResponse {
	s.Status = &v
	return s
}

func (s *GetRuleResponse) SetFilterPattern(v string) *GetRuleResponse {
	s.FilterPattern = &v
	return s
}

func (s *GetRuleResponse) SetTargets(v []*TargetEntry) *GetRuleResponse {
	s.Targets = v
	return s
}

func (s *GetRuleResponse) SetCtime(v int64) *GetRuleResponse {
	s.Ctime = &v
	return s
}

func (s *GetRuleResponse) SetMtime(v int64) *GetRuleResponse {
	s.Mtime = &v
	return s
}

func (s *GetRuleResponse) SetTags(v map[string]*string) *GetRuleResponse {
	s.Tags = v
	return s
}

/**
 * The request of search EventBus
 */
type ListRulesRequest struct {
	EventBusName   *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleNamePrefix *string `json:"RuleNamePrefix,omitempty" xml:"RuleNamePrefix,omitempty"`
	Limit          *int    `json:"Limit,omitempty" xml:"Limit,omitempty"`
	NextToken      *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
}

func (s ListRulesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListRulesRequest) GoString() string {
	return s.String()
}

func (s *ListRulesRequest) SetEventBusName(v string) *ListRulesRequest {
	s.EventBusName = &v
	return s
}

func (s *ListRulesRequest) SetRuleNamePrefix(v string) *ListRulesRequest {
	s.RuleNamePrefix = &v
	return s
}

func (s *ListRulesRequest) SetLimit(v int) *ListRulesRequest {
	s.Limit = &v
	return s
}

func (s *ListRulesRequest) SetNextToken(v string) *ListRulesRequest {
	s.NextToken = &v
	return s
}

/**
 * The response of search EventBus
 */
type ListRulesResponse struct {
	RequestId              *string         `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string         `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	NextToken              *string         `json:"NextToken,omitempty" xml:"NextToken,omitempty" require:"true"`
	Rules                  []*EventRuleDTO `json:"Rules,omitempty" xml:"Rules,omitempty" require:"true" type:"Repeated"`
	Total                  *int            `json:"Total,omitempty" xml:"Total,omitempty" require:"true"`
}

func (s ListRulesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListRulesResponse) GoString() string {
	return s.String()
}

func (s *ListRulesResponse) SetRequestId(v string) *ListRulesResponse {
	s.RequestId = &v
	return s
}

func (s *ListRulesResponse) SetResourceOwnerAccountId(v string) *ListRulesResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *ListRulesResponse) SetNextToken(v string) *ListRulesResponse {
	s.NextToken = &v
	return s
}

func (s *ListRulesResponse) SetRules(v []*EventRuleDTO) *ListRulesResponse {
	s.Rules = v
	return s
}

func (s *ListRulesResponse) SetTotal(v int) *ListRulesResponse {
	s.Total = &v
	return s
}

/**
 * The detail of EventBuses rule
 */
type EventRuleDTO struct {
	RuleARN       *string            `json:"RuleARN,omitempty" xml:"RuleARN,omitempty" require:"true"`
	EventBusName  *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName      *string            `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	Description   *string            `json:"Description,omitempty" xml:"Description,omitempty" require:"true"`
	Status        *string            `json:"Status,omitempty" xml:"Status,omitempty" require:"true"`
	FilterPattern *string            `json:"FilterPattern,omitempty" xml:"FilterPattern,omitempty" require:"true"`
	Targets       []*TargetEntry     `json:"Targets,omitempty" xml:"Targets,omitempty" require:"true" type:"Repeated"`
	Ctime         *int64             `json:"Ctime,omitempty" xml:"Ctime,omitempty" require:"true"`
	Mtime         *int64             `json:"Mtime,omitempty" xml:"Mtime,omitempty" require:"true"`
	Tags          map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s EventRuleDTO) String() string {
	return tea.Prettify(s)
}

func (s EventRuleDTO) GoString() string {
	return s.String()
}

func (s *EventRuleDTO) SetRuleARN(v string) *EventRuleDTO {
	s.RuleARN = &v
	return s
}

func (s *EventRuleDTO) SetEventBusName(v string) *EventRuleDTO {
	s.EventBusName = &v
	return s
}

func (s *EventRuleDTO) SetRuleName(v string) *EventRuleDTO {
	s.RuleName = &v
	return s
}

func (s *EventRuleDTO) SetDescription(v string) *EventRuleDTO {
	s.Description = &v
	return s
}

func (s *EventRuleDTO) SetStatus(v string) *EventRuleDTO {
	s.Status = &v
	return s
}

func (s *EventRuleDTO) SetFilterPattern(v string) *EventRuleDTO {
	s.FilterPattern = &v
	return s
}

func (s *EventRuleDTO) SetTargets(v []*TargetEntry) *EventRuleDTO {
	s.Targets = v
	return s
}

func (s *EventRuleDTO) SetCtime(v int64) *EventRuleDTO {
	s.Ctime = &v
	return s
}

func (s *EventRuleDTO) SetMtime(v int64) *EventRuleDTO {
	s.Mtime = &v
	return s
}

func (s *EventRuleDTO) SetTags(v map[string]*string) *EventRuleDTO {
	s.Tags = v
	return s
}

/**
 * The request of update the EventBus rule
 */
type UpdateRuleRequest struct {
	EventBusName  *string            `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName      *string            `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	Description   *string            `json:"Description,omitempty" xml:"Description,omitempty"`
	Status        *string            `json:"Status,omitempty" xml:"Status,omitempty"`
	FilterPattern *string            `json:"FilterPattern,omitempty" xml:"FilterPattern,omitempty"`
	Tags          map[string]*string `json:"Tags,omitempty" xml:"Tags,omitempty"`
}

func (s UpdateRuleRequest) String() string {
	return tea.Prettify(s)
}

func (s UpdateRuleRequest) GoString() string {
	return s.String()
}

func (s *UpdateRuleRequest) SetEventBusName(v string) *UpdateRuleRequest {
	s.EventBusName = &v
	return s
}

func (s *UpdateRuleRequest) SetRuleName(v string) *UpdateRuleRequest {
	s.RuleName = &v
	return s
}

func (s *UpdateRuleRequest) SetDescription(v string) *UpdateRuleRequest {
	s.Description = &v
	return s
}

func (s *UpdateRuleRequest) SetStatus(v string) *UpdateRuleRequest {
	s.Status = &v
	return s
}

func (s *UpdateRuleRequest) SetFilterPattern(v string) *UpdateRuleRequest {
	s.FilterPattern = &v
	return s
}

func (s *UpdateRuleRequest) SetTags(v map[string]*string) *UpdateRuleRequest {
	s.Tags = v
	return s
}

/**
 * The response of update the EventBus rule
 */
type UpdateRuleResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
}

func (s UpdateRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateRuleResponse) GoString() string {
	return s.String()
}

func (s *UpdateRuleResponse) SetRequestId(v string) *UpdateRuleResponse {
	s.RequestId = &v
	return s
}

func (s *UpdateRuleResponse) SetResourceOwnerAccountId(v string) *UpdateRuleResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

/**
 * The request of create Targets
 */
type CreateTargetsRequest struct {
	EventBusName *string        `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string        `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	Targets      []*TargetEntry `json:"Targets,omitempty" xml:"Targets,omitempty" require:"true" type:"Repeated"`
}

func (s CreateTargetsRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateTargetsRequest) GoString() string {
	return s.String()
}

func (s *CreateTargetsRequest) SetEventBusName(v string) *CreateTargetsRequest {
	s.EventBusName = &v
	return s
}

func (s *CreateTargetsRequest) SetRuleName(v string) *CreateTargetsRequest {
	s.RuleName = &v
	return s
}

func (s *CreateTargetsRequest) SetTargets(v []*TargetEntry) *CreateTargetsRequest {
	s.Targets = v
	return s
}

/**
 * The response of create Targets
 */
type CreateTargetsResponse struct {
	RequestId              *string              `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string              `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	ErrorEntriesCount      *int                 `json:"ErrorEntriesCount,omitempty" xml:"ErrorEntriesCount,omitempty" require:"true"`
	ErrorEntries           []*TargetResultEntry `json:"ErrorEntries,omitempty" xml:"ErrorEntries,omitempty" require:"true" type:"Repeated"`
}

func (s CreateTargetsResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateTargetsResponse) GoString() string {
	return s.String()
}

func (s *CreateTargetsResponse) SetRequestId(v string) *CreateTargetsResponse {
	s.RequestId = &v
	return s
}

func (s *CreateTargetsResponse) SetResourceOwnerAccountId(v string) *CreateTargetsResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *CreateTargetsResponse) SetErrorEntriesCount(v int) *CreateTargetsResponse {
	s.ErrorEntriesCount = &v
	return s
}

func (s *CreateTargetsResponse) SetErrorEntries(v []*TargetResultEntry) *CreateTargetsResponse {
	s.ErrorEntries = v
	return s
}

/**
 * The request of delete Targets
 */
type DeleteTargetsRequest struct {
	EventBusName *string   `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string   `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	TargetIds    []*string `json:"TargetIds,omitempty" xml:"TargetIds,omitempty" require:"true" type:"Repeated"`
}

func (s DeleteTargetsRequest) String() string {
	return tea.Prettify(s)
}

func (s DeleteTargetsRequest) GoString() string {
	return s.String()
}

func (s *DeleteTargetsRequest) SetEventBusName(v string) *DeleteTargetsRequest {
	s.EventBusName = &v
	return s
}

func (s *DeleteTargetsRequest) SetRuleName(v string) *DeleteTargetsRequest {
	s.RuleName = &v
	return s
}

func (s *DeleteTargetsRequest) SetTargetIds(v []*string) *DeleteTargetsRequest {
	s.TargetIds = v
	return s
}

/**
 * The response of delete Targets
 */
type DeleteTargetsResponse struct {
	RequestId              *string              `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string              `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	ErrorEntriesCount      *int                 `json:"ErrorEntriesCount,omitempty" xml:"ErrorEntriesCount,omitempty" require:"true"`
	ErrorEntries           []*TargetResultEntry `json:"ErrorEntries,omitempty" xml:"ErrorEntries,omitempty" require:"true" type:"Repeated"`
}

func (s DeleteTargetsResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteTargetsResponse) GoString() string {
	return s.String()
}

func (s *DeleteTargetsResponse) SetRequestId(v string) *DeleteTargetsResponse {
	s.RequestId = &v
	return s
}

func (s *DeleteTargetsResponse) SetResourceOwnerAccountId(v string) *DeleteTargetsResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *DeleteTargetsResponse) SetErrorEntriesCount(v int) *DeleteTargetsResponse {
	s.ErrorEntriesCount = &v
	return s
}

func (s *DeleteTargetsResponse) SetErrorEntries(v []*TargetResultEntry) *DeleteTargetsResponse {
	s.ErrorEntries = v
	return s
}

/**
 * The result detail of target operation
 */
type TargetResultEntry struct {
	ErrorCode    *string `json:"ErrorCode,omitempty" xml:"ErrorCode,omitempty" require:"true"`
	ErrorMessage *string `json:"ErrorMessage,omitempty" xml:"ErrorMessage,omitempty" require:"true"`
	EntryId      *string `json:"EntryId,omitempty" xml:"EntryId,omitempty" require:"true"`
}

func (s TargetResultEntry) String() string {
	return tea.Prettify(s)
}

func (s TargetResultEntry) GoString() string {
	return s.String()
}

func (s *TargetResultEntry) SetErrorCode(v string) *TargetResultEntry {
	s.ErrorCode = &v
	return s
}

func (s *TargetResultEntry) SetErrorMessage(v string) *TargetResultEntry {
	s.ErrorMessage = &v
	return s
}

func (s *TargetResultEntry) SetEntryId(v string) *TargetResultEntry {
	s.EntryId = &v
	return s
}

/**
 * The request of search Targets
 */
type ListTargetsRequest struct {
	EventBusName *string `json:"EventBusName,omitempty" xml:"EventBusName,omitempty" require:"true"`
	RuleName     *string `json:"RuleName,omitempty" xml:"RuleName,omitempty" require:"true"`
	Limit        *int    `json:"Limit,omitempty" xml:"Limit,omitempty"`
}

func (s ListTargetsRequest) String() string {
	return tea.Prettify(s)
}

func (s ListTargetsRequest) GoString() string {
	return s.String()
}

func (s *ListTargetsRequest) SetEventBusName(v string) *ListTargetsRequest {
	s.EventBusName = &v
	return s
}

func (s *ListTargetsRequest) SetRuleName(v string) *ListTargetsRequest {
	s.RuleName = &v
	return s
}

func (s *ListTargetsRequest) SetLimit(v int) *ListTargetsRequest {
	s.Limit = &v
	return s
}

/**
 * The response of search Targets
 */
type ListTargetsResponse struct {
	RequestId              *string        `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string        `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	Targets                []*TargetEntry `json:"Targets,omitempty" xml:"Targets,omitempty" require:"true" type:"Repeated"`
}

func (s ListTargetsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListTargetsResponse) GoString() string {
	return s.String()
}

func (s *ListTargetsResponse) SetRequestId(v string) *ListTargetsResponse {
	s.RequestId = &v
	return s
}

func (s *ListTargetsResponse) SetResourceOwnerAccountId(v string) *ListTargetsResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *ListTargetsResponse) SetTargets(v []*TargetEntry) *ListTargetsResponse {
	s.Targets = v
	return s
}

/**
 * The request of testEventPattern
 */
type TestEventPatternRequest struct {
	Event        *string `json:"Event,omitempty" xml:"Event,omitempty" require:"true" maxLength:"2048"`
	EventPattern *string `json:"EventPattern,omitempty" xml:"EventPattern,omitempty" require:"true" maxLength:"2048"`
}

func (s TestEventPatternRequest) String() string {
	return tea.Prettify(s)
}

func (s TestEventPatternRequest) GoString() string {
	return s.String()
}

func (s *TestEventPatternRequest) SetEvent(v string) *TestEventPatternRequest {
	s.Event = &v
	return s
}

func (s *TestEventPatternRequest) SetEventPattern(v string) *TestEventPatternRequest {
	s.EventPattern = &v
	return s
}

/**
 * The response of testEventPattern
 */
type TestEventPatternResponse struct {
	RequestId              *string `json:"RequestId,omitempty" xml:"RequestId,omitempty" require:"true"`
	ResourceOwnerAccountId *string `json:"ResourceOwnerAccountId,omitempty" xml:"ResourceOwnerAccountId,omitempty" require:"true"`
	Result                 *bool   `json:"Result,omitempty" xml:"Result,omitempty" require:"true"`
}

func (s TestEventPatternResponse) String() string {
	return tea.Prettify(s)
}

func (s TestEventPatternResponse) GoString() string {
	return s.String()
}

func (s *TestEventPatternResponse) SetRequestId(v string) *TestEventPatternResponse {
	s.RequestId = &v
	return s
}

func (s *TestEventPatternResponse) SetResourceOwnerAccountId(v string) *TestEventPatternResponse {
	s.ResourceOwnerAccountId = &v
	return s
}

func (s *TestEventPatternResponse) SetResult(v bool) *TestEventPatternResponse {
	s.Result = &v
	return s
}

type Client struct {
	Protocol       *string
	ReadTimeout    *int
	ConnectTimeout *int
	HttpProxy      *string
	HttpsProxy     *string
	NoProxy        *string
	MaxIdleConns   *int
	Endpoint       *string
	RegionId       *string
	Credential     credential.Credential
}

/**
 * Init client with Config
 * @param config config contains the necessary information to create a client
 */
func NewClient(config *Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *Config) (_err error) {
	if tea.BoolValue(util.IsUnset(tea.ToMap(config))) {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterMissing",
			"message": "'config' can not be unset",
		})
		return _err
	}

	_err = util.ValidateModel(config)
	if _err != nil {
		return _err
	}
	if !tea.BoolValue(util.Empty(config.AccessKeyId)) && !tea.BoolValue(util.Empty(config.AccessKeySecret)) {
		credentialType := tea.String("access_key")
		if !tea.BoolValue(util.Empty(config.SecurityToken)) {
			credentialType = tea.String("sts")
		}

		credentialConfig := &credential.Config{
			AccessKeyId:     config.AccessKeyId,
			Type:            credentialType,
			AccessKeySecret: config.AccessKeySecret,
			SecurityToken:   config.SecurityToken,
		}
		client.Credential, _err = credential.NewCredential(credentialConfig)
		if _err != nil {
			return _err
		}

	} else if !tea.BoolValue(util.IsUnset(config.Credential)) {
		client.Credential = config.Credential
	} else {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterMissing",
			"message": "'accessKeyId' and 'accessKeySecret' or 'credential' can not be unset",
		})
		return _err
	}

	if tea.BoolValue(util.Empty(config.Endpoint)) {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterMissing",
			"message": "'endpoint' can not be unset",
		})
		return _err
	}

	if tea.BoolValue(eventbridgeutil.StartWith(config.Endpoint, tea.String("http"))) || tea.BoolValue(eventbridgeutil.StartWith(config.Endpoint, tea.String("https"))) {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterError",
			"message": "'endpoint' shouldn't start with 'http' or 'https'",
		})
		return _err
	}

	client.RegionId = config.RegionId
	client.Protocol = config.Protocol
	client.Endpoint = config.Endpoint
	client.ReadTimeout = config.ReadTimeout
	client.ConnectTimeout = config.ConnectTimeout
	client.HttpProxy = config.HttpProxy
	client.HttpsProxy = config.HttpsProxy
	client.MaxIdleConns = config.MaxIdleConns
	return nil
}

/**
 * Encapsulate the request and invoke the network
 * @param action the api name
 * @param protocol http or https
 * @param method e.g. GET
 * @param pathname pathname of every api
 * @param query which contains request params
 * @param body content of request
 * @param runtime which controls some details of call api, such as retry times
 * @return the response
 */
func (client *Client) DoRequest(action *string, protocol *string, method *string, pathname *string, query map[string]*string, body interface{}, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(util.DefaultNumber(runtime.ReadTimeout, client.ReadTimeout)),
		"connectTimeout": tea.IntValue(util.DefaultNumber(runtime.ConnectTimeout, client.ConnectTimeout)),
		"httpProxy":      tea.StringValue(util.DefaultString(runtime.HttpProxy, client.HttpProxy)),
		"httpsProxy":     tea.StringValue(util.DefaultString(runtime.HttpsProxy, client.HttpsProxy)),
		"noProxy":        tea.StringValue(util.DefaultString(runtime.NoProxy, client.NoProxy)),
		"maxIdleConns":   tea.IntValue(util.DefaultNumber(runtime.MaxIdleConns, client.MaxIdleConns)),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(3))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			request_.Protocol = util.DefaultString(client.Protocol, protocol)
			request_.Method = method
			request_.Pathname = pathname
			request_.Headers = map[string]*string{
				"date":                    util.GetDateUTCString(),
				"host":                    client.Endpoint,
				"accept":                  tea.String("application/json"),
				"x-acs-signature-nonce":   util.GetNonce(),
				"x-acs-signature-method":  tea.String("HMAC-SHA1"),
				"x-acs-signature-version": tea.String("1.0"),
				"x-eventbridge-version":   tea.String("2015-06-06"),
				"user-agent":              util.GetUserAgent(tea.String(" aliyun-eventbridge-sdk/1.2.0")),
			}
			if !tea.BoolValue(util.IsUnset(client.RegionId)) {
				request_.Headers["x-eventbridge-regionId"] = client.RegionId
			}

			if !tea.BoolValue(util.IsUnset(body)) {
				request_.Body = tea.ToReader(util.ToJSONString(body))
				request_.Headers["content-type"] = tea.String("application/json; charset=utf-8")
			}

			if tea.BoolValue(util.EqualString(action, tea.String("putEvents"))) {
				request_.Headers["content-type"] = tea.String("application/cloudevents-batch+json; charset=utf-8")
			}

			if !tea.BoolValue(util.IsUnset(query)) {
				request_.Query = query
			}

			accessKeyId, _err := client.Credential.GetAccessKeyId()
			if _err != nil {
				return _result, _err
			}

			accessKeySecret, _err := client.Credential.GetAccessKeySecret()
			if _err != nil {
				return _result, _err
			}

			securityToken, _err := client.Credential.GetSecurityToken()
			if _err != nil {
				return _result, _err
			}

			if !tea.BoolValue(util.Empty(securityToken)) {
				request_.Headers["x-acs-accesskey-id"] = accessKeyId
				request_.Headers["x-acs-security-token"] = securityToken
			}

			stringToSign := eventbridgeutil.GetStringToSign(request_)
			request_.Headers["authorization"] = tea.String("acs:" + tea.StringValue(accessKeyId) + ":" + tea.StringValue(eventbridgeutil.GetSignature(stringToSign, accessKeySecret)))
			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			result, _err := util.ReadAsJSON(response_.Body)
			if _err != nil {
				return _result, _err
			}

			tmp := util.AssertAsMap(result)
			if tea.BoolValue(util.Is4xx(response_.StatusCode)) || tea.BoolValue(util.Is5xx(response_.StatusCode)) {
				_err = tea.NewSDKError(map[string]interface{}{
					"code":    tmp["code"],
					"message": "[EventBridgeError-" + tea.ToString(tmp["requestId"]) + "] " + tea.ToString(tmp["message"]),
					"data":    tmp,
				})
				return _result, _err
			}

			_result = tmp
			return _result, _err
		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

/**
 * Publish event to the aliyun EventBus
 */
func (client *Client) PutEvents(eventList []*CloudEvent) (_result *PutEventsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &PutEventsResponse{}
	_body, _err := client.PutEventsWithOptions(eventList, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Publish event to the aliyun EventBus
 */
func (client *Client) PutEventsWithOptions(eventList []*CloudEvent, runtime *util.RuntimeOptions) (_result *PutEventsResponse, _err error) {
	for _, cloudEvent := range eventList {
		if tea.BoolValue(util.IsUnset(cloudEvent.Specversion)) {
			cloudEvent.Specversion = tea.String("1.0")
		}

		if tea.BoolValue(util.IsUnset(cloudEvent.Datacontenttype)) {
			cloudEvent.Datacontenttype = tea.String("application/json; charset=utf-8")
		}

		_err = util.ValidateModel(cloudEvent)
		if _err != nil {
			return _result, _err
		}
	}
	body := eventbridgeutil.Serialize(eventList)
	_result = &PutEventsResponse{}
	_body, _err := client.DoRequest(tea.String("putEvents"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/putEvents"), nil, body, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Creates a new event bus within your account. This can be a custom event bus which you can use to receive events from your custom applications and services
 */
func (client *Client) CreateEventBus(request *CreateEventBusRequest) (_result *CreateEventBusResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &CreateEventBusResponse{}
	_body, _err := client.CreateEventBusWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Creates a new event bus within your account. This can be a custom event bus which you can use to receive events from your custom applications and services
 */
func (client *Client) CreateEventBusWithOptions(request *CreateEventBusRequest, runtime *util.RuntimeOptions) (_result *CreateEventBusResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &CreateEventBusResponse{}
	_body, _err := client.DoRequest(tea.String("createEventBus"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/createEventBus"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Deletes the specified custom event bus in your account,You can't delete your account's default event bus
 */
func (client *Client) DeleteEventBus(request *DeleteEventBusRequest) (_result *DeleteEventBusResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &DeleteEventBusResponse{}
	_body, _err := client.DeleteEventBusWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Deletes the specified custom event bus in your account,You can't delete your account's default event bus
 */
func (client *Client) DeleteEventBusWithOptions(request *DeleteEventBusRequest, runtime *util.RuntimeOptions) (_result *DeleteEventBusResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &DeleteEventBusResponse{}
	_body, _err := client.DoRequest(tea.String("deleteEventBus"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/deleteEventBus"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Displays details about an event bus in your account
 */
func (client *Client) GetEventBus(request *GetEventBusRequest) (_result *GetEventBusResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &GetEventBusResponse{}
	_body, _err := client.GetEventBusWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Displays details about an event bus in your account
 */
func (client *Client) GetEventBusWithOptions(request *GetEventBusRequest, runtime *util.RuntimeOptions) (_result *GetEventBusResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &GetEventBusResponse{}
	_body, _err := client.DoRequest(tea.String("getEventBus"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/getEventBus"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * List all the EventBus in your account, including the default event bus, custom event buses, which meet the search criteria.
 */
func (client *Client) ListEventBuses(request *ListEventBusesRequest) (_result *ListEventBusesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &ListEventBusesResponse{}
	_body, _err := client.ListEventBusesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * List all the EventBus in your account, including the default event bus, custom event buses, which meet the search criteria.
 */
func (client *Client) ListEventBusesWithOptions(request *ListEventBusesRequest, runtime *util.RuntimeOptions) (_result *ListEventBusesResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &ListEventBusesResponse{}
	_body, _err := client.DoRequest(tea.String("listEventBuses"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/listEventBuses"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Create an EventBus rule on Aliyun
 */
func (client *Client) CreateRule(request *CreateRuleRequest) (_result *CreateRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &CreateRuleResponse{}
	_body, _err := client.CreateRuleWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Create an EventBus rule on Aliyun
 */
func (client *Client) CreateRuleWithOptions(request *CreateRuleRequest, runtime *util.RuntimeOptions) (_result *CreateRuleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &CreateRuleResponse{}
	_body, _err := client.DoRequest(tea.String("createRule"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/createRule"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Deletes the specified rule.
 */
func (client *Client) DeleteRule(request *DeleteRuleRequest) (_result *DeleteRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &DeleteRuleResponse{}
	_body, _err := client.DeleteRuleWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Deletes the specified rule.
 */
func (client *Client) DeleteRuleWithOptions(request *DeleteRuleRequest, runtime *util.RuntimeOptions) (_result *DeleteRuleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &DeleteRuleResponse{}
	_body, _err := client.DoRequest(tea.String("deleteRule"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/deleteRule"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Disables the specified rule
 */
func (client *Client) DisableRule(request *DisableRuleRequest) (_result *DisableRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &DisableRuleResponse{}
	_body, _err := client.DisableRuleWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Disables the specified rule
 */
func (client *Client) DisableRuleWithOptions(request *DisableRuleRequest, runtime *util.RuntimeOptions) (_result *DisableRuleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &DisableRuleResponse{}
	_body, _err := client.DoRequest(tea.String("disableRule"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/disableRule"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Enables the specified rule
 */
func (client *Client) EnableRule(request *EnableRuleRequest) (_result *EnableRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &EnableRuleResponse{}
	_body, _err := client.EnableRuleWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Enables the specified rule
 */
func (client *Client) EnableRuleWithOptions(request *EnableRuleRequest, runtime *util.RuntimeOptions) (_result *EnableRuleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &EnableRuleResponse{}
	_body, _err := client.DoRequest(tea.String("enableRule"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/enableRule"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Describes the specified rule
 */
func (client *Client) GetRule(request *GetRuleRequest) (_result *GetRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &GetRuleResponse{}
	_body, _err := client.GetRuleWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Describes the specified rule
 */
func (client *Client) GetRuleWithOptions(request *GetRuleRequest, runtime *util.RuntimeOptions) (_result *GetRuleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &GetRuleResponse{}
	_body, _err := client.DoRequest(tea.String("getRule"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/getRule"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * List all the rules which meet the search criteria
 */
func (client *Client) ListRules(request *ListRulesRequest) (_result *ListRulesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &ListRulesResponse{}
	_body, _err := client.ListRulesWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * List all the rules which meet the search criteria
 */
func (client *Client) ListRulesWithOptions(request *ListRulesRequest, runtime *util.RuntimeOptions) (_result *ListRulesResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &ListRulesResponse{}
	_body, _err := client.DoRequest(tea.String("listRules"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/listRules"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * update the specified rule.
 */
func (client *Client) UpdateRule(request *UpdateRuleRequest) (_result *UpdateRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &UpdateRuleResponse{}
	_body, _err := client.UpdateRuleWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * update the specified rule.
 */
func (client *Client) UpdateRuleWithOptions(request *UpdateRuleRequest, runtime *util.RuntimeOptions) (_result *UpdateRuleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &UpdateRuleResponse{}
	_body, _err := client.DoRequest(tea.String("updateRule"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/updateRule"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Adds the specified targets to the specified rule
 */
func (client *Client) CreateTargets(request *CreateTargetsRequest) (_result *CreateTargetsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &CreateTargetsResponse{}
	_body, _err := client.CreateTargetsWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Adds the specified targets to the specified rule
 */
func (client *Client) CreateTargetsWithOptions(request *CreateTargetsRequest, runtime *util.RuntimeOptions) (_result *CreateTargetsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &CreateTargetsResponse{}
	_body, _err := client.DoRequest(tea.String("createTargets"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/createTargets"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Delete the specified targets from the specified rule
 */
func (client *Client) DeleteTargets(request *DeleteTargetsRequest) (_result *DeleteTargetsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &DeleteTargetsResponse{}
	_body, _err := client.DeleteTargetsWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Delete the specified targets from the specified rule
 */
func (client *Client) DeleteTargetsWithOptions(request *DeleteTargetsRequest, runtime *util.RuntimeOptions) (_result *DeleteTargetsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &DeleteTargetsResponse{}
	_body, _err := client.DoRequest(tea.String("deleteTargets"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/deleteTargets"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * List all the Targets which meet the search criteria
 */
func (client *Client) ListTargets(request *ListTargetsRequest) (_result *ListTargetsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &ListTargetsResponse{}
	_body, _err := client.ListTargetsWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * List all the Targets which meet the search criteria
 */
func (client *Client) ListTargetsWithOptions(request *ListTargetsRequest, runtime *util.RuntimeOptions) (_result *ListTargetsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &ListTargetsResponse{}
	_body, _err := client.DoRequest(tea.String("listTargets"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/listTargets"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * Tests whether the specified event pattern matches the provided event
 */
func (client *Client) TestEventPattern(request *TestEventPatternRequest) (_result *TestEventPatternResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &TestEventPatternResponse{}
	_body, _err := client.TestEventPatternWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * Tests whether the specified event pattern matches the provided event
 */
func (client *Client) TestEventPatternWithOptions(request *TestEventPatternRequest, runtime *util.RuntimeOptions) (_result *TestEventPatternResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	_result = &TestEventPatternResponse{}
	_body, _err := client.DoRequest(tea.String("testEventPattern"), tea.String("HTTP"), tea.String("POST"), tea.String("/openapi/testEventPattern"), nil, tea.ToMap(request), runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}
