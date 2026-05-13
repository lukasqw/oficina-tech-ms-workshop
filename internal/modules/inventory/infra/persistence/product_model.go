package persistence

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductModel representa o modelo de persistência para produtos
type ProductModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"not null;size:200"`
	Description string    `gorm:"size:1000"`
	Price       int       `gorm:"not null"`
	ProductType string    `gorm:"not null;size:20"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName especifica o nome da tabela no banco de dados
func (ProductModel) TableName() string {
	return "product_models"
}

// ToDomain converte ProductModel para a entidade Product do domínio
func (m *ProductModel) ToDomain() (*product.Product, error) {
	productType, err := product.NewProductType(m.ProductType)
	if err != nil {
		return nil, err
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return product.ReconstructProduct(
		m.ID.String(),
		m.Name,
		m.Description,
		m.Price,
		productType,
		m.CreatedAt,
		m.UpdatedAt,
		deletedAt,
	)
}

// FromDomain converte a entidade Product do domínio para ProductModel
func FromDomain(p *product.Product) (*ProductModel, error) {
	var id uuid.UUID
	var err error

	if p.ID() != "" {
		id, err = uuid.Parse(p.ID())
		if err != nil {
			return nil, err
		}
	}
	// Se ID vazio, GORM gerará via default:gen_random_uuid()

	model := &ProductModel{
		ID:          id,
		Name:        p.Name(),
		Description: p.Description(),
		Price:       p.Price(),
		ProductType: p.ProductType().Value(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}

	if p.DeletedAt() != nil {
		model.DeletedAt = gorm.DeletedAt{
			Time:  *p.DeletedAt(),
			Valid: true,
		}
	}

	return model, nil
}
