package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
)

const (
	testSagaID   = "11111111-1111-4111-8111-111111111111"
	testOrderID  = "22222222-2222-4222-8222-222222222222"
	testProductA = "33333333-3333-4333-8333-333333333333"
	testProductB = "44444444-4444-4444-8444-444444444444"
	testOccurred = "2026-04-30T12:00:00Z"
)

func TestProcessSagaOperation_NewSuccess(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)
	inventoryRepo.add(t, testProductA, 10, 0, 0)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	output, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID:    testSagaID,
		OrderID:   testOrderID,
		Operation: saga_operation.OperationReserve,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 3}},
		Now:       fixedNow,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !output.Succeeded || output.EventName != "OrderInventoryOperationSucceeded" {
		t.Fatalf("expected success event, got %#v", output)
	}

	inv := inventoryRepo.mustFind(t, testProductA)
	if inv.AvailableQuantity() != 7 || inv.ReservedQuantity() != 3 || inv.PendingQuantity() != 0 {
		t.Fatalf("unexpected inventory state: available=%d reserved=%d pending=%d", inv.AvailableQuantity(), inv.ReservedQuantity(), inv.PendingQuantity())
	}
}

func TestProcessSagaOperation_DuplicateReturnsStoredPayload(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)
	stored, err := saga_operation.NewSagaOperation(testSagaID, testOrderID, saga_operation.OperationReserve)
	if err != nil {
		t.Fatal(err)
	}
	stored.Complete([]byte(`{"event":"OrderInventoryOperationSucceeded","saga_id":"11111111-1111-4111-8111-111111111111","order_id":"22222222-2222-4222-8222-222222222222","operation":"RESERVE","occurred_at":"2026-04-30T12:00:00Z"}`))
	if err := sagaRepo.Save(ctx, stored); err != nil {
		t.Fatal(err)
	}

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	output, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID:    testSagaID,
		OrderID:   testOrderID,
		Operation: saga_operation.OperationReserve,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 3}},
		Now:       fixedNow,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !output.Succeeded || output.EventName != "OrderInventoryOperationSucceeded" {
		t.Fatalf("expected stored success event, got %#v", output)
	}
	if len(inventoryRepo.items) != 0 {
		t.Fatalf("duplicate should not touch inventory")
	}
}

func TestProcessSagaOperation_PartialFailureCompensatesAndStoresFailed(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)
	inventoryRepo.add(t, testProductA, 10, 0, 0)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	output, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID:    testSagaID,
		OrderID:   testOrderID,
		Operation: saga_operation.OperationReserve,
		Items: []ProcessSagaItemInput{
			{ProductID: testProductA, Quantity: 3},
			{ProductID: testProductB, Quantity: 1},
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.Succeeded || output.EventName != "OrderInventoryOperationFailed" {
		t.Fatalf("expected failed event, got %#v", output)
	}

	inv := inventoryRepo.mustFind(t, testProductA)
	if inv.AvailableQuantity() != 10 || inv.ReservedQuantity() != 0 || inv.PendingQuantity() != 0 {
		t.Fatalf("expected compensation to restore inventory, got available=%d reserved=%d pending=%d", inv.AvailableQuantity(), inv.ReservedQuantity(), inv.PendingQuantity())
	}

	stored, err := sagaRepo.FindBySagaAndOperation(ctx, testSagaID, saga_operation.OperationReserve)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status() != saga_operation.StatusFailed {
		t.Fatalf("expected failed status, got %s", stored.Status())
	}
}

func TestProcessSagaOperation_ReservedDecreaseSuccess(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)
	inventoryRepo.add(t, testProductA, 5, 4, 0)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	output, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID:    testSagaID,
		OrderID:   testOrderID,
		Operation: saga_operation.OperationReservedDecrease,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 2}},
		Now:       fixedNow,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !output.Succeeded {
		t.Fatalf("expected success")
	}

	inv := inventoryRepo.mustFind(t, testProductA)
	if inv.ReservedQuantity() != 2 {
		t.Fatalf("expected reserved=2, got %d", inv.ReservedQuantity())
	}
}

type memorySagaOperationRepo struct {
	items map[string]*saga_operation.SagaOperation
}

func newMemorySagaOperationRepo() *memorySagaOperationRepo {
	return &memorySagaOperationRepo{items: map[string]*saga_operation.SagaOperation{}}
}

func (r *memorySagaOperationRepo) Save(_ context.Context, operation *saga_operation.SagaOperation) error {
	if operation.ID() == "" {
		if err := operation.SetID("55555555-5555-4555-8555-555555555555"); err != nil {
			return err
		}
	}
	r.items[operation.SagaID()+"|"+string(operation.Operation())] = operation
	return nil
}

func (r *memorySagaOperationRepo) FindBySagaAndOperation(_ context.Context, sagaID string, operation saga_operation.Operation) (*saga_operation.SagaOperation, error) {
	item, ok := r.items[sagaID+"|"+string(operation)]
	if !ok {
		return nil, saga_operation.ErrSagaOperationNotFound
	}
	return item, nil
}

type memoryInventoryRepo struct {
	items map[string]*inventory.Inventory
}

func newMemoryInventoryRepo(_ *testing.T) *memoryInventoryRepo {
	return &memoryInventoryRepo{items: map[string]*inventory.Inventory{}}
}

func (r *memoryInventoryRepo) add(t *testing.T, productID string, available, reserved, pending int) {
	t.Helper()
	inv, err := inventory.ReconstructInventory("66666666-6666-4666-8666-666666666666", productID, available, reserved, pending, time.Now(), time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.items[productID] = inv
}

func (r *memoryInventoryRepo) Save(_ context.Context, inv *inventory.Inventory) error {
	r.items[inv.ProductID()] = inv
	return nil
}

func (r *memoryInventoryRepo) FindByID(context.Context, string) (*inventory.Inventory, error) {
	return nil, errors.New("not implemented")
}

func (r *memoryInventoryRepo) FindByProductID(_ context.Context, productID string) (*inventory.Inventory, error) {
	inv, ok := r.items[productID]
	if !ok {
		return nil, inventory.ErrInventoryNotFound
	}
	return inv, nil
}

func (r *memoryInventoryRepo) FindAll(context.Context) ([]*inventory.Inventory, error) {
	return nil, errors.New("not implemented")
}

func (r *memoryInventoryRepo) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func (r *memoryInventoryRepo) ExistsByProductID(context.Context, string) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *memoryInventoryRepo) mustFind(t *testing.T, productID string) *inventory.Inventory {
	t.Helper()
	inv, err := r.FindByProductID(context.Background(), productID)
	if err != nil {
		t.Fatal(err)
	}
	return inv
}

func fixedNow() time.Time {
	when, _ := time.Parse(time.RFC3339, testOccurred)
	return when
}
