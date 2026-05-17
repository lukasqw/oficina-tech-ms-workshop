package publishers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

	body := string(result.Payload)
	_, err := p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(queueURL),
		MessageBody:       aws.String(body),
		MessageAttributes: observability.InjectTraceToSQS(ctx),
	})
	if err != nil {
		return fmt.Errorf("publish %s: %w", result.EventName, err)
	}
	return nil
}
