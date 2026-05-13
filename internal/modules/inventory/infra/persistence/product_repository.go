package persistence

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRepositoryImpl implementa a interface Repository do domínio
type ProductRepositoryImpl struct {
	db *gorm.DB
}

// NewProductRepository cria uma nova instância do repositório de produtos
func NewProductRepository(db *gorm.DB) product.Repository {
	return &ProductRepositoryImpl{db: db}
}

// Save persiste um produto (cria se ID vazio, atualiza caso contrário)
func (r *ProductRepositoryImpl) Save(ctx context.Context, p *product.Product) error {
	model, err := FromDomain(p)
	if err != nil {
		return err
	}

	// Se ID é vazio, cria novo registro; caso contrário, atualiza
	if p.ID() == "" {
		// Gera UUID v7
		model.ID = uuid.Must(uuid.NewV7())
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
		if err := p.SetID(model.ID.String()); err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}

	return nil
}

// FindByID busca um produto pelo seu identificador único
func (r *ProductRepositoryImpl) FindByID(ctx context.Context, id string) (*product.Product, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, product.ErrInvalidProductID
	}

	var model ProductModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrProductNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindAll retorna todos os produtos não deletados
func (r *ProductRepositoryImpl) FindAll(ctx context.Context) ([]*product.Product, error) {
	var models []ProductModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	products := make([]*product.Product, 0, len(models))
	for _, model := range models {
		p, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// Delete marca um produto como deletado (soft delete)
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return product.ErrInvalidProductID
	}

	var model ProductModel

	// Primeiro verifica se o registro existe (GORM filtra automaticamente deleted_at IS NULL)
	result := r.db.WithContext(ctx).First(&model, "id = ?", uid)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return product.ErrProductNotFound
		}
		return result.Error
	}

	// Executa soft delete (GORM define deleted_at automaticamente)
	if err := r.db.WithContext(ctx).Delete(&model).Error; err != nil {
		return err
	}

	return nil
}
