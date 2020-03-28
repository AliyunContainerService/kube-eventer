### webhook sink

*This sink supports Webhook(will send events to the Webhook URL with `POST` http method and use json serialized events as body)*.
To use the webhook sink add the following flag:

	--sink=webhook:<webhookURL>?sinkRetryMaxTimes=<sinkRetryMaxTimes>&sinkRetryPeriod=<sinkRetryPeriod>&sinkRetryJitterFactor=<sinkRetryJitterFactor>


The following options are available:
* `webhookURL` - Your Webhook URL
* `sinkRetryMaxTimes` - The max retry times when send data to webhook failed(default: 3)
* `sinkRetryPeriod` - The retry period when send data to webhook failed(default: 1000ms)
* `sinkRetryJitterFactor` - The retry jitter factor when send data to webhook failed(default: 1.0)

For example:

    --sink=webhook:https://webhook.example.com/
    --sink=webhook:https://webhook.example.com/?foo=bar&abc=233
or

    --sink=webhook:https://webhook.example.com/?sinkRetryMaxTimes=5&sinkRetryPeriod=10ms&sinkRetryJitterFactor=1.2
    --sink=webhook:https://webhook.example.com/?foo=bar&abc=233&sinkRetryMaxTimes=5&sinkRetryPeriod=10ms&sinkRetryJitterFactor=1.2
