package publishers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
)

type SQSSendMessageClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type OrderInventoryOperationPublisher struct {
	client          SQSSendMessageClient
	succeededQueueURL string
	failedQueueURL    string
}

func NewOrderInventoryOperationPublisher(client SQSSendMessageClient, succeededQueueURL, failedQueueURL string) *OrderInventoryOperationPublisher {
	return &OrderInventoryOperationPublisher{
		client:            client,
		succeededQueueURL: succeededQueueURL,
		failedQueueURL:    failedQueueURL,
	}
}

func (p *OrderInventoryOperationPublisher) Publish(ctx context.Context, result *usecases.ProcessSagaOperationOutput) error {
	if result == nil {
		return fmt.Errorf("nil saga operation result")
	}

	queueURL := p.failedQueueURL
	if result.Succeeded {
		queueURL = p.succeededQueueURL
	}

	ctx, span := otel.Tracer("ms-workshop/messaging").Start(ctx,
		"publish "+result.EventName,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("aws_sqs"),
			semconv.MessagingDestinationName(queueURL),
			attribute.String("messaging.operation.name", "publish"),
		),
	)
	defer span.End()

	body := string(result.Payload)
	_, err := p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(queueURL),
		MessageBody:       aws.String(body),
		MessageAttributes: observability.InjectTraceToSQS(ctx),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("publish %s: %w", result.EventName, err)
	}
	return nil
}
