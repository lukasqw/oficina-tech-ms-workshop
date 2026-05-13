package service

import "context"

type Repository interface {
	Save(ctx context.Context, service *Service) error
	FindByID(ctx context.Context, id string) (*Service, error)
	FindAll(ctx context.Context) ([]*Service, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	Delete(ctx context.Context, id string) error
}
