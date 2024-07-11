package sls

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GetLogRequest for GetLogsV2
type GetLogRequest struct {
	From          int64  `json:"from"`  // unix time, eg time.Now().Unix() - 900
	To            int64  `json:"to"`    // unix time, eg time.Now().Unix()
	Topic         string `json:"topic"` // @note topic is not used anymore, use __topic__ : xxx in query instead
	Lines         int64  `json:"line"`  // max 100; offset, lines and reverse is ignored when use SQL in query
	Offset        int64  `json:"offset"`
	Reverse       bool   `json:"reverse"`
	Query         string `json:"query"`
	PowerSQL      bool   `json:"powerSql"`
	FromNsPart    int32  `json:"fromNs"`
	ToNsPart      int32  `json:"toNs"`
	NeedHighlight bool   `json:"highlight"`
	IsAccurate    bool   `json:"accurate"`
}

func (glr *GetLogRequest) ToURLParams() url.Values {
	urlVal := url.Values{}
	urlVal.Add("type", "log")
	urlVal.Add("from", strconv.Itoa(int(glr.From)))
	urlVal.Add("to", strconv.Itoa(int(glr.To)))
	urlVal.Add("topic", glr.Topic)
	urlVal.Add("line", strconv.Itoa(int(glr.Lines)))
	urlVal.Add("offset", strconv.Itoa(int(glr.Offset)))
	urlVal.Add("reverse", strconv.FormatBool(glr.Reverse))
	urlVal.Add("powerSql", strconv.FormatBool(glr.PowerSQL))
	urlVal.Add("query", glr.Query)
	urlVal.Add("fromNs", strconv.Itoa(int(glr.FromNsPart)))
	urlVal.Add("toNs", strconv.Itoa(int(glr.ToNsPart)))
	urlVal.Add("highlight", strconv.FormatBool(glr.NeedHighlight))
	urlVal.Add("accurate", strconv.FormatBool(glr.IsAccurate))
	return urlVal
}

type PullLogRequest struct {
	Project          string
	Logstore         string
	ShardID          int
	Cursor           string
	EndCursor        string
	LogGroupMaxCount int
	Query            string
	// Deprecated: PullMode is not used
	PullMode     string
	QueryId      string
	CompressType int
}

func (plr *PullLogRequest) ToURLParams() url.Values {
	urlVal := url.Values{}
	urlVal.Add("type", "logs")
	urlVal.Add("cursor", plr.Cursor)
	urlVal.Add("count", strconv.Itoa(plr.LogGroupMaxCount))
	if plr.EndCursor != "" {
		urlVal.Add("end_cursor", plr.EndCursor)
	}
	if plr.Query != "" {
		urlVal.Add("query", plr.Query)
		urlVal.Add("pullMode", "scan_on_stream")
		if plr.QueryId != "" {
			urlVal.Add("queryId", plr.QueryId)
		}
	}
	return urlVal
}

type PullLogMeta struct {
	NextCursor              string
	Netflow                 int
	RawSize                 int
	RawDataCountBeforeQuery int
	RawSizeBeforeQuery      int
	Lines                   int
	LinesBeforeQuery        int
	FailedLines             int
	DataCountBeforeQuery    int
}

// GetHistogramsResponse defines response from GetHistograms call
type SingleHistogram struct {
	Progress string `json:"progress"`
	Count    int64  `json:"count"`
	From     int64  `json:"from"`
	To       int64  `json:"to"`
}

type GetHistogramsResponse struct {
	Progress   string            `json:"progress"`
	Count      int64             `json:"count"`
	Histograms []SingleHistogram `json:"histograms"`
}

func (resp *GetHistogramsResponse) IsComplete() bool {
	return strings.ToLower(resp.Progress) == "complete"
}

// GetLogsResponse defines response from GetLogs call
type GetLogsResponse struct {
	Progress string              `json:"progress"`
	Count    int64               `json:"count"`
	Logs     []map[string]string `json:"logs"`
	Contents string              `json:"contents"`
	HasSQL   bool                `json:"hasSQL"`
	Header   http.Header         `json:"header"`
}

type MetaTerm struct {
	Key  string `json:"key"`
	Term string `json:"term"`
}
type PhraseQueryInfoV3 struct {
	ScanAll     *bool  `json:"scanAll,omitempty"`
	BeginOffset *int64 `json:"beginOffset,omitempty"`
	EndOffset   *int64 `json:"endOffset,omitempty"`
	EndTime     *int64 `json:"endTime,omitempty"`
}

