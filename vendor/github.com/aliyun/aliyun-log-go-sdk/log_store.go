package sls

import (
	"encoding/json"
	"fmt"
	"time"

	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/go-kit/kit/log/level"
	"github.com/gogo/protobuf/proto"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
)

// this file is deprecated and no maintenance
// see client_logstore.go

var (
	zstdReader = newZstdReader()
	zstdWriter = newZstdWriter()
)

func newZstdReader() *zstd.Decoder {
	r, _ := zstd.NewReader(nil)
	return r
}

func newZstdWriter() *zstd.Encoder {
	w, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	return w
}

// LogStore defines LogStore struct
type LogStore struct {
	Name          string `json:"logstoreName"`
	TTL           int    `json:"ttl"`
	ShardCount    int    `json:"shardCount"`
	WebTracking   bool   `json:"enable_tracking"`
	AutoSplit     bool   `json:"autoSplit"`
	MaxSplitShard int    `json:"maxSplitShard"`

	AppendMeta    bool   `json:"appendMeta"`
	TelemetryType string `json:"telemetryType"`
	HotTTL        int32  `json:"hot_ttl,omitempty"`
	Mode          string `json:"mode,omitempty"` // "query" or "standard"(default), can't be modified after creation

	CreateTime     uint32 `json:"createTime,omitempty"`
	LastModifyTime uint32 `json:"lastModifyTime,omitempty"`

	project            *LogProject
	putLogCompressType int
	EncryptConf        *EncryptConf `json:"encrypt_conf,omitempty"`
	ProductType        string       `json:"productType,omitempty"`
	useMetricStoreURL  bool
}

// Shard defines shard struct
type Shard struct {
	ShardID           int    `json:"shardID"`
	Status            string `json:"status"`
	InclusiveBeginKey string `json:"inclusiveBeginKey"`
	ExclusiveBeginKey string `json:"exclusiveEndKey"`
	CreateTime        int    `json:"createTime"`
}

// encrypt struct
type EncryptConf struct {
	Enable      bool                `json:"enable"`
	EncryptType string              `json:"encrypt_type"`
	UserCmkInfo *EncryptUserCmkConf `json:"user_cmk_info,omitempty"`
}

// EncryptUserCmkConf struct
type EncryptUserCmkConf struct {
	CmkKeyId string `json:"cmk_key_id"`
	Arn      string `json:"arn"`
	RegionId string `json:"region_id"`
}

// NewLogStore ...
func NewLogStore(logStoreName string, project *LogProject) (*LogStore, error) {
	return &LogStore{
		Name:    logStoreName,
		project: project,
	}, nil
}

// SetPutLogCompressType set put log's compress type, default lz4
func (s *LogStore) SetPutLogCompressType(compressType int) error {
	if compressType < 0 || compressType >= Compress_Max {
		return InvalidCompressError
	}
	s.putLogCompressType = compressType
	return nil
}

// ListShards returns shard id list of this logstore.
func (s *LogStore) ListShards() (shardIDs []*Shard, err error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/logstores/%v/shards", s.Name)
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return nil, NewClientError(err)
	}
	defer r.Body.Close()
	buf, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := &Error{}
		if jErr := json.Unmarshal(buf, err); jErr != nil {
			return nil, NewBadResponseError(string(buf), r.Header, r.StatusCode)
		}
		return nil, err
	}

	var shards []*Shard
	err = json.Unmarshal(buf, &shards)
	if err != nil {
		return nil, NewBadResponseError(string(buf), r.Header, r.StatusCode)
	}
	return shards, nil
}

func copyIncompressible(src, dst []byte) (int, error) {
	lLen, dn := len(src), len(dst)

	di := 0
	if lLen < 0xF {
		dst[di] = byte(lLen << 4)
	} else {
		dst[di] = 0xF0
		if di++; di == dn {
			return di, nil
		}
		lLen -= 0xF
		for ; lLen >= 0xFF; lLen -= 0xFF {
			dst[di] = 0xFF
			if di++; di == dn {
				return di, nil
			}
		}
		dst[di] = byte(lLen)
	}
	if di++; di+len(src) > dn {
		return di, nil
	}
	di += copy(dst[di:], src)
	return di, nil
}

