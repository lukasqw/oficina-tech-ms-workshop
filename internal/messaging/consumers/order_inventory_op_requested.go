package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
)

type SQSReceiveDeleteClient interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type OrderInventoryOperationPublisher interface {
	Publish(ctx context.Context, result *usecases.ProcessSagaOperationOutput) error
}

type OrderInventoryOperationRequestedConsumer struct {
	client    SQSReceiveDeleteClient
	queueURL  string
	useCase   *usecases.ProcessSagaOperationUseCase
	publisher OrderInventoryOperationPublisher
}

type orderInventoryOperationRequested struct {
	Event      string                   `json:"event"`
	SagaID     string                   `json:"saga_id"`
	OrderID    string                   `json:"order_id"`
	Operation  saga_operation.Operation `json:"operation"`
	Items      []orderInventoryItem     `json:"items"`
	OccurredAt string                   `json:"occurred_at"`
}

type orderInventoryItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func NewOrderInventoryOperationRequestedConsumer(
	client SQSReceiveDeleteClient,
	queueURL string,
	useCase *usecases.ProcessSagaOperationUseCase,
	publisher OrderInventoryOperationPublisher,
) *OrderInventoryOperationRequestedConsumer {
	return &OrderInventoryOperationRequestedConsumer{
		client:    client,
		queueURL:  queueURL,
		useCase:   useCase,
		publisher: publisher,
	}
}

func (c *OrderInventoryOperationRequestedConsumer) Start(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		output, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			WaitTimeSeconds:     20,
			MaxNumberOfMessages: 10,
		})
		if err != nil {
			return fmt.Errorf("receive order inventory operation messages: %w", err)
		}

		for _, message := range output.Messages {
			if err := c.HandleMessage(ctx, message); err != nil {
				slog.Error("failed to process order inventory operation message", "error", err)
			}
		}
	}
}

func (c *OrderInventoryOperationRequestedConsumer) HandleMessage(ctx context.Context, message types.Message) error {
	input, err := decodeOrderInventoryOperationRequested(message)
	if err != nil {
		return err
	}

	result, err := c.useCase.Execute(ctx, input)
	if err != nil {
		return err
	}

	if err := c.publisher.Publish(ctx, result); err != nil {
		return err
	}

	if message.ReceiptHandle == nil {
		return fmt.Errorf("missing SQS receipt handle")
	}

	_, err = c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}

func decodeOrderInventoryOperationRequested(message types.Message) (usecases.ProcessSagaOperationInput, error) {
	if message.Body == nil {
		return usecases.ProcessSagaOperationInput{}, fmt.Errorf("empty SQS message body")
	}

	var event orderInventoryOperationRequested
	if err := json.Unmarshal([]byte(*message.Body), &event); err != nil {
		return usecases.ProcessSagaOperationInput{}, fmt.Errorf("decode OrderInventoryOperationRequested: %w", err)
	}
	if event.Event != "OrderInventoryOperationRequested" {
		return usecases.ProcessSagaOperationInput{}, fmt.Errorf("unexpected event %q", event.Event)
	}
	if _, err := time.Parse(time.RFC3339, event.OccurredAt); err != nil {
		return usecases.ProcessSagaOperationInput{}, fmt.Errorf("invalid occurred_at: %w", err)
	}

	items := make([]usecases.ProcessSagaItemInput, len(event.Items))
	for i, item := range event.Items {
		items[i] = usecases.ProcessSagaItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return usecases.ProcessSagaOperationInput{
		SagaID:    event.SagaID,
		OrderID:   event.OrderID,
		Operation: event.Operation,
		Items:     items,
	}, nil
}
