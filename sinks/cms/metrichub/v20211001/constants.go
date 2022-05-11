package metrichub

const (
	ContentMd5          = "Content-MD5"
	UserAgent           = "User-Agent"
	UserAgentValue      = "cms-go-sdk-v-1.0"
	ContentType         = "Content-Type"
	ContentJson         = "application/json"
	Date                = "Date"
	XCms                = "x-cms-"
	XAcs                = "x-acs-"
	XCmsApiVersion      = XCms + "api-version"
	XCmsApiVersionValue = "1.0"
	XCmsSignature       = XCms + "signature"
	XCmsSignatureMethod = "hmac-sha1"
	XCmsIp              = XCms + "ip"
	XCmsCallerType      = XCms + "caller-type"
	XCmsCallerToken     = "token"
	XCmsSecurityToken   = XCms + "security-token"
	Authorization       = "Authorization"
)
