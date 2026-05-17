package publishers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
)

type mockSQSSendClient struct {
	capturedURL  string
	capturedBody string
	err          error
}

func (m *mockSQSSendClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	if params.QueueUrl != nil {
		m.capturedURL = *params.QueueUrl
	}
	if params.MessageBody != nil {
		m.capturedBody = *params.MessageBody
	}
	return &sqs.SendMessageOutput{}, m.err
}

func succeededResult(payload []byte) *usecases.ProcessSagaOperationOutput {
	return &usecases.ProcessSagaOperationOutput{
		EventName: "OrderInventoryOperationSucceeded",
		Payload:   payload,
		Succeeded: true,
	}
}

func failedResult(payload []byte) *usecases.ProcessSagaOperationOutput {
	return &usecases.ProcessSagaOperationOutput{
		EventName: "OrderInventoryOperationFailed",
		Payload:   payload,
		Succeeded: false,
	}
}

func TestPublish_NilResult(t *testing.T) {
	p := NewOrderInventoryOperationPublisher(&mockSQSSendClient{}, "http://sqs/s", "http://sqs/f")
	if err := p.Publish(context.Background(), nil); err == nil {
		t.Error("expected error for nil result")
	}
}

func TestPublish_SucceededRoutesToSucceededQueue(t *testing.T) {
	const succeededURL = "https://sqs.example.com/succeeded"
	client := &mockSQSSendClient{}
	p := NewOrderInventoryOperationPublisher(client, succeededURL, "https://sqs.example.com/failed")

	if err := p.Publish(context.Background(), succeededResult([]byte(`{}`))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.capturedURL != succeededURL {
		t.Errorf("expected queue %s, got %s", succeededURL, client.capturedURL)
	}
}

func TestPublish_FailedRoutesToFailedQueue(t *testing.T) {
	const failedURL = "https://sqs.example.com/failed"
	client := &mockSQSSendClient{}
	p := NewOrderInventoryOperationPublisher(client, "https://sqs.example.com/succeeded", failedURL)

	if err := p.Publish(context.Background(), failedResult([]byte(`{"reason":"stock"}`))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.capturedURL != failedURL {
		t.Errorf("expected queue %s, got %s", failedURL, client.capturedURL)
	}
}

func TestPublish_SendMessageError(t *testing.T) {
	client := &mockSQSSendClient{err: fmt.Errorf("SQS unavailable")}
	p := NewOrderInventoryOperationPublisher(client, "http://sqs/s", "http://sqs/f")

	err := p.Publish(context.Background(), succeededResult([]byte(`{}`)))
	if err == nil {
		t.Fatal("expected error when SendMessage fails")
	}
	if !strings.Contains(err.Error(), "SQS unavailable") {
		t.Errorf("expected wrapped SQS error, got: %v", err)
	}
}

func TestPublish_PayloadForwardedAsBody(t *testing.T) {
	client := &mockSQSSendClient{}
	p := NewOrderInventoryOperationPublisher(client, "http://sqs/s", "http://sqs/f")

	payload := []byte(`{"event":"OrderInventoryOperationSucceeded","saga_id":"abc-123"}`)
	if err := p.Publish(context.Background(), succeededResult(payload)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.capturedBody != string(payload) {
		t.Errorf("body = %q, want %q", client.capturedBody, payload)
	}
}
