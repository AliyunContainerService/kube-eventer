### Riemann sink
To use the Riemann sink add the following flag:

	--sink="riemann:<RIEMANN_SERVER_URL>[?<OPTIONS>]"

The following options are available:

* `ttl` - TTL for writing to Riemann. Default: `60 seconds`
* `state` - The event state. Default: `""`
* `tags` - Default. `heapster`
* `batchsize` - The Riemann sink sends batch of events. The default size is `1000`

For example,

    --sink=riemann:http://localhost:5555?ttl=120&state=ok&tags=foobar&batchsize=150
