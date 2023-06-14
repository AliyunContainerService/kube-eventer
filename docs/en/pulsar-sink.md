### Pulsar sink

To use the Pulsar sink add the following flag:

    --sink=Pulsar:<?<OPTIONS>>

* `serviceurl` - Pulsar's broker or proxy.
* `eventstopic` - Pulsar's topic for events.
* `token` - Pulsar's JWT token, If you enable [JWT](https://pulsar.apache.org/docs/next/security-jwt/).

For example,

    --sink=pulsar:?serviceurl=pulsar://127.0.0.1:6650&eventstopic=persistent://public/default/event&token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9
    or
    --sink=pulsar:?serviceurl=pulsar://127.0.0.1:6650&eventstopic=persistent://public/default/event