package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
)

// ─── compensationFor ──────────────────────────────────────────────────────────

func TestCompensationFor_AllBranches(t *testing.T) {
	tests := []struct {
		op     saga_operation.Operation
		expect saga_operation.Operation
	}{
		{saga_operation.OperationReserve, saga_operation.OperationCancelReserved},
		{saga_operation.OperationReservedDecrease, saga_operation.OperationCancelConfirmed},
		{saga_operation.OperationCancelReserved, saga_operation.OperationReserve},
		{saga_operation.OperationCancelConfirmed, saga_operation.OperationReservedDecrease},
		{"UNKNOWN_OP", "UNKNOWN_OP"},
	}
	for _, tt := range tests {
		got := compensationFor(tt.op)
		if got != tt.expect {
			t.Errorf("compensationFor(%s) = %s, want %s", tt.op, got, tt.expect)
		}
	}
}

// ─── occurredAt ───────────────────────────────────────────────────────────────

func TestOccurredAt_NilNow(t *testing.T) {
	s := occurredAt(nil)
	if s == "" {
		t.Error("expected non-empty time string for nil now")
	}
}

func TestOccurredAt_WithNow(t *testing.T) {
	fixed, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	s := occurredAt(func() time.Time { return fixed })
	if s != "2026-01-01T00:00:00Z" {
		t.Errorf("expected 2026-01-01T00:00:00Z, got %s", s)
	}
}

// ─── outputFromStoredPayload ──────────────────────────────────────────────────

func TestOutputFromStoredPayload_InvalidJSON(t *testing.T) {
	if _, err := outputFromStoredPayload([]byte("not-json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestOutputFromStoredPayload_FailedEvent(t *testing.T) {
	payload := []byte(`{"event":"OrderInventoryOperationFailed"}`)
	out, err := outputFromStoredPayload(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Succeeded {
		t.Error("expected Succeeded=false for failed event")
	}
}

// ─── Execute — extra branches ─────────────────────────────────────────────────

func TestProcessSagaOperation_AlreadyProcessing(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)

	// Seed a PROCESSING (not complete/failed) operation
	stored, _ := saga_operation.NewSagaOperation(testSagaID, testOrderID, saga_operation.OperationReserve)
	_ = sagaRepo.Save(ctx, stored)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	_, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationReserve,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected error for already-processing duplicate")
	}
}

func TestProcessSagaOperation_FindError(t *testing.T) {
	ctx := context.Background()
	sagaRepo := &errorFindSagaRepo{findErr: errors.New("db down")}
	inventoryRepo := newMemoryInventoryRepo(t)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	_, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationReserve,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected error when find returns DB error")
	}
}

func TestProcessSagaOperation_SaveOperationError(t *testing.T) {
	ctx := context.Background()
	sagaRepo := &errorSaveSagaRepo{saveErr: errors.New("save failed")}
	inventoryRepo := newMemoryInventoryRepo(t)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	_, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationReserve,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected error when initial Save fails")
	}
}

func TestProcessSagaOperation_CancelReservedSuccess(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)
	inventoryRepo.add(t, testProductA, 5, 4, 0)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	output, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationCancelReserved,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 2}},
		Now:       fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Succeeded {
		t.Fatal("expected success")
	}
}

func TestProcessSagaOperation_CancelConfirmedSuccess(t *testing.T) {
	ctx := context.Background()
	sagaRepo := newMemorySagaOperationRepo()
	inventoryRepo := newMemoryInventoryRepo(t)
	inventoryRepo.add(t, testProductA, 5, 0, 4)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	output, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationCancelConfirmed,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 2}},
		Now:       fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Succeeded {
		t.Fatal("expected success")
	}
}

func TestProcessSagaOperation_CompleteSaveError(t *testing.T) {
	ctx := context.Background()
	sagaRepo := &countingSagaRepo{inner: newMemorySagaOperationRepo(), failOnSave: 2}
	inventoryRepo := newMemoryInventoryRepo(t)
	inventoryRepo.add(t, testProductA, 10, 0, 0)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	_, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationReserve,
		Items:     []ProcessSagaItemInput{{ProductID: testProductA, Quantity: 2}},
		Now:       fixedNow,
	})
	if err == nil {
		t.Fatal("expected error when complete Save fails")
	}
}

func TestProcessSagaOperation_FailSaveError(t *testing.T) {
	ctx := context.Background()
	sagaRepo := &countingSagaRepo{inner: newMemorySagaOperationRepo(), failOnSave: 2}
	inventoryRepo := newMemoryInventoryRepo(t)
	// testProductB has no inventory → apply fails → uc.fail is called
	inventoryRepo.add(t, testProductA, 10, 0, 0)

	uc := NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)
	_, err := uc.Execute(ctx, ProcessSagaOperationInput{
		SagaID: testSagaID, OrderID: testOrderID,
		Operation: saga_operation.OperationReserve,
		Items: []ProcessSagaItemInput{
			{ProductID: testProductB, Quantity: 1}, // fails → uc.fail
		},
		Now: fixedNow,
	})
	if err == nil {
		t.Fatal("expected error when fail Save fails")
	}
}

// ─── Mock helpers ─────────────────────────────────────────────────────────────

type errorFindSagaRepo struct{ findErr error }

func (r *errorFindSagaRepo) Save(_ context.Context, _ *saga_operation.SagaOperation) error {
	return nil
}
func (r *errorFindSagaRepo) FindBySagaAndOperation(_ context.Context, _ string, _ saga_operation.Operation) (*saga_operation.SagaOperation, error) {
	return nil, r.findErr
}

type errorSaveSagaRepo struct{ saveErr error }

func (r *errorSaveSagaRepo) Save(_ context.Context, _ *saga_operation.SagaOperation) error {
	return r.saveErr
}
func (r *errorSaveSagaRepo) FindBySagaAndOperation(_ context.Context, _ string, _ saga_operation.Operation) (*saga_operation.SagaOperation, error) {
	return nil, saga_operation.ErrSagaOperationNotFound
}

type countingSagaRepo struct {
	inner      *memorySagaOperationRepo
	saveCount  int
	failOnSave int
}

func (r *countingSagaRepo) Save(ctx context.Context, op *saga_operation.SagaOperation) error {
	r.saveCount++
	if r.saveCount >= r.failOnSave {
		return errors.New("save failed")
	}
	return r.inner.Save(ctx, op)
}
func (r *countingSagaRepo) FindBySagaAndOperation(ctx context.Context, sagaID string, op saga_operation.Operation) (*saga_operation.SagaOperation, error) {
	return r.inner.FindBySagaAndOperation(ctx, sagaID, op)
}
