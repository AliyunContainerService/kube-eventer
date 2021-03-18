### Honeycomb sink

To use the Honeycomb sink add the following flag:

    --sink="honeycomb:<?<OPTIONS>>"

Options can be set in query string, like this:

* `dataset` - Honeycomb Dataset to which to publish metrics/events
* `writekey` - Honeycomb Write Key for your account
* `apihost` - Option to send metrics to a different host (default: https://api.honeycomb.io) (optional)

For example,

    --sink="honeycomb:?dataset=mydataset&writekey=secretwritekey"
