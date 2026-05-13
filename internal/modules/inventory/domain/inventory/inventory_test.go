package inventory

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewInventory(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		wantErr   error
	}{
		{
			name:      "deve criar inventory com productID válido",
			productID: uuid.New().String(),
			wantErr:   nil,
		},
		{
			name:      "deve retornar erro quando productID é vazio",
			productID: "",
			wantErr:   ErrInvalidProductID,
		},
		{
			name:      "deve retornar erro quando productID não é UUID válido",
			productID: "invalid-uuid",
			wantErr:   ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := NewInventory(tt.productID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewInventory() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewInventory() unexpected error = %v", err)
				return
			}

			if inv.ProductID() != tt.productID {
				t.Errorf("ProductID = %v, want %v", inv.ProductID(), tt.productID)
			}
			if inv.AvailableQuantity() != 0 {
				t.Errorf("AvailableQuantity = %v, want 0", inv.AvailableQuantity())
			}
			if inv.ReservedQuantity() != 0 {
				t.Errorf("ReservedQuantity = %v, want 0", inv.ReservedQuantity())
			}
			if inv.PendingQuantity() != 0 {
				t.Errorf("PendingQuantity = %v, want 0", inv.PendingQuantity())
			}
		})
	}
}

func TestReconstructInventory(t *testing.T) {
	validID := uuid.New().String()
	validProductID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name              string
		id                string
		productID         string
		availableQuantity int
		reservedQuantity  int
		pendingQuantity   int
		createdAt         time.Time
		updatedAt         time.Time
		deletedAt         *time.Time
		wantErr           error
	}{
		{
			name:              "deve reconstruir inventory válido",
			id:                validID,
			productID:         validProductID,
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   2,
			createdAt:         now,
			updatedAt:         now,
			deletedAt:         nil,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando id é vazio",
			id:                "",
			productID:         validProductID,
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   2,
			createdAt:         now,
			updatedAt:         now,
			deletedAt:         nil,
			wantErr:           ErrInvalidInventoryID,
		},
		{
			name:              "deve retornar erro quando productID é vazio",
			id:                validID,
			productID:         "",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   2,
			createdAt:         now,
			updatedAt:         now,
			deletedAt:         nil,
			wantErr:           ErrInvalidProductID,
		},
		{
			name:              "deve retornar erro quando availableQuantity é negativo",
			id:                validID,
			productID:         validProductID,
			availableQuantity: -1,
			reservedQuantity:  5,
			pendingQuantity:   2,
			createdAt:         now,
			updatedAt:         now,
			deletedAt:         nil,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando reservedQuantity é negativo",
			id:                validID,
			productID:         validProductID,
			availableQuantity: 10,
			reservedQuantity:  -1,
			pendingQuantity:   2,
			createdAt:         now,
			updatedAt:         now,
			deletedAt:         nil,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando pendingQuantity é negativo",
			id:                validID,
			productID:         validProductID,
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   -1,
			createdAt:         now,
			updatedAt:         now,
			deletedAt:         nil,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := ReconstructInventory(
				tt.id,
				tt.productID,
				tt.availableQuantity,
				tt.reservedQuantity,
				tt.pendingQuantity,
				tt.createdAt,
				tt.updatedAt,
				tt.deletedAt,
			)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ReconstructInventory() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReconstructInventory() unexpected error = %v", err)
				return
			}

			if inv.ID() != tt.id {
				t.Errorf("ID = %v, want %v", inv.ID(), tt.id)
			}
			if inv.ProductID() != tt.productID {
				t.Errorf("ProductID = %v, want %v", inv.ProductID(), tt.productID)
			}
			if inv.AvailableQuantity() != tt.availableQuantity {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.availableQuantity)
			}
		})
	}
}

