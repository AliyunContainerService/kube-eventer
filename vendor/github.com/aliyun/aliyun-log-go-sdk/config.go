package sls

const (
	version         = "0.6.0"     // SDK version
	signatureMethod = "hmac-sha1" // Signature method

	// OffsetNewest stands for the log head offset, i.e. the offset that will be
	// assigned to the next message that will be produced to the shard.
	OffsetNewest = "end"
	// OffsetOldest stands for the oldest offset available on the logstore for a
	// shard.
	OffsetOldest = "begin"

	// ProgressHeader stands for the progress header in GetLogs response
	ProgressHeader = "X-Log-Progress"

	ProcessedRows      = "x-log-processed-rows"
	ProcessedBytes     = "x-log-processed-bytes"
	ElapsedMillisecond = "x-log-elapsed-millisecond"
	TelemetryType      = "x-tlm-type"
	WhereQuery         = "x-log-where-query"
	AggQuery           = "x-log-agg-query"
	CpuSec             = "x-log-cpu-sec"
	CpuCores           = "x-log-cpu-cores"
	PowerSql           = "x-log-power-sql"
	InsertedSql        = "x-log-insertedsql"

	// GetLogsCountHeader stands for the count header in GetLogs response
	GetLogsCountHeader = "X-Log-Count"

	// RequestIDHeader stands for the requestID in all response
	RequestIDHeader = "x-log-requestid"

	GetLogsQueryInfo = "X-Log-Query-Info"
	BodyRawSize      = "X-Log-Bodyrawsize"
	HasSQLHeader     = "x-log-has-sql"
	ETLVersion       = 2
	ETLType          = "ETL"
	ETLSinksType     = "AliyunLOG"
)
