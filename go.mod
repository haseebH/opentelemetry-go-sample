module opentelemetry-go-sample

go 1.15

require (
	cloud.google.com/go/pubsub v1.10.3
	github.com/gin-gonic/gin v1.7.1
	github.com/go-redis/redis/extra/redisotel/v8 v8.8.2
	github.com/go-redis/redis/v8 v8.8.2
	go.opentelemetry.io/contrib v0.20.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.20.0
	go.opentelemetry.io/otel/metric v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/sdk/metric v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
)
