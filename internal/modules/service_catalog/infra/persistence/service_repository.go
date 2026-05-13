package persistence

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceRepositoryImpl struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) service.Repository {
	return &ServiceRepositoryImpl{db: db}
}

func (r *ServiceRepositoryImpl) Save(ctx context.Context, s *service.Service) error {
	model, err := FromDomain(s)
	if err != nil {
		return err
	}

	// If ID is empty, create new record; otherwise update
	if s.ID() == "" {
		// Generate UUID v7
		model.ID = uuid.Must(uuid.NewV7())
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return service.ErrDuplicateServiceName
			}
			return err
		}
		if err := s.SetID(model.ID.String()); err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return service.ErrDuplicateServiceName
			}
			return err
		}
	}

	return nil
}

func (r *ServiceRepositoryImpl) FindByID(ctx context.Context, id string) (*service.Service, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, service.ErrInvalidServiceID
	}

	var model ServiceModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrServiceNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

func (r *ServiceRepositoryImpl) FindAll(ctx context.Context) ([]*service.Service, error) {
	var models []ServiceModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	services := make([]*service.Service, 0, len(models))
	for _, model := range models {
		svc, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}

	return services, nil
}

func (r *ServiceRepositoryImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ServiceModel{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ServiceRepositoryImpl) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return service.ErrInvalidServiceID
	}

	var model ServiceModel

	// First, check if the record exists (GORM automatically filters deleted_at IS NULL)
	result := r.db.WithContext(ctx).First(&model, "id = ?", uid)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return service.ErrServiceNotFound
		}
		return result.Error
	}

	// Perform soft delete (GORM sets deleted_at automatically)
	if err := r.db.WithContext(ctx).Delete(&model).Error; err != nil {
		return err
	}

	return nil
}
