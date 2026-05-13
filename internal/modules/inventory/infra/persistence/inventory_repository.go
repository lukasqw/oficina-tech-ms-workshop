package persistence

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryRepositoryImpl implementa a interface Repository do domínio
type InventoryRepositoryImpl struct {
	db *gorm.DB
}

// NewInventoryRepository cria uma nova instância do repositório de estoque
func NewInventoryRepository(db *gorm.DB) inventory.Repository {
	return &InventoryRepositoryImpl{db: db}
}

// Save persiste um estoque (cria se ID vazio, atualiza caso contrário)
func (r *InventoryRepositoryImpl) Save(ctx context.Context, inv *inventory.Inventory) error {
	model, err := FromDomainInventory(inv)
	if err != nil {
		return err
	}

	// Se ID é vazio, cria novo registro; caso contrário, atualiza
	if inv.ID() == "" {
		// Gera UUID v7
		model.ID = uuid.Must(uuid.NewV7())
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
		if err := inv.SetID(model.ID.String()); err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}

	return nil
}

// FindByID busca um estoque pelo seu identificador único
func (r *InventoryRepositoryImpl) FindByID(ctx context.Context, id string) (*inventory.Inventory, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, inventory.ErrInvalidInventoryID
	}

	var model InventoryModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inventory.ErrInventoryNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByProductID busca um estoque pelo ID do produto
func (r *InventoryRepositoryImpl) FindByProductID(ctx context.Context, productID string) (*inventory.Inventory, error) {
	uid, err := uuid.Parse(productID)
	if err != nil {
		return nil, inventory.ErrInvalidProductID
	}

	var model InventoryModel
	if err := r.db.WithContext(ctx).First(&model, "product_id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inventory.ErrInventoryNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindAll retorna todos os estoques não deletados
func (r *InventoryRepositoryImpl) FindAll(ctx context.Context) ([]*inventory.Inventory, error) {
	var models []InventoryModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	inventories := make([]*inventory.Inventory, 0, len(models))
	for _, model := range models {
		inv, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inv)
	}

	return inventories, nil
}

// Delete marca um estoque como deletado (soft delete)
func (r *InventoryRepositoryImpl) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return inventory.ErrInvalidInventoryID
	}

	var model InventoryModel

	// Primeiro verifica se o registro existe (GORM filtra automaticamente deleted_at IS NULL)
	result := r.db.WithContext(ctx).First(&model, "id = ?", uid)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return inventory.ErrInventoryNotFound
		}
		return result.Error
	}

	// Executa soft delete (GORM define deleted_at automaticamente)
	if err := r.db.WithContext(ctx).Delete(&model).Error; err != nil {
		return err
	}

	return nil
}

// ExistsByProductID verifica se já existe estoque para um produto
func (r *InventoryRepositoryImpl) ExistsByProductID(ctx context.Context, productID string) (bool, error) {
	uid, err := uuid.Parse(productID)
	if err != nil {
		return false, inventory.ErrInvalidProductID
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&InventoryModel{}).Where("product_id = ?", uid).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
