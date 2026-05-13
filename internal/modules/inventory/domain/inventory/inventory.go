package inventory

import (
	"time"

	"github.com/google/uuid"
)

// Inventory representa a entidade de domínio para controle de estoque
type Inventory struct {
	id                string
	productID         string
	availableQuantity int
	reservedQuantity  int
	pendingQuantity   int
	createdAt         time.Time
	updatedAt         time.Time
	deletedAt         *time.Time
}

// NewInventory cria uma nova instância de Inventory com quantidades zeradas
func NewInventory(productID string) (*Inventory, error) {
	if productID == "" {
		return nil, ErrInvalidProductID
	}

	// Valida se productID é um UUID válido
	if _, err := uuid.Parse(productID); err != nil {
		return nil, ErrInvalidProductID
	}

	now := time.Now()
	return &Inventory{
		productID:         productID,
		availableQuantity: 0,
		reservedQuantity:  0,
		pendingQuantity:   0,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

// ReconstructInventory reconstrói uma entidade Inventory a partir da persistência
func ReconstructInventory(
	id string,
	productID string,
	availableQuantity int,
	reservedQuantity int,
	pendingQuantity int,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Inventory, error) {
	if id == "" {
		return nil, ErrInvalidInventoryID
	}
	if productID == "" {
		return nil, ErrInvalidProductID
	}
	if availableQuantity < 0 || reservedQuantity < 0 || pendingQuantity < 0 {
		return nil, ErrInvalidQuantity
	}

	return &Inventory{
		id:                id,
		productID:         productID,
		availableQuantity: availableQuantity,
		reservedQuantity:  reservedQuantity,
		pendingQuantity:   pendingQuantity,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		deletedAt:         deletedAt,
	}, nil
}

// Getters

func (i *Inventory) ID() string {
	return i.id
}

func (i *Inventory) ProductID() string {
	return i.productID
}

func (i *Inventory) AvailableQuantity() int {
	return i.availableQuantity
}

func (i *Inventory) ReservedQuantity() int {
	return i.reservedQuantity
}

func (i *Inventory) PendingQuantity() int {
	return i.pendingQuantity
}

func (i *Inventory) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Inventory) UpdatedAt() time.Time {
	return i.updatedAt
}

func (i *Inventory) DeletedAt() *time.Time {
	return i.deletedAt
}

func (i *Inventory) IsDeleted() bool {
	return i.deletedAt != nil
}

// SetID define o ID após persistência
func (i *Inventory) SetID(id string) error {
	if id == "" {
		return ErrInvalidInventoryID
	}
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidInventoryID
	}
	i.id = id
	return nil
}

// ManualDecrease reduz estoque priorizando disponível, depois reservado
// Se insuficiente, adiciona diferença a pendente
func (i *Inventory) ManualDecrease(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	totalStock := i.availableQuantity + i.reservedQuantity
	if totalStock < quantity {
		return ErrInsufficientStock
	}

	remaining := quantity

	// Primeiro reduz do disponível
	if i.availableQuantity >= remaining {
		i.availableQuantity -= remaining
	} else {
		remaining -= i.availableQuantity
		i.availableQuantity = 0

		// Reduz o restante do reservado e adiciona a diferença ao pendente
		i.reservedQuantity -= remaining
		i.pendingQuantity += remaining
	}

	i.updatedAt = time.Now()
	return nil
}

// ReservedDecrease reduz apenas estoque reservado
func (i *Inventory) ReservedDecrease(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if i.reservedQuantity < quantity {
		return ErrInsufficientReserved
	}

	i.reservedQuantity -= quantity
	i.updatedAt = time.Now()
	return nil
}

// AvailableDecrease reduz apenas estoque disponível
func (i *Inventory) AvailableDecrease(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if i.availableQuantity < quantity {
		return ErrInsufficientAvailable
	}

	i.availableQuantity -= quantity
	i.updatedAt = time.Now()
	return nil
}

// Reserve transfere de disponível para reservado
// Se insuficiente, transfere tudo disponível e adiciona diferença a pendente
func (i *Inventory) Reserve(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if i.availableQuantity >= quantity {
		// Há estoque disponível suficiente
		i.availableQuantity -= quantity
		i.reservedQuantity += quantity
	} else {
		// Estoque disponível insuficiente
		remaining := quantity - i.availableQuantity
		i.reservedQuantity += i.availableQuantity
		i.availableQuantity = 0
		i.pendingQuantity += remaining
	}

	i.updatedAt = time.Now()
	return nil
}

// Increase aumenta estoque priorizando atender pendências
func (i *Inventory) Increase(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if i.pendingQuantity > 0 {
		if i.pendingQuantity >= quantity {
			// Toda a quantidade vai para atender pendências
			i.pendingQuantity -= quantity
			i.reservedQuantity += quantity
		} else {
			// Atende todas as pendências e o restante vai para disponível
			remaining := quantity - i.pendingQuantity
			i.reservedQuantity += i.pendingQuantity
			i.pendingQuantity = 0
			i.availableQuantity += remaining
		}
	} else {
		// Não há pendências, tudo vai para disponível
		i.availableQuantity += quantity
	}

	i.updatedAt = time.Now()
	return nil
}

// CancelReserved cancela reservas priorizando remoção de pendências
func (i *Inventory) CancelReserved(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	totalReserved := i.pendingQuantity + i.reservedQuantity
	if totalReserved < quantity {
		return ErrInsufficientReserved
	}

	if i.pendingQuantity >= quantity {
		// Toda a quantidade vem de pendências
		i.pendingQuantity -= quantity
	} else {
		// Remove todas as pendências e o restante vem de reservado
		remaining := quantity - i.pendingQuantity
		i.pendingQuantity = 0
		i.reservedQuantity -= remaining
		i.availableQuantity += remaining
	}

	i.updatedAt = time.Now()
	return nil
}

// CancelConfirmed confirma pedidos pendentes movendo para reservado
func (i *Inventory) CancelConfirmed(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if i.pendingQuantity > 0 {
		if i.pendingQuantity >= quantity {
			// Toda a quantidade vem de pendências
			i.pendingQuantity -= quantity
			i.reservedQuantity += quantity
		} else {
			// Atende todas as pendências e o restante vai para disponível
			remaining := quantity - i.pendingQuantity
			i.reservedQuantity += i.pendingQuantity
			i.pendingQuantity = 0
			i.availableQuantity += remaining
		}
	} else {
		// Não há pendências, tudo vai para disponível
		i.availableQuantity += quantity
	}

	i.updatedAt = time.Now()
	return nil
}
