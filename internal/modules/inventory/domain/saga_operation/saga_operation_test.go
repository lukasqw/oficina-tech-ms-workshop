package saga_operation

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newValidUUID() string {
	return uuid.New().String()
}

func TestNewSagaOperation(t *testing.T) {
	sagaID := newValidUUID()
	orderID := newValidUUID()

	tests := []struct {
		name      string
		sagaID    string
		orderID   string
		operation Operation
		wantErr   error
	}{
		{
			name:      "reserve operation",
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationReserve,
		},
		{
			name:      "reserved decrease operation",
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationReservedDecrease,
		},
		{
			name:      "cancel reserved operation",
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationCancelReserved,
		},
		{
			name:      "cancel confirmed operation",
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationCancelConfirmed,
		},
		{
			name:      "invalid sagaID",
			sagaID:    "not-a-uuid",
			orderID:   orderID,
			operation: OperationReserve,
			wantErr:   ErrInvalidSagaID,
		},
		{
			name:      "invalid orderID",
			sagaID:    sagaID,
			orderID:   "not-a-uuid",
			operation: OperationReserve,
			wantErr:   ErrInvalidOrderID,
		},
		{
			name:      "invalid operation",
			sagaID:    sagaID,
			orderID:   orderID,
			operation: "INVALID",
			wantErr:   ErrInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := NewSagaOperation(tt.sagaID, tt.orderID, tt.operation)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if op != nil {
					t.Error("expected nil result on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if op.SagaID() != tt.sagaID {
				t.Errorf("expected sagaID %s, got %s", tt.sagaID, op.SagaID())
			}
			if op.OrderID() != tt.orderID {
				t.Errorf("expected orderID %s, got %s", tt.orderID, op.OrderID())
			}
			if op.Operation() != tt.operation {
				t.Errorf("expected operation %s, got %s", tt.operation, op.Operation())
			}
			if op.Status() != StatusProcessing {
				t.Errorf("expected status PROCESSING, got %s", op.Status())
			}
			if op.ID() != "" {
				t.Error("expected empty ID for new operation")
			}
			if op.ResultPayload() != nil {
				t.Error("expected nil result payload for new operation")
			}
			if op.ProcessedAt() != nil {
				t.Error("expected nil processedAt for new operation")
			}
		})
	}
}

func TestReconstructSagaOperation(t *testing.T) {
	id := newValidUUID()
	sagaID := newValidUUID()
	orderID := newValidUUID()
	now := time.Now()
	payload := []byte(`{"result":"ok"}`)

	tests := []struct {
		name          string
		id            string
		sagaID        string
		orderID       string
		operation     Operation
		status        Status
		resultPayload []byte
		processedAt   *time.Time
		wantErr       error
	}{
		{
			name:          "valid completed operation",
			id:            id,
			sagaID:        sagaID,
			orderID:       orderID,
			operation:     OperationReserve,
			status:        StatusCompleted,
			resultPayload: payload,
			processedAt:   &now,
		},
		{
			name:      "valid processing operation",
			id:        id,
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationCancelReserved,
			status:    StatusProcessing,
		},
		{
			name:      "valid failed operation",
			id:        id,
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationReservedDecrease,
			status:    StatusFailed,
		},
		{
			name:      "invalid id",
			id:        "bad-id",
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationReserve,
			status:    StatusCompleted,
			wantErr:   ErrInvalidSagaOperationID,
		},
		{
			name:      "invalid sagaID",
			id:        id,
			sagaID:    "bad-saga",
			orderID:   orderID,
			operation: OperationReserve,
			status:    StatusCompleted,
			wantErr:   ErrInvalidSagaID,
		},
		{
			name:      "invalid orderID",
			id:        id,
			sagaID:    sagaID,
			orderID:   "bad-order",
			operation: OperationReserve,
			status:    StatusCompleted,
			wantErr:   ErrInvalidOrderID,
		},
		{
			name:      "invalid operation",
			id:        id,
			sagaID:    sagaID,
			orderID:   orderID,
			operation: "UNKNOWN",
			status:    StatusCompleted,
			wantErr:   ErrInvalidOperation,
		},
		{
			name:      "invalid status",
			id:        id,
			sagaID:    sagaID,
			orderID:   orderID,
			operation: OperationReserve,
			status:    "UNKNOWN",
			wantErr:   ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := ReconstructSagaOperation(tt.id, tt.sagaID, tt.orderID, tt.operation, tt.status, tt.resultPayload, tt.processedAt)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if op != nil {
					t.Error("expected nil result on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if op.ID() != tt.id {
				t.Errorf("expected id %s, got %s", tt.id, op.ID())
			}
			if op.SagaID() != tt.sagaID {
				t.Errorf("expected sagaID %s, got %s", tt.sagaID, op.SagaID())
			}
			if op.OrderID() != tt.orderID {
				t.Errorf("expected orderID %s, got %s", tt.orderID, op.OrderID())
			}
			if op.Operation() != tt.operation {
				t.Errorf("expected operation %s, got %s", tt.operation, op.Operation())
			}
			if op.Status() != tt.status {
				t.Errorf("expected status %s, got %s", tt.status, op.Status())
			}
		})
	}
}