// PutRawLog put raw log data to log service, no marshal
func (s *LogStore) PutRawLog(rawLogData []byte) (err error) {
	if len(rawLogData) == 0 {
		// empty log group
		return nil
	}

	var out []byte
	var h map[string]string
	var outLen int
	switch s.putLogCompressType {
	case Compress_LZ4:
		// Compresse body with lz4
		out = make([]byte, lz4.CompressBlockBound(len(rawLogData)))
		var hashTable [1 << 16]int
		n, err := lz4.CompressBlock(rawLogData, out, hashTable[:])
		if err != nil {
			return NewClientError(err)
		}
		// copy incompressible data as lz4 format
		if n == 0 {
			n, _ = copyIncompressible(rawLogData, out)
		}

		h = map[string]string{
			"x-log-compresstype": "lz4",
			"x-log-bodyrawsize":  strconv.Itoa(len(rawLogData)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = n
	case Compress_ZSTD:
		out = zstdWriter.EncodeAll(rawLogData, nil)
		h = map[string]string{
			"x-log-compresstype": "zstd",
			"x-log-bodyrawsize":  strconv.Itoa(len(rawLogData)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = len(out)
	case Compress_None:
		// no compress
		out = rawLogData
		h = map[string]string{
			"x-log-bodyrawsize": strconv.Itoa(len(rawLogData)),
			"Content-Type":      "application/x-protobuf",
		}
		outLen = len(out)
	}

	uri := fmt.Sprintf("/logstores/%v", s.Name)
	r, err := request(s.project, "POST", uri, h, out[:outLen])
	if err != nil {
		return NewClientError(err)
	}
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(body, err); jErr != nil {
			return NewBadResponseError(string(body), r.Header, r.StatusCode)
		}
		return err
	}
	return nil
}

func (s *LogStore) PostRawLogs(body []byte, hashKey *string) (err error) {
	if len(body) == 0 {
		// empty log group or empty hashkey
		return nil
	}

	if hashKey == nil || *hashKey == "" {
		// empty hash call PutLogs
		return s.PutRawLog(body)
	}

	var out []byte
	var h map[string]string
	var outLen int
	switch s.putLogCompressType {
	case Compress_LZ4:
		// Compresse body with lz4
		out = make([]byte, lz4.CompressBlockBound(len(body)))
		var hashTable [1 << 16]int
		n, err := lz4.CompressBlock(body, out, hashTable[:])
		if err != nil {
			return NewClientError(err)
		}
		// copy incompressible data as lz4 format
		if n == 0 {
			n, _ = copyIncompressible(body, out)
		}

		h = map[string]string{
			"x-log-compresstype": "lz4",
			"x-log-bodyrawsize":  strconv.Itoa(len(body)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = n
	case Compress_ZSTD:
		// Compress body with zstd
		out = zstdWriter.EncodeAll(body, nil)
		h = map[string]string{
			"x-log-compresstype": "zstd",
			"x-log-bodyrawsize":  strconv.Itoa(len(body)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = len(out)
	case Compress_None:
		// no compress
		out = body
		h = map[string]string{
			"x-log-bodyrawsize": strconv.Itoa(len(body)),
			"Content-Type":      "application/x-protobuf",
		}
		outLen = len(out)
	}

	uri := fmt.Sprintf("/logstores/%v/shards/route?key=%v", s.Name, *hashKey)
	r, err := request(s.project, "POST", uri, h, out[:outLen])
	if err != nil {
		return NewClientError(err)
	}
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(body, err); jErr != nil {
			return NewBadResponseError(string(body), r.Header, r.StatusCode)
		}
		return err
	}
	return nil
}

// PutLogs put logs into logstore.
// The callers should transform user logs into LogGroup.
func (s *LogStore) PutLogs(lg *LogGroup) (err error) {
	if len(lg.Logs) == 0 {
		// empty log group
		return nil
	}

	body, err := proto.Marshal(lg)
	if err != nil {
		return NewClientError(err)
	}

	var out []byte
	var h map[string]string
	var outLen int
	switch s.putLogCompressType {
	case Compress_LZ4:
		// Compresse body with lz4
		out = make([]byte, lz4.CompressBlockBound(len(body)))
		var hashTable [1 << 16]int
		n, err := lz4.CompressBlock(body, out, hashTable[:])
		if err != nil {
			return NewClientError(err)
		}
		// copy incompressible data as lz4 format
		if n == 0 {
			n, _ = copyIncompressible(body, out)
		}

		h = map[string]string{
			"x-log-compresstype": "lz4",
			"x-log-bodyrawsize":  strconv.Itoa(len(body)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = n
	case Compress_ZSTD:
		// Compress body with zstd
		out = zstdWriter.EncodeAll(body, nil)
		h = map[string]string{
			"x-log-compresstype": "zstd",
			"x-log-bodyrawsize":  strconv.Itoa(len(body)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = len(out)
	case Compress_None:
		// no compress
		out = body
		h = map[string]string{
			"x-log-bodyrawsize": strconv.Itoa(len(body)),
			"Content-Type":      "application/x-protobuf",
		}
		outLen = len(out)
	}
	var uri string
	if s.useMetricStoreURL {
		uri = fmt.Sprintf("/prometheus/%s/%s/api/v1/write", s.project.Name, s.Name)
	} else {
		uri = fmt.Sprintf("/logstores/%v", s.Name)
	}
	r, err := request(s.project, "POST", uri, h, out[:outLen])
	if err != nil {
		return NewClientError(err)
	}
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(body, err); jErr != nil {
			return NewBadResponseError(string(body), r.Header, r.StatusCode)
		}
		return err
	}
	return nil
}

// PostLogStoreLogs put logs into Shard logstore by hashKey.
// The callers should transform user logs into LogGroup.
func (s *LogStore) PostLogStoreLogs(lg *LogGroup, hashKey *string) (err error) {
	if len(lg.Logs) == 0 {
		// empty log group or empty hashkey
		return nil
	}

	if hashKey == nil || *hashKey == "" || s.useMetricStoreURL {
		// empty hash call PutLogs
		return s.PutLogs(lg)
	}

	body, err := proto.Marshal(lg)
	if err != nil {
		return NewClientError(err)
	}

	var out []byte
	var h map[string]string
	var outLen int
	switch s.putLogCompressType {
	case Compress_LZ4:
		// Compresse body with lz4
		out = make([]byte, lz4.CompressBlockBound(len(body)))
		var hashTable [1 << 16]int
		n, err := lz4.CompressBlock(body, out, hashTable[:])
		if err != nil {
			return NewClientError(err)
		}
		// copy incompressible data as lz4 format
		if n == 0 {
			n, _ = copyIncompressible(body, out)
		}

		h = map[string]string{
			"x-log-compresstype": "lz4",
			"x-log-bodyrawsize":  strconv.Itoa(len(body)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = n
	case Compress_ZSTD:
		// Compress body with zstd
		out = zstdWriter.EncodeAll(body, nil)
		h = map[string]string{
			"x-log-compresstype": "zstd",
			"x-log-bodyrawsize":  strconv.Itoa(len(body)),
			"Content-Type":       "application/x-protobuf",
		}
		outLen = len(out)
	case Compress_None:
		// no compress
		out = body
		h = map[string]string{
			"x-log-bodyrawsize": strconv.Itoa(len(body)),
			"Content-Type":      "application/x-protobuf",
		}
		outLen = len(out)
	}

	uri := fmt.Sprintf("/logstores/%v/shards/route?key=%v", s.Name, *hashKey)
	r, err := request(s.project, "POST", uri, h, out[:outLen])
	if err != nil {
		return NewClientError(err)
	}
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(body, err); jErr != nil {
			return NewBadResponseError(string(body), r.Header, r.StatusCode)
		}
		return err
	}
	return nil
}

// GetCursor gets log cursor of one shard specified by shardId.
// The from can be in three form: a) unix timestamp in seccond, b) "begin", c) "end".
// For more detail please read: https://help.aliyun.com/document_detail/29024.html
func (s *LogStore) GetCursor(shardID int, from string) (cursor string, err error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/logstores/%v/shards/%v?type=cursor&from=%v",
		s.Name, shardID, from)
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	if r.StatusCode != http.StatusOK {
		errMsg := &Error{}
		err = json.Unmarshal(buf, errMsg)
		if err != nil {
			err = fmt.Errorf("failed to get cursor")
			dump, _ := httputil.DumpResponse(r, true)
			if IsDebugLevelMatched(1) {
				level.Error(Logger).Log("msg", string(dump))
			}
			return
		}
		err = fmt.Errorf("%v:%v", errMsg.Code, errMsg.Message)
		return
	}

	type Body struct {
		Cursor string
	}
	body := &Body{}

	err = json.Unmarshal(buf, body)
	if err != nil {
		return "", NewBadResponseError(string(buf), r.Header, r.StatusCode)
	}
	cursor = body.Cursor
	return cursor, nil
}

func (s *LogStore) GetLogsBytes(shardID int, cursor, endCursor string,
	logGroupMaxCount int) (out []byte, nextCursor string, err error) {
	plr := &PullLogRequest{
		ShardID:          shardID,
		Cursor:           cursor,
		EndCursor:        endCursor,
		LogGroupMaxCount: logGroupMaxCount,
	}
	return s.GetLogsBytesV2(plr)
}

// Deprecated: use GetLogsBytesWithQuery instead
func (s *LogStore) GetLogsBytesV2(plr *PullLogRequest) ([]byte, string, error) {
	out, plm, err := s.GetLogsBytesWithQuery(plr)
	if err != nil {
		return nil, "", err
	}
	return out, plm.NextCursor, err
}

// GetLogsBytes gets logs binary data from shard specified by shardId according cursor and endCursor.
// The logGroupMaxCount is the max number of logGroup could be returned.
// The nextCursor is the next curosr can be used to read logs at next time.
func (s *LogStore) GetLogsBytesWithQuery(plr *PullLogRequest) (out []byte, pullLogMeta *PullLogMeta, err error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Accept":            "application/x-protobuf",
	}
	if plr.CompressType < 0 || Compress_LZ4 >= Compress_Max {
		return nil, nil, fmt.Errorf("unsupported compress type: %d", plr.CompressType)
	}
	if plr.CompressType == Compress_ZSTD {
		h["Accept-Encoding"] = "zstd"
	} else {
		h["Accept-Encoding"] = "lz4"
	}
	urlVal := plr.ToURLParams()
	uri := fmt.Sprintf("/logstores/%v/shards/%v?%s", s.Name, plr.ShardID, urlVal.Encode())

	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	pullLogMeta = &PullLogMeta{}
	pullLogMeta.Netflow = len(buf)
	if r.StatusCode != http.StatusOK {
		errMsg := &Error{}
		err = json.Unmarshal(buf, errMsg)
		if err != nil {
			err = fmt.Errorf("failed to get cursor")
			dump, _ := httputil.DumpResponse(r, true)
			if IsDebugLevelMatched(1) {
				level.Error(Logger).Log("msg", string(dump))
			}
			return
		}
		err = fmt.Errorf("%v:%v", errMsg.Code, errMsg.Message)
		return
	}
	v, ok := r.Header["X-Log-Compresstype"]
	if !ok || len(v) == 0 {
		err = fmt.Errorf("can't find 'x-log-compresstype' header")
		return
	}
	var compressType = Compress_None
	if v[0] == "lz4" {
		compressType = Compress_LZ4
	} else if v[0] == "zstd" {
		compressType = Compress_ZSTD
	} else {
		err = fmt.Errorf("unexpected compress type:%v", compressType)
		return
	}

	v, ok = r.Header["X-Log-Cursor"]
	if !ok || len(v) == 0 {
		err = fmt.Errorf("can't find 'x-log-cursor' header")
		return
	}
	pullLogMeta.NextCursor = v[0]
	pullLogMeta.RawSize, err = ParseHeaderInt(r, "X-Log-Bodyrawsize")
	if err != nil {
		return
	}
	if pullLogMeta.RawSize > 0 {
		out = make([]byte, pullLogMeta.RawSize)
		switch compressType {
		case Compress_LZ4:
			uncompressedSize := 0
			if uncompressedSize, err = lz4.UncompressBlock(buf, out); err != nil {
				return
			}
			if uncompressedSize != pullLogMeta.RawSize {
				return nil, nil, fmt.Errorf("uncompressed size %d does not match 'x-log-bodyrawsize' %d", uncompressedSize, pullLogMeta.RawSize)
			}
		case Compress_ZSTD:
			out, err = zstdReader.DecodeAll(buf, out[:0])
			if err != nil {
				return nil, nil, err
			}
			if len(out) != pullLogMeta.RawSize {
				return nil, nil, fmt.Errorf("uncompressed size %d does not match 'x-log-bodyrawsize' %d", len(out), pullLogMeta.RawSize)
			}
		default:
			return nil, nil, fmt.Errorf("unexpected compress type: %d", compressType)
		}
	}
	// todo: add query meta
	// If query is not nil, extract more headers
	// if plr.Query != "" {
	// 	pullLogMeta.RawSizeBeforeQuery, err = ParseHeaderInt(r, "X-Log-Rawdatasize")
	// 	if err != nil {
	// 		return
	// 	}
	// 	pullLogMeta.DataCountBeforeQuery, err = ParseHeaderInt(r, "X-Log-Rawdatacount")
	// 	if err != nil {
	// 		return
	// 	}
	// 	pullLogMeta.Lines, err = ParseHeaderInt(r, "X-Log-Resultlines")
	// 	if err != nil {
	// 		return
	// 	}
	// 	pullLogMeta.LinesBeforeQuery, err = ParseHeaderInt(r, "X-Log-Rawdatalines")
	// 	if err != nil {
	// 		return
	// 	}
	// 	pullLogMeta.FailedLines, err = ParseHeaderInt(r, "X-Log-Failedlines")
	// 	if err != nil {
	// 		return
	// 	}
	// }
	return
}

// LogsBytesDecode decodes logs binary data returned by GetLogsBytes API
func LogsBytesDecode(data []byte) (gl *LogGroupList, err error) {

	gl = &LogGroupList{}
	err = proto.Unmarshal(data, gl)
	if err != nil {
		return nil, err
	}

	return gl, nil
}

// PullLogs gets logs from shard specified by shardId according cursor and endCursor.
// The logGroupMaxCount is the max number of logGroup could be returned.
// The nextCursor is the next cursor can be used to read logs at next time.
// @note if you want to pull logs continuous, set endCursor = ""
func (s *LogStore) PullLogs(shardID int, cursor, endCursor string,
	logGroupMaxCount int) (gl *LogGroupList, nextCursor string, err error) {
	plr := &PullLogRequest{
		ShardID:          shardID,
		Cursor:           cursor,
		EndCursor:        endCursor,
		LogGroupMaxCount: logGroupMaxCount,
	}
	return s.PullLogsV2(plr)
}

// Deprecated: use PullLogsWithQuery instead
func (s *LogStore) PullLogsV2(plr *PullLogRequest) (*LogGroupList, string, error) {
	gl, plm, err := s.PullLogsWithQuery(plr)
	if err != nil {
		return nil, "", err
	}
	return gl, plm.NextCursor, err
}

func (s *LogStore) PullLogsWithQuery(plr *PullLogRequest) (gl *LogGroupList, plm *PullLogMeta, err error) {
	out, plm, err := s.GetLogsBytesWithQuery(plr)
	if err != nil {
		return nil, nil, err
	}

	gl, err = LogsBytesDecode(out)
	if err != nil {
		return nil, nil, err
	}

	return
}

// GetHistograms query logs with [from, to) time range
func (s *LogStore) GetHistograms(topic string, from int64, to int64, queryExp string) (*GetHistogramsResponse, error) {

	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Accept":            "application/json",
	}

	urlVal := url.Values{}
	urlVal.Add("type", "histogram")
	urlVal.Add("from", strconv.Itoa(int(from)))
	urlVal.Add("to", strconv.Itoa(int(to)))
	urlVal.Add("topic", topic)
	urlVal.Add("query", queryExp)

	uri := fmt.Sprintf("/logstores/%s?%s", s.Name, urlVal.Encode())
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return nil, NewClientError(err)
	}
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(body, err); jErr != nil {
			return nil, NewBadResponseError(string(body), r.Header, r.StatusCode)
		}
		return nil, err
	}

	histograms := []SingleHistogram{}
	err = json.Unmarshal(body, &histograms)
	if err != nil {
		return nil, NewBadResponseError(string(body), r.Header, r.StatusCode)
	}

	count, err := strconv.ParseInt(r.Header.Get(GetLogsCountHeader), 10, 64)
	if err != nil {
		return nil, err
	}
	getHistogramsResponse := GetHistogramsResponse{
		Progress:   r.Header[ProgressHeader][0],
		Count:      count,
		Histograms: histograms,
	}

	return &getHistogramsResponse, nil
}

// GetLogLines query logs with [from, to) time range
func (s *LogStore) GetLogLines(topic string, from int64, to int64, queryExp string,
	maxLineNum int64, offset int64, reverse bool) (*GetLogLinesResponse, error) {

	var req GetLogRequest
	req.Topic = topic
	req.From = from
	req.To = to
	req.Query = queryExp
	req.Lines = maxLineNum
	req.Offset = offset
	req.Reverse = reverse
	return s.GetLogLinesV2(&req)
}

// GetLogLinesByNano query logs with [fromInNS, toInNs) nano time range
func (s *LogStore) GetLogLinesByNano(topic string, fromInNS int64, toInNs int64, queryExp string,
	maxLineNum int64, offset int64, reverse bool) (*GetLogLinesResponse, error) {

	var req GetLogRequest
	req.Topic = topic
	req.From = fromInNS / 1e9
	req.To = toInNs / 1e9
	req.FromNsPart = int32(fromInNS % 1e9)
	req.ToNsPart = int32(toInNs % 1e9)
	req.Query = queryExp
	req.Lines = maxLineNum
	req.Offset = offset
	req.Reverse = reverse
	return s.GetLogLinesV2(&req)
}

// GetLogLinesV2 query logs with [from, to) time range
func (s *LogStore) GetLogLinesV2(req *GetLogRequest) (*GetLogLinesResponse, error) {
	v3Rsp, httpRsp, err := s.getLogsV3Internal(req)
	if err != nil {
		return nil, err
	}

	// ensured no error
	data, _ := json.Marshal(&v3Rsp.Logs)
	var logs []json.RawMessage
	_ = json.Unmarshal(data, &logs)

	v2Rsp, err := toLogRespV2(v3Rsp, httpRsp.Header)
	if err != nil {
		return nil, err
	}

	lineRsp := GetLogLinesResponse{
		GetLogsResponse: *v2Rsp,
		Lines:           logs,
	}
	return &lineRsp, nil
}

// GetLogs query logs with [from, to) time range
func (s *LogStore) GetLogs(topic string, from int64, to int64, queryExp string,
	maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error) {
	var req GetLogRequest
	req.Topic = topic
	req.From = from
	req.To = to
	req.Query = queryExp
	req.Lines = maxLineNum
	req.Offset = offset
	req.Reverse = reverse
	return s.GetLogsV2(&req)
}

func (s *LogStore) GetLogsByNano(topic string, fromInNS int64, toInNs int64, queryExp string,
	maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error) {
	var req GetLogRequest
	req.Topic = topic
	req.From = fromInNS / 1e9
	req.To = toInNs / 1e9
	req.FromNsPart = int32(fromInNS % 1e9)
	req.ToNsPart = int32(toInNs % 1e9)
	req.Query = queryExp
	req.Lines = maxLineNum
	req.Offset = offset
	req.Reverse = reverse
	return s.GetLogsV2(&req)
}

func (s *LogStore) getToCompleted(f func() (bool, error)) {
	interval := 100 * time.Millisecond
	retryCount := MaxCompletedRetryCount
	isCompleted := false
	timeoutTime := time.Now().Add(MaxCompletedRetryLatency)
	for retryCount > 0 && timeoutTime.After(time.Now()) {
		var err error
		isCompleted, err = f()
		if err != nil || isCompleted {
			return
		}
		time.Sleep(interval)
		retryCount--
		if interval < 10*time.Second {
			interval = interval * 2
		}
		if interval > 10*time.Second {
			interval = 10 * time.Second
		}
	}
	return
}

// GetLogsToCompleted query logs with [from, to) time range to completed
func (s *LogStore) GetLogsToCompleted(topic string, from int64, to int64, queryExp string,
	maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error) {
	var res *GetLogsResponse
	var err error
	f := func() (bool, error) {
		res, err = s.GetLogs(topic, from, to, queryExp, maxLineNum, offset, reverse)
		if err == nil {
			return res.IsComplete(), nil
		}
		return false, err
	}
	s.getToCompleted(f)
	return res, err
}

// GetLogsToCompletedV2 query logs with [from, to) time range to completed
func (s *LogStore) GetLogsToCompletedV2(req *GetLogRequest) (*GetLogsResponse, error) {
	var res *GetLogsResponse
	var err error
	f := func() (bool, error) {
		res, err = s.GetLogsV2(req)
		if err == nil {
			return res.IsComplete(), nil
		}
		return false, err
	}
	s.getToCompleted(f)
	return res, err
}

// GetLogsToCompletedV3 query logs with [from, to) time range to completed
func (s *LogStore) GetLogsToCompletedV3(req *GetLogRequest) (*GetLogsV3Response, error) {
	var res *GetLogsV3Response
	var err error
	f := func() (bool, error) {
		res, err = s.GetLogsV3(req)
		if err == nil {
			return res.IsComplete(), nil
		}
		return false, err
	}
	s.getToCompleted(f)
	return res, err
}

// GetHistogramsToCompleted query logs with [from, to) time range to completed
func (s *LogStore) GetHistogramsToCompleted(topic string, from int64, to int64, queryExp string) (*GetHistogramsResponse, error) {
	var res *GetHistogramsResponse
	var err error
	f := func() (bool, error) {
		res, err = s.GetHistograms(topic, from, to, queryExp)
		if err == nil {
			return res.IsComplete(), nil
		}
		return false, err
	}
	s.getToCompleted(f)
	return res, err
}

// GetLogsV2 query logs with [from, to) time range
func (s *LogStore) GetLogsV2(req *GetLogRequest) (*GetLogsResponse, error) {
	resp, httpRsp, err := s.getLogsV3Internal(req)
	if err != nil {
		return nil, err
	}
	return toLogRespV2(resp, httpRsp.Header)
}

func toLogRespV2(v3Resp *GetLogsV3Response, respHeader http.Header) (*GetLogsResponse, error) {
	queryInfo, err := v3Resp.Meta.constructQueryInfo()
	if err != nil {
		return nil, fmt.Errorf("fail to construct x-log-query-info: %w", err)
	}
	convertToLogRespV2Header(v3Resp, respHeader, queryInfo)
	return &GetLogsResponse{
		Logs:     v3Resp.Logs,
		Progress: v3Resp.Meta.Progress,
		Count:    v3Resp.Meta.Count,
		HasSQL:   v3Resp.Meta.HasSQL,
		Contents: queryInfo,
		Header:   respHeader,
	}, nil
}

func convertToLogRespV2Header(v3Resp *GetLogsV3Response, header http.Header, queryInfo string) {
	header.Add(GetLogsCountHeader, strconv.FormatInt(v3Resp.Meta.Count, 10))
	header.Add(ProcessedRows, strconv.FormatInt(v3Resp.Meta.ProcessedRows, 10))
	header.Add(ProgressHeader, v3Resp.Meta.Progress)
	header.Add(ProcessedBytes, strconv.FormatInt(v3Resp.Meta.ProcessedBytes, 10))
	header.Add(ElapsedMillisecond, strconv.FormatInt(v3Resp.Meta.ElapsedMillisecond, 10))
	header.Add(HasSQLHeader, strconv.FormatBool(v3Resp.Meta.HasSQL))
	header.Add(TelemetryType, v3Resp.Meta.TelemetryType)
	header.Add(WhereQuery, v3Resp.Meta.WhereQuery)
	header.Add(AggQuery, v3Resp.Meta.AggQuery)
	header.Add(CpuSec, strconv.FormatFloat(v3Resp.Meta.CpuSec, 'E', -1, 64))
	header.Add(CpuCores, strconv.FormatFloat(v3Resp.Meta.CpuCores, 'E', -1, 64))
	header.Add(PowerSql, strconv.FormatBool(v3Resp.Meta.PowerSql))
	header.Add(InsertedSql, v3Resp.Meta.InsertedSql)
	header.Add(GetLogsQueryInfo, queryInfo)
}

// GetLogsV3 query logs with [from, to) time range
func (s *LogStore) GetLogsV3(req *GetLogRequest) (*GetLogsV3Response, error) {
	result, _, err := s.getLogsV3Internal(req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LogStore) getLogsV3Internal(req *GetLogRequest) (*GetLogsV3Response, *http.Response, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}
	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(reqBody)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "lz4",
	}
	uri := fmt.Sprintf("/logstores/%s/logs", s.Name)
	r, err := request(s.project, "POST", uri, h, reqBody)
	if err != nil {
		return nil, nil, NewClientError(err)
	}
	defer r.Body.Close()

	respBody, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(respBody, err); jErr != nil {
			return nil, nil, NewBadResponseError(string(respBody), r.Header, r.StatusCode)
		}
		return nil, nil, err
	}
	if _, ok := r.Header[BodyRawSize]; ok {
		if len(r.Header[BodyRawSize]) > 0 {
			bodyRawSize, err := strconv.ParseInt(r.Header[BodyRawSize][0], 10, 64)
			if err != nil {
				return nil, nil, NewBadResponseError(string(respBody), r.Header, r.StatusCode)
			}
			out := make([]byte, bodyRawSize)
			if bodyRawSize != 0 {
				len, err := lz4.UncompressBlock(respBody, out)
				if err != nil || int64(len) != bodyRawSize {
					return nil, nil, NewBadResponseError(string(respBody), r.Header, r.StatusCode)
				}
			}
			respBody = out
		}
	}
	var result GetLogsV3Response
	if err = json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, NewBadResponseError(string(respBody), r.Header, r.StatusCode)
	}
	return &result, r, nil
}

// GetContextLogs ...
func (s *LogStore) GetContextLogs(backLines int32, forwardLines int32,
	packID string, packMeta string) (*GetContextLogsResponse, error) {

	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Accept":            "application/json",
	}

	urlVal := url.Values{}
	urlVal.Add("type", "context_log")
	urlVal.Add("back_lines", strconv.Itoa(int(backLines)))
	urlVal.Add("forward_lines", strconv.Itoa(int(forwardLines)))
	urlVal.Add("pack_id", packID)
	urlVal.Add("pack_meta", packMeta)

	uri := fmt.Sprintf("/logstores/%s?%s", s.Name, urlVal.Encode())
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return nil, NewClientError(err)

	}
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		if jErr := json.Unmarshal(body, err); jErr != nil {
			return nil, NewBadResponseError(string(body), r.Header, r.StatusCode)

		}
		return nil, err

	}

	resp := GetContextLogsResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, NewBadResponseError(string(body), r.Header, r.StatusCode)

	}
	return &resp, nil
}

