package product

import "context"

// Repository define a interface para persistência de produtos
type Repository interface {
	// Save persiste um produto (cria se ID vazio, atualiza caso contrário)
	Save(ctx context.Context, product *Product) error

	// FindByID busca um produto pelo seu identificador único
	FindByID(ctx context.Context, id string) (*Product, error)

	// FindAll retorna todos os produtos não deletados
	FindAll(ctx context.Context) ([]*Product, error)

	// Delete marca um produto como deletado (soft delete)
	Delete(ctx context.Context, id string) error
}
