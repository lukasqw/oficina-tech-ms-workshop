package dto

// ProductDTO is a shared data transfer object for product information
// Used for inter-module communication
type ProductDTO struct {
	ID          string
	Name        string
	Description string
	Price       int
	ProductType string
}
