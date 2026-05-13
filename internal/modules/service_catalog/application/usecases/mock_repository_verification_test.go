package usecases

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
)

// TestMockRepository_Save verifies the mock Save method works correctly
func TestMockRepository_Save(t *testing.T) {
	t.Run("default behavior returns nil", func(t *testing.T) {
		mock := &mockServiceRepository{}
		svc, _ := service.NewService("Test Service", "Test description for service", 10000)

		err := mock.Save(context.Background(), svc)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("configured behavior is called", func(t *testing.T) {
		expectedErr := errors.New("save error")
		called := false

		mock := &mockServiceRepository{
			saveFunc: func(_ context.Context, s *service.Service) error {
				called = true
				return expectedErr
			},
		}

		svc, _ := service.NewService("Test Service", "Test description for service", 10000)
		err := mock.Save(context.Background(), svc)

		if !called {
			t.Error("saveFunc was not called")
		}
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestMockRepository_FindByID verifies the mock FindByID method works correctly
func TestMockRepository_FindByID(t *testing.T) {
	t.Run("default behavior returns nil", func(t *testing.T) {
		mock := &mockServiceRepository{}

		result, err := mock.FindByID(context.Background(), "test-id")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("configured behavior is called", func(t *testing.T) {
		expectedService, _ := service.NewService("Test Service", "Test description for service", 10000)
		called := false

		mock := &mockServiceRepository{
			findByIDFunc: func(_ context.Context, id string) (*service.Service, error) {
				called = true
				if id == "valid-id" {
					return expectedService, nil
				}
				return nil, service.ErrServiceNotFound
			},
		}

		result, err := mock.FindByID(context.Background(), "valid-id")

		if !called {
			t.Error("findByIDFunc was not called")
		}
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result != expectedService {
			t.Error("expected service to match")
		}
	})
}

// TestMockRepository_FindAll verifies the mock FindAll method works correctly
func TestMockRepository_FindAll(t *testing.T) {
	t.Run("default behavior returns empty slice", func(t *testing.T) {
		mock := &mockServiceRepository{}

		result, err := mock.FindAll(context.Background())
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result == nil {
			t.Error("expected non-nil slice")
		}
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("configured behavior is called", func(t *testing.T) {
		svc1, _ := service.NewService("Service 1", "Description for service 1", 10000)
		svc2, _ := service.NewService("Service 2", "Description for service 2", 20000)
		expectedServices := []*service.Service{svc1, svc2}
		called := false

		mock := &mockServiceRepository{
			findAllFunc: func(_ context.Context) ([]*service.Service, error) {
				called = true
				return expectedServices, nil
			},
		}

		result, err := mock.FindAll(context.Background())

		if !called {
			t.Error("findAllFunc was not called")
		}
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 services, got %d", len(result))
		}
	})
}

// TestMockRepository_ExistsByName verifies the mock ExistsByName method works correctly
func TestMockRepository_ExistsByName(t *testing.T) {
	t.Run("default behavior returns false", func(t *testing.T) {
		mock := &mockServiceRepository{}

		result, err := mock.ExistsByName(context.Background(), "test-name")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result {
			t.Error("expected false, got true")
		}
	})

	t.Run("configured behavior is called", func(t *testing.T) {
		called := false

		mock := &mockServiceRepository{
			existsByNameFunc: func(_ context.Context, name string) (bool, error) {
				called = true
				return name == "existing-name", nil
			},
		}

		result, err := mock.ExistsByName(context.Background(), "existing-name")

		if !called {
			t.Error("existsByNameFunc was not called")
		}
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if !result {
			t.Error("expected true, got false")
		}
	})
}

// TestMockRepository_Delete verifies the mock Delete method works correctly
func TestMockRepository_Delete(t *testing.T) {
	t.Run("default behavior returns nil", func(t *testing.T) {
		mock := &mockServiceRepository{}

		err := mock.Delete(context.Background(), "test-id")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("configured behavior is called", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		called := false

		mock := &mockServiceRepository{
			deleteFunc: func(_ context.Context, id string) error {
				called = true
				return expectedErr
			},
		}

		err := mock.Delete(context.Background(), "test-id")

		if !called {
			t.Error("deleteFunc was not called")
		}
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
