package persistence

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"uniqueIndex;not null;size:100"`
	Description string    `gorm:"not null;size:500"`
	Price       int       `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (ServiceModel) TableName() string {
	return "services"
}

func (m *ServiceModel) ToDomain() (*service.Service, error) {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return service.ReconstructService(
		m.ID.String(),
		m.Name,
		m.Description,
		m.Price,
		m.CreatedAt,
		m.UpdatedAt,
		deletedAt,
	)
}

func FromDomain(s *service.Service) (*ServiceModel, error) {
	var id uuid.UUID
	var err error

	if s.ID() != "" {
		id, err = uuid.Parse(s.ID())
		if err != nil {
			return nil, err
		}
	}

	model := &ServiceModel{
		ID:          id,
		Name:        s.Name(),
		Description: s.Description(),
		Price:       s.Price(),
		CreatedAt:   s.CreatedAt(),
		UpdatedAt:   s.UpdatedAt(),
	}

	if s.DeletedAt() != nil {
		model.DeletedAt = gorm.DeletedAt{
			Time:  *s.DeletedAt(),
			Valid: true,
		}
	}

	return model, nil
}
