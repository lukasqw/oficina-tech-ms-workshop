package dto

// ServiceDTO is a shared data transfer object for service information
// Used for inter-module communication
type ServiceDTO struct {
	ID          string
	Name        string
	Description string
	Price       int
}
