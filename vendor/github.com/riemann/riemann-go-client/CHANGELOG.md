# Change Log

## Release v0.4.0

### Bugfixes

- fix warning "literal copies lock value from config" for tls client

### Features

- Force TLS Handshack in Connect() for the TLS client
- Add support for uint types in Metric

## Release v0.3.0

### Bugfixes

- Fix a bug which produced a panic when an event without Metric was serialized to protobuf.

## Release v0.2.0

### Breaking change

- Event.Time is no longer an int64 but a time.Time

### Features

- Support microsecond resolution

## Release v0.1.0

### Features

- Sending events, queries.
- Support TCP, UDP, TLS clients.
- Do not support microseconds time resolution