func TestOperation_Valid(t *testing.T) {
	tests := []struct {
		op    Operation
		valid bool
	}{
		{OperationReserve, true},
		{OperationReservedDecrease, true},
		{OperationCancelReserved, true},
		{OperationCancelConfirmed, true},
		{"UNKNOWN", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			if got := tt.op.Valid(); got != tt.valid {
				t.Errorf("Operation(%q).Valid() = %v, want %v", tt.op, got, tt.valid)
			}
		})
	}
}

func TestStatus_Valid(t *testing.T) {
	tests := []struct {
		status Status
		valid  bool
	}{
		{StatusProcessing, true},
		{StatusCompleted, true},
		{StatusFailed, true},
		{"UNKNOWN", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.valid {
				t.Errorf("Status(%q).Valid() = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestSagaOperation_SetID(t *testing.T) {
	op, err := NewSagaOperation(newValidUUID(), newValidUUID(), OperationReserve)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("valid UUID", func(t *testing.T) {
		id := newValidUUID()
		if err := op.SetID(id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op.ID() != id {
			t.Errorf("expected ID %s, got %s", id, op.ID())
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		err := op.SetID("not-a-uuid")
		if err != ErrInvalidSagaOperationID {
			t.Errorf("expected ErrInvalidSagaOperationID, got %v", err)
		}
	})
}

func TestSagaOperation_Complete(t *testing.T) {
	op, err := NewSagaOperation(newValidUUID(), newValidUUID(), OperationReserve)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := []byte(`{"stock":"ok"}`)
	before := time.Now()
	op.Complete(payload)
	after := time.Now()

	if op.Status() != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %s", op.Status())
	}
	if string(op.ResultPayload()) != string(payload) {
		t.Errorf("expected payload %s, got %s", payload, op.ResultPayload())
	}
	if op.ProcessedAt() == nil {
		t.Fatal("expected processedAt to be set")
	}
	if op.ProcessedAt().Before(before) || op.ProcessedAt().After(after) {
		t.Error("processedAt not in expected range")
	}

	// payload must be a defensive copy
	original := op.ResultPayload()[0]
	payload[0] = 'X'
	if op.ResultPayload()[0] != original {
		t.Error("expected ResultPayload to be a defensive copy")
	}
}

func TestSagaOperation_Fail(t *testing.T) {
	op, err := NewSagaOperation(newValidUUID(), newValidUUID(), OperationReserve)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := []byte(`{"error":"insufficient stock"}`)
	before := time.Now()
	op.Fail(payload)
	after := time.Now()

	if op.Status() != StatusFailed {
		t.Errorf("expected StatusFailed, got %s", op.Status())
	}
	if string(op.ResultPayload()) != string(payload) {
		t.Errorf("expected payload %s, got %s", payload, op.ResultPayload())
	}
	if op.ProcessedAt() == nil {
		t.Fatal("expected processedAt to be set")
	}
	if op.ProcessedAt().Before(before) || op.ProcessedAt().After(after) {
		t.Error("processedAt not in expected range")
	}
}

func TestSagaOperation_Getters(t *testing.T) {
	sagaID := newValidUUID()
	orderID := newValidUUID()
	now := time.Now()
	payload := []byte(`{"data":"value"}`)
	id := newValidUUID()

	op, err := ReconstructSagaOperation(id, sagaID, orderID, OperationCancelConfirmed, StatusFailed, payload, &now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if op.ID() != id {
		t.Errorf("ID() = %s, want %s", op.ID(), id)
	}
	if op.SagaID() != sagaID {
		t.Errorf("SagaID() = %s, want %s", op.SagaID(), sagaID)
	}
	if op.OrderID() != orderID {
		t.Errorf("OrderID() = %s, want %s", op.OrderID(), orderID)
	}
	if op.Operation() != OperationCancelConfirmed {
		t.Errorf("Operation() = %s, want %s", op.Operation(), OperationCancelConfirmed)
	}
	if op.Status() != StatusFailed {
		t.Errorf("Status() = %s, want %s", op.Status(), StatusFailed)
	}
	if string(op.ResultPayload()) != string(payload) {
		t.Errorf("ResultPayload() = %s, want %s", op.ResultPayload(), payload)
	}
	if op.ProcessedAt() == nil || !op.ProcessedAt().Equal(now) {
		t.Errorf("ProcessedAt() = %v, want %v", op.ProcessedAt(), now)
	}
}
