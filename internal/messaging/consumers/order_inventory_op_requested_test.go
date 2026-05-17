package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	inventorydomain "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	sagaop "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
)

// --- mocks ---

type mockSQSRDClient struct {
	deleteErr error
}

func (m *mockSQSRDClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return &sqs.ReceiveMessageOutput{}, nil
}

func (m *mockSQSRDClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return &sqs.DeleteMessageOutput{}, m.deleteErr
}

type mockOpPublisher struct {
	err error
}

func (m *mockOpPublisher) Publish(ctx context.Context, result *usecases.ProcessSagaOperationOutput) error {
	return m.err
}

type mockSagaRepo struct {
	saveErr error
}

func (m *mockSagaRepo) Save(ctx context.Context, op *sagaop.SagaOperation) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if op.ID() == "" {
		_ = op.SetID(uuid.New().String())
	}
	return nil
}

func (m *mockSagaRepo) FindBySagaAndOperation(_ context.Context, _ string, _ sagaop.Operation) (*sagaop.SagaOperation, error) {
	return nil, sagaop.ErrSagaOperationNotFound
}

type mockInvRepo struct {
	inv *inventorydomain.Inventory
}

func (m *mockInvRepo) FindByProductID(_ context.Context, _ string) (*inventorydomain.Inventory, error) {
	if m.inv == nil {
		return nil, fmt.Errorf("inventory not found")
	}
	return m.inv, nil
}

