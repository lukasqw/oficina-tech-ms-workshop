package dto

// CustomerDTO is a shared data transfer object for customer information
// Used for inter-module communication
type CustomerDTO struct {
	ID    string
	Name  string
	Email string
	Phone string
}
