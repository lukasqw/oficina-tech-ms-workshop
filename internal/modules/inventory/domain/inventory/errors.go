package inventory

import "errors"

var (
	// ErrInvalidInventoryID indica que o ID do estoque é inválido
	ErrInvalidInventoryID = errors.New("ID de estoque inválido")

	// ErrInvalidProductID indica que o ID do produto é inválido
	ErrInvalidProductID = errors.New("ID de produto inválido")

	// ErrInvalidQuantity indica que a quantidade fornecida é inválida
	ErrInvalidQuantity = errors.New("quantidade inválida")

	// ErrInsufficientStock indica que não há estoque suficiente para a operação
	ErrInsufficientStock = errors.New("estoque insuficiente")

	// ErrInsufficientAvailable indica que não há estoque disponível suficiente
	ErrInsufficientAvailable = errors.New("estoque disponível insuficiente")

	// ErrInsufficientReserved indica que não há estoque reservado suficiente
	ErrInsufficientReserved = errors.New("estoque reservado insuficiente")

	// ErrInventoryNotFound indica que o estoque não foi encontrado
	ErrInventoryNotFound = errors.New("estoque não encontrado")

	// ErrProductAlreadyHasInventory indica que o produto já possui registro de estoque
	ErrProductAlreadyHasInventory = errors.New("produto já possui registro de estoque")

	// ErrInvalidOperation indica que o tipo de operação de estoque é inválido
	ErrInvalidOperation = errors.New("tipo de operação de estoque inválido")
)
