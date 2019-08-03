### MySQL sink

*This sink supports Mysql versions v5.7 and above*.
To use the Mysql sink add the following flag:

	--sink=mysql:?<MYSQL_JDBC_URL>

For example:

    --sink="mysql:?root:transwarp@tcp(172.16.180.132:3306)/kube_event?charset=utf8"