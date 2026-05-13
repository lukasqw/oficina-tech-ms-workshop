package product

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"strings"
	"time"
)

// Product representa um item no inventário da oficina
type Product struct {
	id          string
	name        string
	description string
	price       int // Preço em centavos
	productType ProductType
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// NewProduct cria um novo produto com validação
func NewProduct(name, description string, price int, productType ProductType) (*Product, error) {
	// Validar nome
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 200 {
		return nil, ErrInvalidProductName
	}

	// Validar descrição
	if len(description) > 1000 {
		return nil, ErrInvalidDescription
	}

	// Validar preço
	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	now := time.Now()
	return &Product{
		name:        name,
		description: description,
		price:       price,
		productType: productType,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructProduct reconstrói um produto a partir da persistência
func ReconstructProduct(
	id string,
	name, description string,
	price int,
	productType ProductType,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) (*Product, error) {
	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		return nil, ErrInvalidProductID
	}

	return &Product{
		id:          id,
		name:        name,
		description: description,
		price:       price,
		productType: productType,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}, nil
}

// Getters

// ID retorna o identificador único do produto
func (p *Product) ID() string {
	return p.id
}

// Name retorna o nome do produto
func (p *Product) Name() string {
	return p.name
}

// Description retorna a descrição do produto
func (p *Product) Description() string {
	return p.description
}

// Price retorna o preço do produto em centavos
func (p *Product) Price() int {
	return p.price
}

// ProductType retorna o tipo do produto
func (p *Product) ProductType() ProductType {
	return p.productType
}

// CreatedAt retorna a data de criação do produto
func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt retorna a data da última atualização do produto
func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

// DeletedAt retorna a data de exclusão do produto (soft delete)
func (p *Product) DeletedAt() *time.Time {
	return p.deletedAt
}

// IsDeleted verifica se o produto foi marcado como deletado
func (p *Product) IsDeleted() bool {
	return p.deletedAt != nil
}

// SetID define o ID do produto após persistência
func (p *Product) SetID(id string) error {
	if err := utils.ValidateUUID(id); err != nil {
		return ErrInvalidProductID
	}
	p.id = id
	return nil
}

// Métodos de negócio

// ChangeName altera o nome do produto com validação
func (p *Product) ChangeName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 200 {
		return ErrInvalidProductName
	}
	p.name = name
	p.updatedAt = time.Now()
	return nil
}

// ChangeDescription altera a descrição do produto
func (p *Product) ChangeDescription(description string) error {
	if len(description) > 1000 {
		return ErrInvalidDescription
	}
	p.description = description
	p.updatedAt = time.Now()
	return nil
}

// ChangePrice altera o preço do produto com validação
func (p *Product) ChangePrice(price int) error {
	if price <= 0 {
		return ErrInvalidPrice
	}
	p.price = price
	p.updatedAt = time.Now()
	return nil
}

// ChangeProductType altera o tipo do produto
func (p *Product) ChangeProductType(productType ProductType) {
	p.productType = productType
	p.updatedAt = time.Now()
}

// MarkAsDeleted marca o produto como deletado (soft delete)
func (p *Product) MarkAsDeleted() {
	now := time.Now()
	p.deletedAt = &now
	p.updatedAt = now
}
