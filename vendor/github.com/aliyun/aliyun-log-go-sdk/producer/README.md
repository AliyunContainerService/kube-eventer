# Aliyun LOG Go Producer

Aliyun LOG Go Producer 是一个易于使用且高度可配置的 golang类库，专门为大数据情况下设计的 go语言版本的日志发送利器。

## 功能特点

1. 线程安全 -  producer 内所有的方法以及暴露的接口都是线程安全的。
2. 异步发送 - 调用send方法后回立即返回，日志将会被传递到io线程中异步发送，不阻塞用户发送日志操作。
3. 失败重试 - 用户可以通过设置初始化的参数Retries来指定日志发送失败的次数，超过重试次数将被投递到失败队列。
4. 优雅关闭 - 用户调用关闭方法进行关闭时，producer 会将所有其缓存的数据进行发送，防止日志丢失，关闭分为有限关闭和安全关闭，详细的区别会在下文中列出。
5. 本地调试 - 可通过配置支持将日志内容输出到本地或控制台，并支持轮转、日志数、轮转大小设置。
6. 高性能 - go版本的producer 基于go 语言特性进行开发，go的goroutine在并发多任务处理能力上有着与生俱来的优势。所以producer 对每一个可发送的任务会开启单独的groutine去执行发送任务，相对比直接使用cpu线程处理，对系统性能消耗更小，效率更高。
7. 使用简单 - 在整个使用过程中，producer给提供了3个方法，start,send和close,用户启动producer 以后只需要调用send方法即可发送日志，producer 提供不同的send 的方法，用来满足用户的发送需求。
8. 结果可控制 - 用户可以自己实现producer 提供的CallBack 接口，里面包含日志发送成功和失败后调用的方法，可以自行在CallBack接口写日志发送结果处理逻辑。



# **安装**

1.在$GOPATH/src/github.com目录下创建aliyun目录，