func (m *mockInvRepo) Save(_ context.Context, _ *inventorydomain.Inventory) error        { return nil }
func (m *mockInvRepo) FindByID(_ context.Context, _ string) (*inventorydomain.Inventory, error) {
	return nil, nil
}
func (m *mockInvRepo) FindAll(_ context.Context) ([]*inventorydomain.Inventory, error) {
	return nil, nil
}
func (m *mockInvRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockInvRepo) ExistsByProductID(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// --- helpers ---

func makeInventory(productID string) *inventorydomain.Inventory {
	inv, _ := inventorydomain.ReconstructInventory(uuid.New().String(), productID, 100, 0, 0, time.Now(), time.Now(), nil)
	return inv
}

func makeConsumer(sagaRepo sagaop.Repository, invRepo inventorydomain.Repository, sqsCl SQSReceiveDeleteClient, pub OrderInventoryOperationPublisher) *OrderInventoryOperationRequestedConsumer {
	uc := usecases.NewProcessSagaOperationUseCase(sagaRepo, invRepo)
	return NewOrderInventoryOperationRequestedConsumer(sqsCl, "https://sqs.example.com/test-queue", uc, pub)
}

func makeMessage(sagaID, orderID, productID string) sqstypes.Message {
	body := fmt.Sprintf(
		`{"event":"OrderInventoryOperationRequested","saga_id":%q,"order_id":%q,"operation":"RESERVE","items":[{"product_id":%q,"quantity":5}],"occurred_at":"2024-01-15T10:00:00Z"}`,
		sagaID, orderID, productID,
	)
	receipt := "receipt-handle-abc"
	return sqstypes.Message{Body: aws.String(body), ReceiptHandle: aws.String(receipt)}
}

// --- decodeOrderInventoryOperationRequested ---

func TestDecode_NilBody(t *testing.T) {
	_, err := decodeOrderInventoryOperationRequested(sqstypes.Message{Body: nil})
	if err == nil {
		t.Error("expected error for nil body")
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	_, err := decodeOrderInventoryOperationRequested(sqstypes.Message{Body: aws.String("not-json")})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecode_WrongEvent(t *testing.T) {
	body := `{"event":"SomethingElse","saga_id":"x","order_id":"x","operation":"RESERVE","items":[],"occurred_at":"2024-01-01T00:00:00Z"}`
	_, err := decodeOrderInventoryOperationRequested(sqstypes.Message{Body: aws.String(body)})
	if err == nil {
		t.Error("expected error for unexpected event type")
	}
}

func TestDecode_InvalidOccurredAt(t *testing.T) {
	body := `{"event":"OrderInventoryOperationRequested","saga_id":"x","order_id":"x","operation":"RESERVE","items":[],"occurred_at":"not-a-time"}`
	_, err := decodeOrderInventoryOperationRequested(sqstypes.Message{Body: aws.String(body)})
	if err == nil {
		t.Error("expected error for invalid occurred_at format")
	}
}

func TestDecode_Valid(t *testing.T) {
	sagaID := uuid.New().String()
	orderID := uuid.New().String()
	productID := uuid.New().String()

	body := fmt.Sprintf(
		`{"event":"OrderInventoryOperationRequested","saga_id":%q,"order_id":%q,"operation":"RESERVE","items":[{"product_id":%q,"quantity":3}],"occurred_at":"2024-01-15T10:00:00Z"}`,
		sagaID, orderID, productID,
	)
	input, err := decodeOrderInventoryOperationRequested(sqstypes.Message{Body: aws.String(body)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.SagaID != sagaID {
		t.Errorf("SagaID = %s, want %s", input.SagaID, sagaID)
	}
	if input.OrderID != orderID {
		t.Errorf("OrderID = %s, want %s", input.OrderID, orderID)
	}
	if len(input.Items) != 1 || input.Items[0].Quantity != 3 {
		t.Errorf("expected 1 item with quantity 3, got %+v", input.Items)
	}
	if input.Operation != sagaop.OperationReserve {
		t.Errorf("Operation = %s, want RESERVE", input.Operation)
	}
}

// --- HandleMessage ---

func TestHandleMessage_NilBody(t *testing.T) {
	c := makeConsumer(&mockSagaRepo{}, &mockInvRepo{}, &mockSQSRDClient{}, &mockOpPublisher{})
	err := c.HandleMessage(context.Background(), sqstypes.Message{Body: nil})
	if err == nil {
		t.Error("expected error for nil body")
	}
}

func TestHandleMessage_WrongEvent(t *testing.T) {
	c := makeConsumer(&mockSagaRepo{}, &mockInvRepo{}, &mockSQSRDClient{}, &mockOpPublisher{})
	body := `{"event":"WrongEvent","saga_id":"x","order_id":"x","operation":"RESERVE","items":[],"occurred_at":"2024-01-01T00:00:00Z"}`
	err := c.HandleMessage(context.Background(), sqstypes.Message{Body: aws.String(body)})
	if err == nil {
		t.Error("expected error for wrong event type")
	}
}

func TestHandleMessage_UseCaseError(t *testing.T) {
	sagaID, orderID, productID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	// Saga repo Save will fail → use case returns error
	c := makeConsumer(&mockSagaRepo{saveErr: fmt.Errorf("db error")}, &mockInvRepo{inv: makeInventory(productID)}, &mockSQSRDClient{}, &mockOpPublisher{})
	err := c.HandleMessage(context.Background(), makeMessage(sagaID, orderID, productID))
	if err == nil {
		t.Error("expected error when use case fails")
	}
}

func TestHandleMessage_PublisherError(t *testing.T) {
	sagaID, orderID, productID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	c := makeConsumer(&mockSagaRepo{}, &mockInvRepo{inv: makeInventory(productID)}, &mockSQSRDClient{}, &mockOpPublisher{err: fmt.Errorf("publish failed")})
	err := c.HandleMessage(context.Background(), makeMessage(sagaID, orderID, productID))
	if err == nil {
		t.Error("expected error when publisher fails")
	}
}

func TestHandleMessage_MissingReceiptHandle(t *testing.T) {
	sagaID, orderID, productID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	c := makeConsumer(&mockSagaRepo{}, &mockInvRepo{inv: makeInventory(productID)}, &mockSQSRDClient{}, &mockOpPublisher{})

	body := fmt.Sprintf(
		`{"event":"OrderInventoryOperationRequested","saga_id":%q,"order_id":%q,"operation":"RESERVE","items":[{"product_id":%q,"quantity":5}],"occurred_at":"2024-01-15T10:00:00Z"}`,
		sagaID, orderID, productID,
	)
	err := c.HandleMessage(context.Background(), sqstypes.Message{Body: aws.String(body)})
	if err == nil {
		t.Error("expected error for missing receipt handle")
	}
}

func TestHandleMessage_DeleteError(t *testing.T) {
	sagaID, orderID, productID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	sqsCl := &mockSQSRDClient{deleteErr: fmt.Errorf("delete error")}
	c := makeConsumer(&mockSagaRepo{}, &mockInvRepo{inv: makeInventory(productID)}, sqsCl, &mockOpPublisher{})
	err := c.HandleMessage(context.Background(), makeMessage(sagaID, orderID, productID))
	if err == nil {
		t.Error("expected error when DeleteMessage fails")
	}
}

func TestHandleMessage_Success(t *testing.T) {
	sagaID, orderID, productID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	c := makeConsumer(&mockSagaRepo{}, &mockInvRepo{inv: makeInventory(productID)}, &mockSQSRDClient{}, &mockOpPublisher{})
	if err := c.HandleMessage(context.Background(), makeMessage(sagaID, orderID, productID)); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
