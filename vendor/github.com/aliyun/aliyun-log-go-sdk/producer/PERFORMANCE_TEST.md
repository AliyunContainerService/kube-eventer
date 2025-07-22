## 测试环境:

### ECS虚拟机

实例环境: ecs.c5.xlarge
cpu: 4core
内存: 8Gib
操作系统: CentOS 7.6 64位

## GOLANG 版本:

go version go1.12.4 linux/amd64 

## 日志样例

测试中使用的日志包含 8 个键值对以及 topic、source 两个字段。为了模拟数据的随机性，我们给每个字段值追加了一个随机后缀。其中，topic 后缀取值范围是 [0, 5)，source 后缀取值范围是 [0, 10)，其余 8 个键值对后缀取值范围是 [0, 单线程发送次数)。单条日志大小约为 550 字节，格式如下：

```
__topic__:  topic-<suffix>  
__source__:  source-<suffix>
content_key_1:  1abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>
content_key_2:  2abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>
content_key_3:  3abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>
content_key_4:  4abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>
content_key_5:  5abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>  
content_key_6:  6abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>  
content_key_7:  7abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>  
content_key_8:  8abcdefghijklmnopqrstuvwxyz!@#$%^&*()_0123456789-<suffix>  
```

### Project & Logstore

- Project：在 ECS 所在 region 创建目标 project 并通过 VPC 网络服务入口进行访问。
- Logstore：在该 project 下创建一个分区数为 10 的 logstore（未开启索引），该 logstore 的写入流量最大为 50 MB/s，参阅[数据读写](https://help.aliyun.com/document_detail/92571.html)。

## 测试用例

### 测试程序说明

- ProducerConfig.totalSizeInBytes: 具体用例中调整
- ProducerConfig.maxBatchSizeInBytes: 3 * 1024 * 1024
- ProducerConfig.maxBatchCount：40960
- ProducerConfig.lingerMs：2000
- 调用`Producer.send()`方法的groutine数量：10
- 每个线程发送日志条数：20,000,000
- 发送日志总大小：约 115 GB
- 客户端压缩后大小：约 12 GB
- 发送日志总条数：200,000,000

### 调整使用cpu数量

将 ProducerConfig.totalSizeInBytes 设置为默认值 104,857,600（即 100 MB），设置发送任务使用的groutine数量为默认值50个，通过在程序开始时设置程序使用cpu核心数:runtime.GOMAXPROCS(1)，来观察程序性能。

| cpu数量 | 线程池groutine数量 | 原始数据吞吐量 | 压缩后数据吞吐量 | cpu使用率 | 说明                        |
| ------- | ------------------ | -------------- | ---------------- | --------- | --------------------------- |
| 1       | 50                 | 73.386MB/s     | 7.099MB/s        | 24%       | 未达到10个shard写入能力上限 |
| 2       | 50                 | 136.533MB/s    | 13.141MB/s       | 50%       | 未达到10个shard写入能力上限 |
| 4       | 50                 | 163.84MB/s     | 15.701MB/s       | 84%       | 未达到10个shard写入能力上限 |



## 调整MaxIoWorkerCount



| cpu数量 | 线程池groutine数量 | 原始数据吞吐量 | 压缩后数据吞吐量 | cpu使用率 | 说明                        |
| ------- | ------------------ | -------------- | ---------------- | --------- | --------------------------- |
| 1       | 100                | 69.97MB/s      | 6.656MB/s        | 23%       | 未达到10个shard写入能力上限 |
| 2       | 100                | 131.41MB/s     | 12.62MB/s        | 49%       | 未达到10个shard写入能力上限 |
| 4       | 100                | 162.133MB/s    | 15.701MB/s       | 84%       | 未达到10个shard写入能力上限 |

## 调整totalSizeInBytes

将使用ProducerConfig.ioThreadCount设置为2 (注意这个配置为当前使用cpu核数量),通过调整 ProducerConfig.totalSizeInBytes 观察程序性能。

| TotalSizeInBytes | 线程池groutine数量 | 原始数据吞吐量 | 压缩后数据吞吐量 | cpu使用率 | 说明                        |
| ---------------- | ------------------ | -------------- | ---------------- | --------- | --------------------------- |
| 52,428,800       | 50                 | 34.133MB/s     | 3.3792MB/s       | 15%       | 未达到10个shard写入能力上限 |
| 209,715,200      | 50                 | 136.53MB/s     | 13.141MB/s       | 49%       | 未达到10个shard写入能力上限 |
| 419,430,400      | 50                 | 133.12MB/s     | 12.8MB/s         | 50%       | 未达到10个shard写入能力上限 |

## 总结

1. 增加程序使用cpu数目可以显著提高吞吐量。
2. 在使用cpu数目不变的情况下，增加线程池groutine的数量并不一定会带来性能的提升，相反多开的groutine会提高程序gc的cpu使用率，所以线程池开启groutine的数量应该按照机器的性能去调整一个合适的数值，建议可以使用默认值。
3. 在cpu数目和线程池groutine数目不变的情况下，调整 totalSizeInBytes 对吞吐量影响不够显著，增加 totalSizeInBytes 会造成更多的 CPU 消耗，建议使用默认值。