### MongoDB

*This sink supports MongoDB 2.6 and higher*.
To use the MongoDB sink add the following flag:
```
    --sink=mongo:?<MONGO_URI>
```


For example,
```
    --sink=mongo:?mongodb://root:123456@mongo-replset-0-0.dba-c1.example.com:30694,mongo-replset-0-1.dba-c1.example.com:32761,mongo-replset-0-2.dba-c1.example.com:31958/?authSource=admin
```