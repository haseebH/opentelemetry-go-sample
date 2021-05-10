package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"log"
	"os"
	"sync"
	"time"
)

func handleErr1(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider1() func() {
	ctx := context.Background()

	otelAgentAddr, ok := os.LookupEnv("OTEL_AGENT_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:30080"
	}

	exp, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(otelAgentAddr),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	))
	handleErr1(err, "failed to create exporter")

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("subscriber"),
		),
	)
	handleErr1(err, "failed to create resource")

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	_ = bsp
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSyncer(exp),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithCollectPeriod(7*time.Second),
		controller.WithExporter(exp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)
	global.SetMeterProvider(cont.MeterProvider())

	handleErr1(cont.Start(context.Background()), "failed to start metric controller")

	return func() {
		handleErr1(tracerProvider.Shutdown(ctx), "failed to shutdown provider")
		handleErr1(cont.Stop(context.Background()), "failed to stop metrics controller") // pushes any last exports to the receiver
		handleErr1(exp.Shutdown(ctx), "failed to stop exporter")
	}
}
func main() {
	shutdown := initProvider1()
	defer shutdown()
	ctx := context.Background()
	proj := "cloudplex-nextgen-test-project"
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	// Pull messages via the subscription.
	if err := pullMsgs(client); err != nil {
		log.Fatal(err)
	}
}

func pullMsgs(client *pubsub.Client) error {
	ctx := context.Background()

	var mu sync.Mutex
	received := 0
	sub := client.Subscription("htesting-sub")
	cctx, cancel := context.WithCancel(ctx)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		fmt.Printf("Got message: %q\n", string(msg.Data))
		mu.Lock()
		defer mu.Unlock()
		traceId, err := trace.TraceIDFromHex(msg.Attributes["traceId"])
		if err != nil {
			fmt.Println("err", err)
			return
		}
		spanId, err := trace.SpanIDFromHex(msg.Attributes["spanId"])
		if err != nil {
			fmt.Println("err", err)
			return
		}

		spanCtx := trace.SpanContextFromContext(context.Background()).WithTraceID(traceId).WithSpanID(spanId)
		ctx = baggage.ContextWithValues(ctx)

		commonLabels := []attribute.KeyValue{
			attribute.String("method", "repl"),
			attribute.String("client", "cli"),
		}
		ctx, span := otel.Tracer("subscriber").Start(
			trace.ContextWithRemoteSpanContext(ctx, spanCtx),
			"main-server",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(commonLabels...),
		)

		span1 := trace.SpanFromContext(ctx)

		span1.SetName("subscriber")
		span1.AddEvent("subscriber")
		span1.SetAttributes(attribute.Key("subscriber").String("value"))

		received++
		if received == 10 {
			cancel()
		}
		span.End()
		span1.End()
	})
	if err != nil {
		return err
	}
	return nil
}
