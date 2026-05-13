package service

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"strings"
	"time"
)

type Service struct {
	id          string
	name        string
	description string
	price       int
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// NewService creates a new Service with validation
func NewService(name, description string, price int) (*Service, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if len(name) < 3 || len(name) > 100 {
		return nil, ErrInvalidServiceName
	}

	if len(description) < 10 || len(description) > 500 {
		return nil, ErrInvalidDescription
	}

	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	now := time.Now()
	return &Service{
		name:        name,
		description: description,
		price:       price,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructService reconstructs a Service from persistence
func ReconstructService(id string, name, description string, price int, createdAt, updatedAt time.Time, deletedAt *time.Time) (*Service, error) {
	if err := utils.ValidateUUID(id); err != nil {
		return nil, ErrInvalidServiceID
	}

	return &Service{
		id:          id,
		name:        name,
		description: description,
		price:       price,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}, nil
}

// Getters
func (s *Service) ID() string {
	return s.id
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Description() string {
	return s.description
}

func (s *Service) Price() int {
	return s.price
}

func (s *Service) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Service) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Service) DeletedAt() *time.Time {
	return s.deletedAt
}

// IsDeleted checks if the service has been soft deleted
func (s *Service) IsDeleted() bool {
	return s.deletedAt != nil
}

// MarkAsDeleted marks the service as deleted
func (s *Service) MarkAsDeleted() {
	now := time.Now()
	s.deletedAt = &now
	s.updatedAt = now
}

// SetID sets the ID after persistence
func (s *Service) SetID(id string) error {
	if err := utils.ValidateUUID(id); err != nil {
		return ErrInvalidServiceID
	}
	s.id = id
	return nil
}

// ChangeName updates the service name with validation
func (s *Service) ChangeName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < 3 || len(name) > 100 {
		return ErrInvalidServiceName
	}

	s.name = name
	s.updatedAt = time.Now()
	return nil
}

// ChangeDescription updates the service description with validation
func (s *Service) ChangeDescription(description string) error {
	description = strings.TrimSpace(description)

	if len(description) < 10 || len(description) > 500 {
		return ErrInvalidDescription
	}

	s.description = description
	s.updatedAt = time.Now()
	return nil
}

// ChangePrice updates the service price with validation
func (s *Service) ChangePrice(price int) error {
	if price <= 0 {
		return ErrInvalidPrice
	}

	s.price = price
	s.updatedAt = time.Now()
	return nil
}
