### RabbitMQ sink

To use the rabbitMQ sink add the following flag:

    --sink=rabbitmq:<?<OPTIONS>>

The RabbitMQ sink is used to configure the integration with RabbitMQ for event streaming.
This sink allows you to specify RabbitMQ server details,
such as host, port, username, password, and event topic.

To use the RabbitMQ sink, add the following flag to your command line or configuration file:

* `host`: RabbitMQ server host address. (e.g., localhost)
* `port`: RabbitMQ server port. (e.g., 5672)
* `username`: RabbitMQ username.
* `password`: RabbitMQ password.
* `eventtopic`: RabbitMQ's topic for events. 

For example,

    --sink=rabbitmq:?host=localhost&port=5672&username=username&password=password&eventtopic=testtopic
    or
    --sink=rabbitmq:?host=localhost&port=5672&username=username&password=password&metricsTopic=testtopic