type GetLogsV3ResponseMeta struct {
	Progress           string  `json:"progress"`
	AggQuery           string  `json:"aggQuery"`
	WhereQuery         string  `json:"whereQuery"`
	HasSQL             bool    `json:"hasSQL"`
	ProcessedRows      int64   `json:"processedRows"`
	ElapsedMillisecond int64   `json:"elapsedMillisecond"`
	CpuSec             float64 `json:"cpuSec"`
	CpuCores           float64 `json:"cpuCores"`
	Limited            int64   `json:"limited"`
	Count              int64   `json:"count"`
	ProcessedBytes     int64   `json:"processedBytes"`
	TelemetryType      string  `json:"telementryType"` // telementryType, ignore typo
	PowerSql           bool    `json:"powerSql"`
	InsertedSql        string  `json:"insertedSQL"`

	Keys            []string            `json:"keys,omitempty"`
	Terms           []MetaTerm          `json:"terms,omitempty"`
	Marker          *string             `json:"marker,omitempty"`
	Mode            *int                `json:"mode,omitempty"`
	PhraseQueryInfo *PhraseQueryInfoV3  `json:"phraseQueryInfo,omitempty"`
	Shard           *int                `json:"shard,omitempty"`
	ScanBytes       *int64              `json:"scanBytes,omitempty"`
	IsAccurate      *bool               `json:"isAccurate,omitempty"`
	ColumnTypes     []string            `json:"columnTypes,omitempty"`
	Highlights      []map[string]string `json:"highlights,omitempty"`
}

type PhraseQueryInfoV2 struct {
	ScanAll     string `json:"scanAll,omitempty"`
	BeginOffset string `json:"beginOffset,omitempty"`
	EndOffset   string `json:"endOffset,omitempty"`
	EndTime     string `json:"endTime,omitempty"`
}

func (s *PhraseQueryInfoV3) toPhraseQueryInfoV2() *PhraseQueryInfoV2 {
	if s == nil {
		return nil
	}
	return &PhraseQueryInfoV2{
		ScanAll:     BoolPtrToStringNum(s.ScanAll),
		BeginOffset: Int64PtrToString(s.BeginOffset),
		EndOffset:   Int64PtrToString(s.EndOffset),
		EndTime:     Int64PtrToString(s.EndTime),
	}
}

type QueryInfoV2 struct {
	Keys            []string            `json:"keys,omitempty"`
	Terms           [][]string          `json:"terms,omitempty"` // [[term, key], [term2, key2]]
	Limited         string              `json:"limited,omitempty"`
	Marker          *string             `json:"marker,omitempty"`
	Mode            *int                `json:"mode,omitempty"`
	PhraseQueryInfo *PhraseQueryInfoV2  `json:"phraseQueryInfo,omitempty"`
	Shard           *int                `json:"shard,omitempty"`
	ScanBytes       *int64              `json:"scanBytes,omitempty"`
	IsAccurate      *int64              `json:"isAccurate,omitempty"`
	ColumnTypes     []string            `json:"columnTypes,omitempty"`
	Highlights      []map[string]string `json:"highlight,omitempty"`
}

