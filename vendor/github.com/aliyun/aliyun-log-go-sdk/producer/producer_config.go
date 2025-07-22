package producer

import (
	"net/http"
	"sync"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

const Delimiter = "|"

type UpdateStsTokenFunc = func() (accessKeyID, accessKeySecret, securityToken string, expireTime time.Time, err error)

type ProducerConfig struct {
	TotalSizeLnBytes      int64
	MaxIoWorkerCount      int64
	MaxBlockSec           int
	MaxBatchSize          int64
	MaxBatchCount         int
	LingerMs              int64
	Retries               int
	MaxReservedAttempts   int
	BaseRetryBackoffMs    int64
	MaxRetryBackoffMs     int64
	AdjustShargHash       bool
	Buckets               int
	AllowLogLevel         string
	LogFileName           string
	IsJsonType            bool
	LogMaxSize            int
	LogMaxBackups         int
	LogCompress           bool
	Endpoint              string
	NoRetryStatusCodeList []int
	HTTPClient            *http.Client
	UserAgent             string
	LogTags               []*sls.LogTag
	GeneratePackId        bool
	CredentialsProvider   sls.CredentialsProvider
	UseMetricStoreURL     bool

	packLock   sync.Mutex
	packPrefix string
	packNumber int64

	// Deprecated: use CredentialsProvider and UpdateFuncProviderAdapter instead.
	//
	// Example:
	//   provider := sls.NewUpdateFuncProviderAdapter(updateStsTokenFunc)
	//   config := &ProducerConfig{
	//			CredentialsProvider: provider,
	//   }
	UpdateStsToken   UpdateStsTokenFunc
	StsTokenShutDown chan struct{}
	AccessKeyID      string // Deprecated: use CredentialsProvider instead
	AccessKeySecret  string // Deprecated: use CredentialsProvider instead
	Region           string
	AuthVersion      sls.AuthVersionType
	CompressType     int // only work for logstore now
}

func GetDefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		TotalSizeLnBytes:      100 * 1024 * 1024,
		MaxIoWorkerCount:      50,
		MaxBlockSec:           60,
		MaxBatchSize:          512 * 1024,
		LingerMs:              2000,
		Retries:               10,
		MaxReservedAttempts:   11,
		BaseRetryBackoffMs:    100,
		MaxRetryBackoffMs:     50 * 1000,
		AdjustShargHash:       true,
		Buckets:               64,
		MaxBatchCount:         4096,
		NoRetryStatusCodeList: []int{400, 404},
		CompressType:          sls.Compress_LZ4,
	}
}