func TestInventory_SetID(t *testing.T) {
	productID := uuid.New().String()
	inv, _ := NewInventory(productID)

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "deve definir ID válido",
			id:      uuid.New().String(),
			wantErr: nil,
		},
		{
			name:    "deve retornar erro quando ID é vazio",
			id:      "",
			wantErr: ErrInvalidInventoryID,
		},
		{
			name:    "deve retornar erro quando ID não é UUID válido",
			id:      "invalid-uuid",
			wantErr: ErrInvalidInventoryID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inv.SetID(tt.id)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("SetID() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("SetID() unexpected error = %v", err)
				return
			}

			if inv.ID() != tt.id {
				t.Errorf("ID = %v, want %v", inv.ID(), tt.id)
			}
		})
	}
}

func TestInventory_ManualDecrease(t *testing.T) {
	tests := []struct {
		name              string
		availableQuantity int
		reservedQuantity  int
		pendingQuantity   int
		decreaseQuantity  int
		wantAvailable     int
		wantReserved      int
		wantPending       int
		wantErr           error
	}{
		{
			name:              "deve reduzir apenas do disponível quando suficiente",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			decreaseQuantity:  5,
			wantAvailable:     5,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve reduzir do disponível e reservado quando disponível insuficiente",
			availableQuantity: 3,
			reservedQuantity:  5,
			pendingQuantity:   0,
			decreaseQuantity:  7,
			wantAvailable:     0,
			wantReserved:      1,
			wantPending:       4,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando estoque total insuficiente",
			availableQuantity: 3,
			reservedQuantity:  2,
			pendingQuantity:   0,
			decreaseQuantity:  10,
			wantAvailable:     3,
			wantReserved:      2,
			wantPending:       0,
			wantErr:           ErrInsufficientStock,
		},
		{
			name:              "deve retornar erro quando quantidade é zero",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			decreaseQuantity:  0,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando quantidade é negativa",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			decreaseQuantity:  -5,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				tt.availableQuantity,
				tt.reservedQuantity,
				tt.pendingQuantity,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.ManualDecrease(tt.decreaseQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ManualDecrease() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ManualDecrease() unexpected error = %v", err)
			}

			if inv.AvailableQuantity() != tt.wantAvailable {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.wantAvailable)
			}
			if inv.ReservedQuantity() != tt.wantReserved {
				t.Errorf("ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.wantReserved)
			}
			if inv.PendingQuantity() != tt.wantPending {
				t.Errorf("PendingQuantity = %v, want %v", inv.PendingQuantity(), tt.wantPending)
			}
		})
	}
}

func TestInventory_ReservedDecrease(t *testing.T) {
	tests := []struct {
		name             string
		reservedQuantity int
		decreaseQuantity int
		wantReserved     int
		wantErr          error
	}{
		{
			name:             "deve reduzir estoque reservado com sucesso",
			reservedQuantity: 10,
			decreaseQuantity: 5,
			wantReserved:     5,
			wantErr:          nil,
		},
		{
			name:             "deve retornar erro quando reservado insuficiente",
			reservedQuantity: 3,
			decreaseQuantity: 5,
			wantReserved:     3,
			wantErr:          ErrInsufficientReserved,
		},
		{
			name:             "deve retornar erro quando quantidade é zero",
			reservedQuantity: 10,
			decreaseQuantity: 0,
			wantReserved:     10,
			wantErr:          ErrInvalidQuantity,
		},
		{
			name:             "deve retornar erro quando quantidade é negativa",
			reservedQuantity: 10,
			decreaseQuantity: -5,
			wantReserved:     10,
			wantErr:          ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				0,
				tt.reservedQuantity,
				0,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.ReservedDecrease(tt.decreaseQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ReservedDecrease() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ReservedDecrease() unexpected error = %v", err)
			}

			if inv.ReservedQuantity() != tt.wantReserved {
				t.Errorf("ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.wantReserved)
			}
		})
	}
}

func TestInventory_AvailableDecrease(t *testing.T) {
	tests := []struct {
		name              string
		availableQuantity int
		decreaseQuantity  int
		wantAvailable     int
		wantErr           error
	}{
		{
			name:              "deve reduzir estoque disponível com sucesso",
			availableQuantity: 10,
			decreaseQuantity:  5,
			wantAvailable:     5,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando disponível insuficiente",
			availableQuantity: 3,
			decreaseQuantity:  5,
			wantAvailable:     3,
			wantErr:           ErrInsufficientAvailable,
		},
		{
			name:              "deve retornar erro quando quantidade é zero",
			availableQuantity: 10,
			decreaseQuantity:  0,
			wantAvailable:     10,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando quantidade é negativa",
			availableQuantity: 10,
			decreaseQuantity:  -5,
			wantAvailable:     10,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				tt.availableQuantity,
				0,
				0,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.AvailableDecrease(tt.decreaseQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("AvailableDecrease() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("AvailableDecrease() unexpected error = %v", err)
			}

			if inv.AvailableQuantity() != tt.wantAvailable {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.wantAvailable)
			}
		})
	}
}

