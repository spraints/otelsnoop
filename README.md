# Spy on your otlp exporter

## Synopsis

```
$ git clone https://github.com/spraints/otelsnoop
$ go run .
```

###

Runs an http server on 127.0.0.1:8360 and dumps traces that are posted. This is
useful if you set up, for example, the Ruby opentelemetry tracer something like
this:

```
$ export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://localhost:8360/traces/otlp/v0.9"
$ ruby run-my-app.rb
```