// CreateIndex ...
func (s *LogStore) CreateIndex(index Index) error {
	body, err := json.Marshal(index)
	if err != nil {
		return err
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "POST", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

// CreateIndexString ...
func (s *LogStore) CreateIndexString(indexStr string) error {
	body := []byte(indexStr)
	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "POST", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

// UpdateIndex ...
func (s *LogStore) UpdateIndex(index Index) error {
	body, err := json.Marshal(index)
	if err != nil {
		return err
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "PUT", uri, h, body)
	if r != nil {
		r.Body.Close()
	}
	return err
}

// UpdateIndexString ...
func (s *LogStore) UpdateIndexString(indexStr string) error {
	body := []byte(indexStr)

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "PUT", uri, h, body)
	if r != nil {
		r.Body.Close()
	}
	return err
}

// DeleteIndex ...
func (s *LogStore) DeleteIndex() error {

	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "DELETE", uri, h, nil)
	if r != nil {
		r.Body.Close()
	}
	return err
}

// GetIndex ...
func (s *LogStore) GetIndex() (*Index, error) {
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": "0",
		"Accept-Encoding":   "deflate",
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}
	index := &Index{}
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(data, index)
	if err != nil {
		return nil, NewBadResponseError(string(data), r.Header, r.StatusCode)
	}

	return index, nil
}

