package inventory

import "context"

// Repository define a interface para operações de persistência de estoque
type Repository interface {
	// Save persiste um registro de estoque (cria ou atualiza)
	Save(ctx context.Context, inventory *Inventory) error

	// FindByID busca um estoque por ID
	FindByID(ctx context.Context, id string) (*Inventory, error)

	// FindByProductID busca um estoque por ID do produto
	FindByProductID(ctx context.Context, productID string) (*Inventory, error)

	// FindAll retorna todos os estoques não deletados
	FindAll(ctx context.Context) ([]*Inventory, error)

	// Delete realiza soft delete de um estoque
	Delete(ctx context.Context, id string) error

	// ExistsByProductID verifica se já existe estoque para um produto
	ExistsByProductID(ctx context.Context, productID string) (bool, error)
}
