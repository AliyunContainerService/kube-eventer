### MySQL sink

*This sink supports Mysql versions v5.7 and above*.
To use the Mysql sink add the following flag:

	--sink=mysql:?<MYSQL_JDBC_URL>

*First build table*

```
drop database if exists kube_event ;
create database kube_event;
use kube_event;


create table kube_event
(
    id               bigint(20)   not null auto_increment primary key comment 'key',
    name             varchar(64)  not null default '' comment 'name',
    namespace        varchar(64)  not null default '' comment 'namespace',
    event_id         varchar(64)  not null default '' comment 'event_id',
    type             varchar(64)  not null default '' comment 'type',
    reason           varchar(64)  not null default '' comment 'reason',
    message          text  not null  comment 'message' ,
    kind             varchar(64)  not null default '' comment 'kind' ,
    first_occurrence_time   varchar(64)    not null default '' comment '创建时间',
    last_occurrence_time    varchar(64)    not null default '' comment '更新时间',
    unique index event_id_index (event_id)
) ENGINE = InnoDB default CHARSET = utf8 comment ='Event表';
```

For example:

    --sink="mysql:?root:transwarp@tcp(172.16.180.132:3306)/kube_event?charset=utf8"