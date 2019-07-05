module github.com/AliyunContainerService/kube-eventer

go 1.12

require (
	github.com/Shopify/sarama v1.22.1
	github.com/denverdino/aliyungo v0.0.0-20190410085603-611ead8a6fed
	github.com/golang/protobuf v1.3.1
	github.com/google/cadvisor v0.33.1
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/influxdata/influxdb v1.7.7
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mailru/easyjson v0.0.0-20190614124828-94de47d64c63 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v1.0.0
	github.com/riemann/riemann-go-client v0.4.0
	github.com/smartystreets/go-aws-auth v0.0.0-20180515143844-0c1422d1fdb9
	github.com/smartystreets/gunit v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8 // indirect
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/olivere/elastic.v3 v3.0.75
	gopkg.in/olivere/elastic.v5 v5.0.81
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190627205229-acea843d18eb
	k8s.io/apimachinery v0.0.0-20190627205106-bc5732d141a8
	k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.1
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190606204050-af9c91bd2759
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503
)
