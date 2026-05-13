package handlers

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
)

// Mock inventory repository
type mockInventoryRepository struct {
	findByIDFunc          func(context.Context, string) (*inventory.Inventory, error)
	findByProductIDFunc   func(context.Context, string) (*inventory.Inventory, error)
	saveFunc              func(context.Context, *inventory.Inventory) error
	deleteFunc            func(context.Context, string) error
	existsByProductIDFunc func(context.Context, string) (bool, error)
	findAllFunc           func(context.Context) ([]*inventory.Inventory, error)
}

func (m *mockInventoryRepository) FindByID(ctx context.Context, id string) (*inventory.Inventory, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInventoryRepository) FindByProductID(ctx context.Context, productID string) (*inventory.Inventory, error) {
	if m.findByProductIDFunc != nil {
		return m.findByProductIDFunc(ctx, productID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInventoryRepository) Save(ctx context.Context, inv *inventory.Inventory) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, inv)
	}
	return errors.New("not implemented")
}

func (m *mockInventoryRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockInventoryRepository) ExistsByProductID(ctx context.Context, productID string) (bool, error) {
	if m.existsByProductIDFunc != nil {
		return m.existsByProductIDFunc(ctx, productID)
	}
	return false, errors.New("not implemented")
}

func (m *mockInventoryRepository) FindAll(ctx context.Context) ([]*inventory.Inventory, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

// Mock product repository
type mockProductRepository struct {
	findByIDFunc     func(context.Context, string) (*product.Product, error)
	existsFunc       func(context.Context, string) (bool, error)
	saveFunc         func(context.Context, *product.Product) error
	findAllFunc      func(context.Context) ([]*product.Product, error)
	deleteFunc       func(context.Context, string) error
	existsByNameFunc func(context.Context, string) (bool, error)
}

func (m *mockProductRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockProductRepository) Exists(ctx context.Context, id string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, id)
	}
	return false, errors.New("not implemented")
}

func (m *mockProductRepository) Save(ctx context.Context, p *product.Product) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, p)
	}
	return errors.New("not implemented")
}

func (m *mockProductRepository) FindAll(ctx context.Context) ([]*product.Product, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockProductRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockProductRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	if m.existsByNameFunc != nil {
		return m.existsByNameFunc(ctx, name)
	}
	return false, errors.New("not implemented")
}