2.克隆代码到创建的aliyun目录下 (源码地址：[aliyun-go-consumer-library](https://github.com/aliyun/aliyun-log-go-sdk))。

```shell
git clone https://github.com/aliyun/aliyun-log-go-sdk.git
```

3.安装google提供的序列化工具包到自己的GOPATH目录下面

```shell
go get github.com/gogo/protobuf/proto
```

# 使用步骤

**1.配置ProducerConfig**

ProducerConfig 是提供给用户的配置类，用于配制发送策略，您可以根据不同的需求设置不同的值，具体的参数含义如文章尾producer配置详解所示。
producer同时提供了简单的使用代码simple:([producer_simple_demo](https://github.com/aliyun/aliyun-log-go-sdk/blob/master/example/producer/producer_simple_demo.go))

**2.启动producer进程**

```go langgo l
producerConfig := producer.GetDefaultProducerConfig()
producerConfig.Endpoint = os.Getenv("Endpoint")
provider := sls.NewStaticCredentailsProvider(os.Getenv("AccessKeyID"), os.Getenv("AccessKeySecret"), "")
producerConfig.CredentialsProvider = provider
producerInstance:=producer.InitProducer(producerConfig)
ch := make(chan os.Signal)
signal.Notify(ch, os.Kill, os.Interrupt)
producerInstance.Start() // 启动producer实例
```

当调用producerInstance.Start()这个函数会开启一个groutine去监听producer中是否有日志写入以及符合发送条件的日志组，将符合发送条件的日志组发送到服务端LogHub中。

**3.调用Send方法发送日志**

```go
for i:=0;i<10000;i++ {
   // GenerateLog  is producer's function for generating SLS format logs
   log := producer.GenerateLog(uint32(time.Now().Unix()), map[string]string{"content": "test", "content2": fmt.Sprintf("%v",i)})
   err := producerInstance.SendLog("projectName", "logstorName", "127.0.0.1","topic",log)
   if err != nil {
      fmt.Println(err)
   }
}
```

producer中提供了GenerateLog方法供用户生成可以投递到LogHub的日志实例。GenerateLog方法中使用了proto去对数据进行了序列，效率较低，推荐用户使用原生的sls.Log接口去创建日志，该方法仅供测试调试使用。

**4.关闭producer**

producer提供了两种关闭模式，分为有限关闭和安全关闭，安全关闭会等待producer中缓存的所有的数据全部发送完成以后在关闭producer，有限关闭会接收用户传递的一个参数值，时间单位为秒，当开始关闭producer的时候开始计时，超过传递的设定值还未能完全关闭producer的话会强制退出producer，此时可能会有部分数据未被成功发送而丢失。

```go
producerInstance.Close(60) // 有限关闭，传递int值，参数值需为正整数，单位为秒
producerInstance.SafeClose()// 安全关闭
```

**5.获取发送结果**

producer 每次向服务端发送请求都是异步的，所以需要用户实现callback接口，去获得每次发送的结果。

实现Callback接口需要实现其中的Success()方法和Fail()方法，两个方法分别会在日志发送成功或失败的时候去调用，两个方法会都会接收一个Result 实例，用户可以根据Result实例在CallBack回调方法中去获得每次日志发送的结果。下面写了一个简单的使用样例。

```go
type Callback struct{

}
func(callback *Callback)Success(result *producer.Result){
   attemptList := result.GetReservedAttempts() // 遍历获得所有的发送记录
   for _,attempt:=range attemptList{
      fmt.Println(attempt)
   }
}

func(callback *Callback)Fail(result *producer.Result){
   fmt.Println(result.IsSuccessful()) // 获得发送日志是否成功
   fmt.Println(result.GetErrorCode()) // 获得最后一次发送失败错误码
   fmt.Println(result.GetErrorMessage()) // 获得最后一次发送失败信息
   fmt.Println(result.GetReservedAttempts()) // 获得producerBatch 每次尝试被发送的信息
   fmt.Println(result.GetRequestId()) // 获得最后一次发送失败请求Id
   fmt.Println(result.GetTimeStampMs()) // 获得最后一次发送失败请求时间
}
```

用户可以根据自己的需求调用Result实例提供的方法来获取日志发送结果信息，日志每次尝试被发送都会生成attempt信息，默认会保留11次，这个数字可以根据配置参数MaxReservedAttempts进行修改。



## **producer配置详解**

| 参数                | 类型        | 描述                                                                                                                                                                                                                    |
| ------------------- |-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| TotalSizeLnBytes    | Int64     | 单个 producer 实例能缓存的日志大小上限，默认为 100MB。                                                                                                                                                                                   |
| MaxIoWorkerCount    | Int64     | 单个producer能并发的最多groutine的数量，默认为50，该参数用户可以根据自己实际服务器的性能去配置。                                                                                                                                                             |
| MaxBlockSec         | Int       | 如果 producer 可用空间不足，调用者在 send 方法上的最大阻塞时间，默认为 60 秒。<br/>如果超过这个时间后所需空间仍无法得到满足，send 方法会抛出TimeoutException。如果将该值设为0，当所需空间无法得到满足时，send 方法会立即抛出 TimeoutException。如果您希望 send 方法一直阻塞直到所需空间得到满足，可将该值设为负数。                       |
| MaxBatchSize        | Int64     | 当一个 ProducerBatch 中缓存的日志大小大于等于 batchSizeThresholdInBytes 时，该 batch 将被发送，默认为 512 KB，最大可设置成 5MB。                                                                                                                        |
| MaxBatchCount       | Int       | 当一个 ProducerBatch 中缓存的日志条数大于等于 batchCountThreshold 时，该 batch 将被发送，默认为 4096，最大可设置成 40960。                                                                                                                              |
| LingerMs            | Int64     | 一个 ProducerBatch 从创建到可发送的逗留时间，默认为 2 秒，最小可设置成 100 毫秒。                                                                                                                                                                  |
| Retries             | Int       | 如果某个 ProducerBatch 首次发送失败，能够对其重试的次数，默认为 10 次。<br/>如果 retries 小于等于 0，该 ProducerBatch 首次发送失败后将直接进入失败队列。                                                                                                                 |
| MaxReservedAttempts | Int       | 每个 ProducerBatch 每次被尝试发送都对应着一个 Attemp，此参数用来控制返回给用户的 attempt 个数，默认只保留最近的 11 次 attempt 信息。<br/>该参数越大能让您追溯更多的信息，但同时也会消耗更多的内存。                                                                                            |
| BaseRetryBackoffMs  | Int64     | 首次重试的退避时间，默认为 100 毫秒。 Producer 采样指数退避算法，第 N 次重试的计划等待时间为 baseRetryBackoffMs * 2^(N-1)。                                                                                                                                 |
| MaxRetryBackoffMs   | Int64     | 重试的最大退避时间，默认为 50 秒。                                                                                                                                                                                                   |
| AdjustShargHash     | Bool      | 如果调用 send 方法时指定了 shardHash，该参数用于控制是否需要对其进行调整，默认为 true。                                                                                                                                                                |
| Buckets             | Int       | 当且仅当 adjustShardHash 为 true 时，该参数才生效。此时，producer 会自动将 shardHash 重新分组，分组数量为 buckets。<br/>如果两条数据的 shardHash 不同，它们是无法合并到一起发送的，会降低 producer 吞吐量。将 shardHash 重新分组后，能让数据有更多地机会被批量发送。该参数的取值范围是 [1, 256]，且必须是 2 的整数次幂，默认为 64。 |
| AllowLogLevel       | String    | 设置日志输出级别，默认值是Info,consumer中一共有4种日志输出级别，分别为debug,info,warn和error。                                                                                                                                                      |
| LogFileName         | String    | 日志文件输出路径，不设置的话默认输出到stdout。                                                                                                                                                                                            |
| IsJsonType          | Bool      | 是否格式化文件输出格式，默认为false。                                                                                                                                                                                                 |
| LogMaxSize          | Int       | 单个日志存储数量，默认为10M。                                                                                                                                                                                                      |
| LogMaxBackups       | Int       | 日志轮转数量，默认为10。                                                                                                                                                                                                         |
| LogCompass          | Bool      | 是否使用gzip 压缩日志，默认为false。                                                                                                                                                                                               |
| Endpoint            | String    | 服务入口，关于如何确定project对应的服务入口可参考文章[服务入口](https://help.aliyun.com/document_detail/29008.html?spm=a2c4e.11153940.blogcont682761.14.446e7720gs96LB)。                                                                         |
| AccessKeyID         | String    | 账户的AK id。                                                                                                                                                                                                             |
| AccessKeySecret     | String    | 账户的AK 密钥。                                                                                                                                                                                                             |
|CredentialsProvider| Interface | 可选，可自定义CredentialsProvider，来提供动态的 AccessKeyId/AccessKeySecret/StsToken，该接口应当缓存 AK，且必须线程安全                                                                                                                             |
| NoRetryStatusCodeList  | []int     | 用户配置的不需要重试的错误码列表，当发送日志失败时返回的错误码在列表中，则不会重试。默认包含400，404两个值。                                                                                                                                                             |
| UpdateStsToken      | Func      | 函数类型，该函数内去实现自己的获取ststoken 的逻辑，producer 会自动刷新ststoken并放入client 当中。                                                                                                                                                     
| StsTokenShutDown    | channel   | 关闭ststoken 自动刷新的通讯信道，当该信道关闭时，不再自动刷新ststoken值。当producer关闭的时候，该参数不为nil值，则会主动调用close去关闭该信道停止ststoken的自动刷新。                                                                                                               |
| Region              | String    | 日志服务的区域，当签名版本使用 AuthV4 时必选。 例如cn-hangzhou。                                                                                                                                                                            |
| AuthVersion         | String    | 使用的签名版本，可选枚举值为 AuthV1， AuthV4。AuthV4 签名示例可参考程序 [producer_test.go](producer_test.go)。                                                                                                                                  |
| UseMetricStoreURL         | bool      | 使用 Metricstore地址进行发送日志,可以提升大基数时间线下的查询性能。                                                                                                                                                                              |

## 关于性能

- [性能测试报告](https://github.com/aliyun/aliyun-log-go-sdk/blob/master/producer/PERFORMANCE_TEST.md)



## 问题反馈

如果您在使用过程中遇到了问题，可以创建 [GitHub Issue](<https://github.com/aliyun/aliyun-log-go-sdk>)或者前往阿里云支持中心[提交工单](https://workorder.console.aliyun.com/#/ticket/createIndex)。



