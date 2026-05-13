package product

import "fmt"

// ProductType representa o tipo de produto no inventário
type ProductType struct {
	value string
}

// Valores válidos para ProductType
const (
	ProductTypeConsumable = "CONSUMABLE"
	ProductTypeSimple     = "SIMPLE"
	ProductTypeParts      = "PARTS"
)

// NewProductType cria um novo ProductType com validação
func NewProductType(value string) (ProductType, error) {
	switch value {
	case ProductTypeConsumable, ProductTypeSimple, ProductTypeParts:
		return ProductType{value: value}, nil
	default:
		return ProductType{}, ErrInvalidProductType
	}
}

// String retorna a representação em string do ProductType
func (pt ProductType) String() string {
	return pt.value
}

// Value retorna o valor interno do ProductType
func (pt ProductType) Value() string {
	return pt.value
}

// MarshalJSON implementa json.Marshaler
func (pt ProductType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, pt.value)), nil
}