// GetIndexString ...
func (s *LogStore) GetIndexString() (string, error) {
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": "0",
		"Accept-Encoding":   "deflate",
	}

	uri := fmt.Sprintf("/logstores/%s/index", s.Name)
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	return string(data), err
}

// CheckIndexExist check index exist or not
func (s *LogStore) CheckIndexExist() (bool, error) {
	if _, err := s.GetIndex(); err != nil {
		if slsErr, ok := err.(*Error); ok {
			if slsErr.Code == "IndexConfigNotExist" {
				return false, nil
			}
			return false, slsErr
		}
		return false, err
	}

	return true, nil
}

func (s *LogStore) GetMeteringMode() (*GetMeteringModeResponse, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}
	uri := fmt.Sprintf("/logstores/%s/meteringmode", s.Name)
	r, err := request(s.project, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, NewBadResponseError("", r.Header, r.StatusCode)
	}
	res := GetMeteringModeResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, NewBadResponseError(string(data), r.Header, r.StatusCode)

	}
	return &res, nil

}

func (s *LogStore) UpdateMeteringMode(meteringMode string) error {

	body := map[string]string{
		"meteringMode": meteringMode,
	}
	uri := fmt.Sprintf("/logstores/%s/meteringmode", s.Name)
	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cant marshal body:%w", err)
	}

	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": strconv.Itoa(len(requestBody)),
	}

	r, err := request(s.project, "PUT", uri, h, requestBody)

	if err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}
