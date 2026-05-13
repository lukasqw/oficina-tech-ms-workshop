package product

import "errors"

var (
	// ErrProductNotFound indica que o produto não foi encontrado
	ErrProductNotFound = errors.New("product not found")

	// ErrInvalidProductID indica que o ID do produto é inválido
	ErrInvalidProductID = errors.New("invalid product ID format")

	// ErrInvalidProductName indica que o nome do produto é inválido
	ErrInvalidProductName = errors.New("invalid product name: must be between 2 and 200 characters")

	// ErrInvalidPrice indica que o preço é inválido
	ErrInvalidPrice = errors.New("price must be greater than zero")

	// ErrInvalidDescription indica que a descrição é muito longa
	ErrInvalidDescription = errors.New("description too long: maximum 1000 characters")

	// ErrInvalidProductType indica que o tipo de produto é inválido
	ErrInvalidProductType = errors.New("invalid product type: must be CONSUMABLE, SIMPLE, or PARTS")
)
