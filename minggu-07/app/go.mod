module app

go 1.21

require (
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.4.0
	go.opentelemetry.io/otel v1.22.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.22.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.22.0
	go.opentelemetry.io/otel/sdk v1.22.0
	go.opentelemetry.io/otel/semconv/v1.24.0 v1.22.0
)