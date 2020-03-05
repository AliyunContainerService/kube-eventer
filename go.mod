module github.com/AliyunContainerService/kube-eventer

go 1.12

require (
	github.com/Shopify/sarama v1.22.1
	github.com/aws/aws-sdk-go v1.19.6
	github.com/denverdino/aliyungo v0.0.0-20190410085603-611ead8a6fed
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.2
	github.com/google/cadvisor v0.33.1
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/influxdata/influxdb v1.7.7
	github.com/kr/pretty v0.1.0 // indirect
	github.com/olivere/elastic v6.2.23+incompatible // indirect
	github.com/olivere/elastic/v7 v7.0.6
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v1.0.0
	github.com/riemann/riemann-go-client v0.4.0
	github.com/smartystreets/go-aws-auth v0.0.0-20180515143844-0c1422d1fdb9
	github.com/smartystreets/gunit v1.0.0 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/olivere/elastic.v3 v3.0.75
	gopkg.in/olivere/elastic.v5 v5.0.81
	gopkg.in/olivere/elastic.v6 v6.2.23
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.17.3
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190606204050-af9c91bd2759
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503
)
