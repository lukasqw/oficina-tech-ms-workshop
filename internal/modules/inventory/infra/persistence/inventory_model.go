package persistence

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryModel representa o modelo de persistência para estoque
type InventoryModel struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ProductID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"`
	AvailableQuantity int            `gorm:"not null;default:0"`
	ReservedQuantity  int            `gorm:"not null;default:0"`
	PendingQuantity   int            `gorm:"not null;default:0"`
	CreatedAt         time.Time      `gorm:"not null"`
	UpdatedAt         time.Time      `gorm:"not null"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

// TableName especifica o nome da tabela no banco de dados
func (InventoryModel) TableName() string {
	return "inventories"
}

// ToDomain converte InventoryModel para a entidade Inventory do domínio
func (m *InventoryModel) ToDomain() (*inventory.Inventory, error) {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return inventory.ReconstructInventory(
		m.ID.String(),
		m.ProductID.String(),
		m.AvailableQuantity,
		m.ReservedQuantity,
		m.PendingQuantity,
		m.CreatedAt,
		m.UpdatedAt,
		deletedAt,
	)
}

// FromDomainInventory converte a entidade Inventory do domínio para InventoryModel
func FromDomainInventory(inv *inventory.Inventory) (*InventoryModel, error) {
	var id uuid.UUID
	var err error

	if inv.ID() != "" {
		id, err = uuid.Parse(inv.ID())
		if err != nil {
			return nil, err
		}
	}
	// Se ID vazio, será gerado no repositório usando UUID v7

	productID, err := uuid.Parse(inv.ProductID())
	if err != nil {
		return nil, err
	}

	model := &InventoryModel{
		ID:                id,
		ProductID:         productID,
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		CreatedAt:         inv.CreatedAt(),
		UpdatedAt:         inv.UpdatedAt(),
	}

	if inv.DeletedAt() != nil {
		model.DeletedAt = gorm.DeletedAt{
			Time:  *inv.DeletedAt(),
			Valid: true,
		}
	}

	return model, nil
}
