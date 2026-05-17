package observability

import (
	"context"
	"log/slog"
	"os"
	"testing"

	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TestMain sets up a real (in-process) tracer provider and W3C propagator so that
// span context injection/extraction work correctly in all tests in this package.
func TestMain(m *testing.M) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	if err := InitMetrics(otel.GetMeterProvider().Meter("test")); err != nil {
		panic("failed to init test metrics: " + err.Error())
	}
	os.Exit(m.Run())
}

// --- sqs_propagation ---

func TestInjectTraceToSQS_NoActiveSpan(t *testing.T) {
	// With a fresh background context (no active span), inject should produce
	// no traceparent attribute.
	attrs := InjectTraceToSQS(context.Background())
	if _, ok := attrs["traceparent"]; ok {
		t.Error("expected no traceparent for context without active span")
	}
}

func TestInjectTraceToSQS_WithActiveSpan(t *testing.T) {
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	attrs := InjectTraceToSQS(ctx)
	if _, ok := attrs["traceparent"]; !ok {
		t.Error("expected traceparent attribute when active span is present")
	}
	v, ok := attrs["traceparent"]
	if !ok || v.StringValue == nil || *v.StringValue == "" {
		t.Error("traceparent attribute must have a non-empty string value")
	}
}

func TestExtractSpanLinkFromSQS_NoAttributes(t *testing.T) {
	msg := sqstypes.Message{MessageAttributes: map[string]sqstypes.MessageAttributeValue{}}
	_, ok := ExtractSpanLinkFromSQS(msg)
	if ok {
		t.Error("expected false for message with no trace attributes")
	}
}

func TestExtractSpanLinkFromSQS_NilMessageAttributes(t *testing.T) {
	msg := sqstypes.Message{}
	_, ok := ExtractSpanLinkFromSQS(msg)
	if ok {
		t.Error("expected false for message with nil MessageAttributes")
	}
}

func TestExtractSpanLinkFromSQS_WithValidTraceparent(t *testing.T) {
	// Inject a real span into a carrier, then use those attributes as message attrs.
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "producer-span")
	defer span.End()

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	attrs := make(map[string]sqstypes.MessageAttributeValue)
	for k, v := range carrier {
		val := v
		attrs[k] = sqstypes.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(val),
		}
	}

	msg := sqstypes.Message{MessageAttributes: attrs}
	link, ok := ExtractSpanLinkFromSQS(msg)
	if !ok {
		t.Fatal("expected true for message with valid traceparent")
	}
	if !link.SpanContext.IsValid() {
		t.Error("expected valid span context in the extracted link")
	}
}

func TestExtractSpanLinkFromSQS_NilStringValue(t *testing.T) {
	attrs := map[string]sqstypes.MessageAttributeValue{
		"traceparent": {DataType: aws.String("String"), StringValue: nil},
	}
	msg := sqstypes.Message{MessageAttributes: attrs}
	_, ok := ExtractSpanLinkFromSQS(msg)
	if ok {
		t.Error("expected false when StringValue is nil (no valid trace context)")
	}
}

// --- logging ---

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	// After NewLogger, slog.Default() should be the returned logger.
	if slog.Default() == nil {
		t.Error("expected slog.Default() to be non-nil after NewLogger")
	}
}

func TestLoggerFromContext_NoSpan(t *testing.T) {
	// Background context has no span → must return slog.Default() without panic.
	logger := LoggerFromContext(context.Background())
	if logger == nil {
		t.Error("expected non-nil logger for context without span")
	}
}

func TestLoggerFromContext_WithActiveSpan(t *testing.T) {
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-handler")
	defer span.End()

	logger := LoggerFromContext(ctx)
	if logger == nil {
		t.Error("expected non-nil logger for context with active span")
	}
}

// --- metrics ---

func TestInitMetrics_WithNoopMeter(t *testing.T) {
	// Already called in TestMain; calling again replaces instruments — must not error.
	if err := InitMetrics(otel.GetMeterProvider().Meter("test2")); err != nil {
		t.Errorf("InitMetrics returned unexpected error: %v", err)
	}
}

func TestMetricInstruments_NotNil(t *testing.T) {
	if HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration must not be nil after InitMetrics")
	}
	if HTTPRequestCount == nil {
		t.Error("HTTPRequestCount must not be nil after InitMetrics")
	}
	if DBQueryDuration == nil {
		t.Error("DBQueryDuration must not be nil after InitMetrics")
	}
	if ServiceOrderStatusTransition == nil {
		t.Error("ServiceOrderStatusTransition must not be nil after InitMetrics")
	}
}

// --- otel helpers ---

func TestOTelInitialized_FalseBeforeInit(t *testing.T) {
	// We never called InitOTel, so this must be false.
	if OTelInitialized() {
		t.Error("expected OTelInitialized() = false when InitOTel has not been called")
	}
}

func TestTracer_NotNil(t *testing.T) {
	if Tracer() == nil {
		t.Error("Tracer() must not be nil")
	}
}

func TestMeter_NotNil(t *testing.T) {
	if Meter() == nil {
		t.Error("Meter() must not be nil")
	}
}

func TestSpanHandler_CreatesSpan(t *testing.T) {
	ctx, span := SpanHandler(context.Background(), "auth.login")
	defer span.End()
	if ctx == nil {
		t.Error("expected non-nil context from SpanHandler")
	}
	if span == nil {
		t.Error("expected non-nil span from SpanHandler")
	}
}

func TestSpanUseCase_CreatesSpan(t *testing.T) {
	ctx, span := SpanUseCase(context.Background(), "inventory.reserve_stock")
	defer span.End()
	if ctx == nil {
		t.Error("expected non-nil context from SpanUseCase")
	}
	if span == nil {
		t.Error("expected non-nil span from SpanUseCase")
	}
}