func TestInventory_Reserve(t *testing.T) {
	tests := []struct {
		name              string
		availableQuantity int
		reservedQuantity  int
		pendingQuantity   int
		reserveQuantity   int
		wantAvailable     int
		wantReserved      int
		wantPending       int
		wantErr           error
	}{
		{
			name:              "deve reservar quando disponível suficiente",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			reserveQuantity:   5,
			wantAvailable:     5,
			wantReserved:      10,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve reservar tudo disponível e adicionar diferença a pendente",
			availableQuantity: 3,
			reservedQuantity:  5,
			pendingQuantity:   0,
			reserveQuantity:   7,
			wantAvailable:     0,
			wantReserved:      8,
			wantPending:       4,
			wantErr:           nil,
		},
		{
			name:              "deve adicionar tudo a pendente quando disponível é zero",
			availableQuantity: 0,
			reservedQuantity:  5,
			pendingQuantity:   0,
			reserveQuantity:   5,
			wantAvailable:     0,
			wantReserved:      5,
			wantPending:       5,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando quantidade é zero",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			reserveQuantity:   0,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando quantidade é negativa",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			reserveQuantity:   -5,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				tt.availableQuantity,
				tt.reservedQuantity,
				tt.pendingQuantity,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.Reserve(tt.reserveQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Reserve() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Reserve() unexpected error = %v", err)
			}

			if inv.AvailableQuantity() != tt.wantAvailable {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.wantAvailable)
			}
			if inv.ReservedQuantity() != tt.wantReserved {
				t.Errorf("ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.wantReserved)
			}
			if inv.PendingQuantity() != tt.wantPending {
				t.Errorf("PendingQuantity = %v, want %v", inv.PendingQuantity(), tt.wantPending)
			}
		})
	}
}

func TestInventory_Increase(t *testing.T) {
	tests := []struct {
		name              string
		availableQuantity int
		reservedQuantity  int
		pendingQuantity   int
		increaseQuantity  int
		wantAvailable     int
		wantReserved      int
		wantPending       int
		wantErr           error
	}{
		{
			name:              "deve aumentar disponível quando não há pendências",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			increaseQuantity:  5,
			wantAvailable:     15,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve atender todas pendências e adicionar restante a disponível",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   3,
			increaseQuantity:  7,
			wantAvailable:     14,
			wantReserved:      8,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve atender parcialmente pendências",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   10,
			increaseQuantity:  5,
			wantAvailable:     10,
			wantReserved:      10,
			wantPending:       5,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando quantidade é zero",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			increaseQuantity:  0,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando quantidade é negativa",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			increaseQuantity:  -5,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				tt.availableQuantity,
				tt.reservedQuantity,
				tt.pendingQuantity,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.Increase(tt.increaseQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Increase() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Increase() unexpected error = %v", err)
			}

			if inv.AvailableQuantity() != tt.wantAvailable {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.wantAvailable)
			}
			if inv.ReservedQuantity() != tt.wantReserved {
				t.Errorf("ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.wantReserved)
			}
			if inv.PendingQuantity() != tt.wantPending {
				t.Errorf("PendingQuantity = %v, want %v", inv.PendingQuantity(), tt.wantPending)
			}
		})
	}
}

