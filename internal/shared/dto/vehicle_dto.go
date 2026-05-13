package dto

// VehicleDTO is a shared data transfer object for vehicle information
// Used for inter-module communication
type VehicleDTO struct {
	ID              string
	CustomerID      string
	LicensePlate    string
	Brand           string
	Model           string
	ModelYear       int
	ManufactureYear int
}
