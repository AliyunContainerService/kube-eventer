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
    id               bigint(20)   not null auto_increment primary key comment 'event primary key',
    name             varchar(64)  not null default '' comment 'evenet name',
    namespace        varchar(64)  not null default '' comment 'evenet namespace',
    event_id         varchar(64)  not null default '' comment 'event_id',
    type             varchar(64)  not null default '' comment 'event type Warning or Normal',
    reason           varchar(64)  not null default '' comment 'event reason',
    message          text  not null  comment 'event message' ,
    kind             varchar(64)  not null default '' comment 'event kind' ,
    first_occurrence_time   varchar(64)    not null default '' comment 'event first occurrence time',
    last_occurrence_time    varchar(64)    not null default '' comment 'event last occurrence time',
    unique index event_id_index (event_id)
) ENGINE = InnoDB default CHARSET = utf8 comment ='Event info tables';
```

For example:

    --sink="mysql:?root:transwarp@tcp(172.16.180.132:3306)/kube_event?charset=utf8"