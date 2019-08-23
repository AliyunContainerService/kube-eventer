create database if not exists kube_event;

use kube_event;
drop table kube_event;
CREATE TABLE `kube_event` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `namespace` varchar(255) DEFAULT '',
  `kind` varchar(255) DEFAULT '',
  `name` varchar(255) DEFAULT '',
  `type` varchar(255) DEFAULT '',
  `reason` varchar(255) DEFAULT '',
  `message` varchar(255) DEFAULT '',
  `event_id` varchar(255) DEFAULT '',
  `first_occurrence_time` varchar(255) DEFAULT '',
  `last_occurrence_time` varchar(255) DEFAULT '',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4