func (meta *GetLogsV3ResponseMeta) constructQueryInfo() (string, error) {
	var terms [][]string
	for _, term := range meta.Terms {
		terms = append(terms, []string{term.Term, term.Key})
	}
	var isAccurate *int64
	if meta.IsAccurate != nil {
		res := BoolToInt64(*meta.IsAccurate)
		isAccurate = &res
	}
	limited := ""
	if meta.Limited != 0 {
		limited = strconv.FormatInt(meta.Limited, 10)
	}
	queryInfo := &QueryInfoV2{
		Keys:            meta.Keys,
		Terms:           terms,
		Limited:         limited,
		Marker:          meta.Marker,
		Mode:            meta.Mode,
		PhraseQueryInfo: meta.PhraseQueryInfo.toPhraseQueryInfoV2(),
		Shard:           meta.Shard,
		ScanBytes:       meta.ScanBytes,
		IsAccurate:      isAccurate,
		ColumnTypes:     meta.ColumnTypes,
		Highlights:      meta.Highlights,
	}
	contents, err := json.Marshal(queryInfo)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// GetLogsV3Response defines response from GetLogs call
type GetLogsV3Response struct {
	Meta GetLogsV3ResponseMeta `json:"meta"`
	Logs []map[string]string   `json:"data"`
}

func (resp *GetLogsV3Response) IsComplete() bool {
	return strings.ToLower(resp.Meta.Progress) == "complete"
}

// GetLogLinesResponse defines response from GetLogLines call
// note: GetLogLinesResponse.Logs is nil when use GetLogLinesResponse
type GetLogLinesResponse struct {
	GetLogsResponse
	Lines []json.RawMessage
}

func (resp *GetLogsResponse) IsComplete() bool {
	return strings.ToLower(resp.Progress) == "complete"
}

func (resp *GetLogsResponse) GetKeys() (error, []string) {
	type Content map[string][]interface{}
	var content Content
	err := json.Unmarshal([]byte(resp.Contents), &content)
	if err != nil {
		return err, nil
	}
	result := []string{}
	for _, v := range content["keys"] {
		result = append(result, v.(string))
	}
	return nil, result
}

type GetContextLogsResponse struct {
	Progress     string              `json:"progress"`
	TotalLines   int64               `json:"total_lines"`
	BackLines    int64               `json:"back_lines"`
	ForwardLines int64               `json:"forward_lines"`
	Logs         []map[string]string `json:"logs"`
}

func (resp *GetContextLogsResponse) IsComplete() bool {
	return strings.ToLower(resp.Progress) == "complete"
}

type JsonKey struct {
	Type     string `json:"type"`
	Alias    string `json:"alias,omitempty"`
	DocValue bool   `json:"doc_value,omitempty"`
}

// IndexKey ...
type IndexKey struct {
	Token         []string            `json:"token"` // tokens that split the log line.
	CaseSensitive bool                `json:"caseSensitive"`
	Type          string              `json:"type"` // text, long, double
	DocValue      bool                `json:"doc_value,omitempty"`
	Alias         string              `json:"alias,omitempty"`
	Chn           bool                `json:"chn"` // parse chinese or not
	JsonKeys      map[string]*JsonKey `json:"json_keys,omitempty"`
}

type IndexLine struct {
	Token         []string `json:"token"`
	CaseSensitive bool     `json:"caseSensitive"`
	IncludeKeys   []string `json:"include_keys,omitempty"`
	ExcludeKeys   []string `json:"exclude_keys,omitempty"`
	Chn           bool     `json:"chn"` // parse chinese or not
}

// Index is an index config for a log store.
type Index struct {
	Keys                   map[string]IndexKey `json:"keys,omitempty"`
	Line                   *IndexLine          `json:"line,omitempty"`
	Ttl                    uint32              `json:"ttl,omitempty"`
	MaxTextLen             uint32              `json:"max_text_len,omitempty"`
	LogReduce              bool                `json:"log_reduce"`
	LogReduceWhiteListDict []string            `json:"log_reduce_white_list,omitempty"`
	LogReduceBlackListDict []string            `json:"log_reduce_black_list,omitempty"`
}

// CreateDefaultIndex return a full text index config
func CreateDefaultIndex() *Index {
	return &Index{
		Line: &IndexLine{
			Token:         []string{" ", "\n", "\t", "\r", ",", ";", "[", "]", "{", "}", "(", ")", "&", "^", "*", "#", "@", "~", "=", "<", ">", "/", "\\", "?", ":", "'", "\""},
			CaseSensitive: false,
		},
	}
}

type GetMeteringModeResponse struct {
	MeteringMode string `json:"meteringMode"`
}

const (
	CHARGE_BY_FUNCTION    = "ChargeByFunction"
	CHARGE_BY_DATA_INGEST = "ChargeByDataIngest"
)

type PostLogStoreLogsRequest struct {
	LogGroup     *LogGroup
	HashKey      *string
	CompressType int
}

type StoreView struct {
	Name      string            `json:"name"`
	StoreType string            `json:"storeType"`
	Stores    []*StoreViewStore `json:"stores"`
}

// storeType of storeView
const (
	STORE_VIEW_STORE_TYPE_LOGSTORE    = "logstore"
	STORE_VIEW_STORE_TYPE_METRICSTORE = "metricstore"
)

type StoreViewStore struct {
	Project   string `json:"project"`
	StoreName string `json:"storeName"`
	Query     string `json:"query,omitempty"`
}

type GetStoreViewIndexResponse struct {
	Indexes         []*StoreViewIndex  `json:"indexes"`
	StoreViewErrors []*StoreViewErrors `json:"storeViewErrors"`
}

type StoreViewIndex struct {
	ProjectName string `json:"projectName"`
	LogStore    string `json:"logstore"`
	Index       Index  `json:"index"`
}

type StoreViewErrors struct {
	ProjectName string `json:"projectName"`
	LogStore    string `json:"logstore"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type ListStoreViewsRequest struct {
	Offset int `json:"offset"`
	Size   int `json:"size"`
}

type ListStoreViewsResponse struct {
	Total      int      `json:"total"`
	Count      int      `json:"count"`
	StoreViews []string `json:"storeviews"`
}