func TestInventory_CancelReserved(t *testing.T) {
	tests := []struct {
		name              string
		availableQuantity int
		reservedQuantity  int
		pendingQuantity   int
		cancelQuantity    int
		wantAvailable     int
		wantReserved      int
		wantPending       int
		wantErr           error
	}{
		{
			name:              "deve cancelar apenas de pendências quando suficiente",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   7,
			cancelQuantity:    5,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       2,
			wantErr:           nil,
		},
		{
			name:              "deve cancelar todas pendências e restante de reservado",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   3,
			cancelQuantity:    7,
			wantAvailable:     14,
			wantReserved:      1,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve cancelar apenas de reservado quando não há pendências",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			cancelQuantity:    3,
			wantAvailable:     13,
			wantReserved:      2,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando reservado total insuficiente",
			availableQuantity: 10,
			reservedQuantity:  3,
			pendingQuantity:   2,
			cancelQuantity:    10,
			wantAvailable:     10,
			wantReserved:      3,
			wantPending:       2,
			wantErr:           ErrInsufficientReserved,
		},
		{
			name:              "deve retornar erro quando quantidade é zero",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			cancelQuantity:    0,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando quantidade é negativa",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			cancelQuantity:    -5,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				tt.availableQuantity,
				tt.reservedQuantity,
				tt.pendingQuantity,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.CancelReserved(tt.cancelQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("CancelReserved() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("CancelReserved() unexpected error = %v", err)
			}

			if inv.AvailableQuantity() != tt.wantAvailable {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.wantAvailable)
			}
			if inv.ReservedQuantity() != tt.wantReserved {
				t.Errorf("ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.wantReserved)
			}
			if inv.PendingQuantity() != tt.wantPending {
				t.Errorf("PendingQuantity = %v, want %v", inv.PendingQuantity(), tt.wantPending)
			}
		})
	}
}

func TestInventory_CancelConfirmed(t *testing.T) {
	tests := []struct {
		name              string
		availableQuantity int
		reservedQuantity  int
		pendingQuantity   int
		confirmQuantity   int
		wantAvailable     int
		wantReserved      int
		wantPending       int
		wantErr           error
	}{
		{
			name:              "deve confirmar todas pendências e adicionar restante a disponível",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   3,
			confirmQuantity:   7,
			wantAvailable:     14,
			wantReserved:      8,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve confirmar parcialmente pendências",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   10,
			confirmQuantity:   5,
			wantAvailable:     10,
			wantReserved:      10,
			wantPending:       5,
			wantErr:           nil,
		},
		{
			name:              "deve adicionar tudo a disponível quando não há pendências",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			confirmQuantity:   5,
			wantAvailable:     15,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           nil,
		},
		{
			name:              "deve retornar erro quando quantidade é zero",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			confirmQuantity:   0,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
		{
			name:              "deve retornar erro quando quantidade é negativa",
			availableQuantity: 10,
			reservedQuantity:  5,
			pendingQuantity:   0,
			confirmQuantity:   -5,
			wantAvailable:     10,
			wantReserved:      5,
			wantPending:       0,
			wantErr:           ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productID := uuid.New().String()
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				tt.availableQuantity,
				tt.reservedQuantity,
				tt.pendingQuantity,
				time.Now(),
				time.Now(),
				nil,
			)

			err := inv.CancelConfirmed(tt.confirmQuantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("CancelConfirmed() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("CancelConfirmed() unexpected error = %v", err)
			}

			if inv.AvailableQuantity() != tt.wantAvailable {
				t.Errorf("AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.wantAvailable)
			}
			if inv.ReservedQuantity() != tt.wantReserved {
				t.Errorf("ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.wantReserved)
			}
			if inv.PendingQuantity() != tt.wantPending {
				t.Errorf("PendingQuantity = %v, want %v", inv.PendingQuantity(), tt.wantPending)
			}
		})
	}
}

func TestInventory_IsDeleted(t *testing.T) {
	productID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name      string
		deletedAt *time.Time
		want      bool
	}{
		{
			name:      "deve retornar false quando não está deletado",
			deletedAt: nil,
			want:      false,
		},
		{
			name:      "deve retornar true quando está deletado",
			deletedAt: &now,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, _ := ReconstructInventory(
				uuid.New().String(),
				productID,
				10,
				5,
				0,
				now,
				now,
				tt.deletedAt,
			)

			if got := inv.IsDeleted(); got != tt.want {
				t.Errorf("IsDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}
