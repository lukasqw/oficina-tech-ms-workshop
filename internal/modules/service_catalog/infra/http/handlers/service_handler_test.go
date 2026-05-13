package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockServiceRepository is a mock implementation of service.Repository for handler testing
type mockServiceRepository struct {
	saveFunc         func(context.Context, *service.Service) error
	findByIDFunc     func(context.Context, string) (*service.Service, error)
	findAllFunc      func(context.Context) ([]*service.Service, error)
	existsByNameFunc func(context.Context, string) (bool, error)
	deleteFunc       func(context.Context, string) error
}

func (m *mockServiceRepository) Save(ctx context.Context, s *service.Service) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, s)
	}
	return nil
}

func (m *mockServiceRepository) FindByID(ctx context.Context, id string) (*service.Service, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockServiceRepository) FindAll(ctx context.Context) ([]*service.Service, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return []*service.Service{}, nil
}

func (m *mockServiceRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	if m.existsByNameFunc != nil {
		return m.existsByNameFunc(ctx, name)
	}
	return false, nil
}

func (m *mockServiceRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// assertErrorResponse checks if the response contains the expected error code
func assertErrorResponse(t *testing.T, response map[string]interface{}, expectedError string) {
	t.Helper()
	if errors, ok := response["errors"].([]interface{}); ok && len(errors) > 0 {
		if errorData, ok := errors[0].(map[string]interface{}); ok {
			if errorData["code"] != expectedError {
				t.Errorf("expected error code %s, got %v", expectedError, errorData["code"])
			}
		} else {
			t.Errorf("expected error object, got %v", errors[0])
		}
	} else {
		t.Errorf("expected errors array in response, got %v", response)
	}
}

func TestServiceHandler_CreateService(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mockServiceRepository)
		expectedStatus int
		expectedError  string
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "success - create service",
			requestBody: map[string]interface{}{
				"name":        "Troca de Óleo",
				"description": "Troca de óleo do motor com filtro",
				"price":       15000,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					_ = s.SetID(utils.GenerateUUIDv7())
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				if data["name"] != "Troca de Óleo" {
					t.Errorf("expected name 'Troca de Óleo', got %v", data["name"])
				}
				if data["price"] != float64(15000) {
					t.Errorf("expected price 15000, got %v", data["price"])
				}
			},
		},
		{
			name:           "error - invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeInvalidRequest,
		},
		{
			name: "error - missing required field name",
			requestBody: map[string]interface{}{
				"description": "Troca de óleo do motor com filtro",
				"price":       15000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name: "error - name too short",
			requestBody: map[string]interface{}{
				"name":        "AB",
				"description": "Troca de óleo do motor com filtro",
				"price":       15000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name: "error - description too short",
			requestBody: map[string]interface{}{
				"name":        "Troca de Óleo",
				"description": "Curto",
				"price":       15000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name: "error - invalid price (zero)",
			requestBody: map[string]interface{}{
				"name":        "Troca de Óleo",
				"description": "Troca de óleo do motor com filtro",
				"price":       0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name: "error - duplicate service name",
			requestBody: map[string]interface{}{
				"name":        "Troca de Óleo",
				"description": "Troca de óleo do motor com filtro",
				"price":       15000,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return true, nil
				}
			},
			expectedStatus: http.StatusConflict,
			expectedError:  utils.ErrCodeDuplicateResource,
		},
		{
			name: "error - repository error",
			requestBody: map[string]interface{}{
				"name":        "Troca de Óleo",
				"description": "Troca de óleo do motor com filtro",
				"price":       15000,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  utils.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewServiceHandler(mockRepo)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateService(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assertErrorResponse(t, response, tt.expectedError)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, response)
			}
		})
	}
}

func TestServiceHandler_GetAllServices(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*mockServiceRepository)
		expectedStatus int
		expectedError  string
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "success - get all services",
			mockSetup: func(m *mockServiceRepository) {
				m.findAllFunc = func(_ context.Context) ([]*service.Service, error) {
					svc1, _ := service.ReconstructService(
						utils.GenerateUUIDv7(),
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					svc2, _ := service.ReconstructService(
						utils.GenerateUUIDv7(),
						"Alinhamento",
						"Alinhamento e balanceamento",
						8000,
						time.Now(),
						time.Now(),
						nil,
					)
					return []*service.Service{svc1, svc2}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				if len(data) != 2 {
					t.Errorf("expected 2 services, got %d", len(data))
				}
			},
		},
		{
			name: "success - empty list",
			mockSetup: func(m *mockServiceRepository) {
				m.findAllFunc = func(_ context.Context) ([]*service.Service, error) {
					return []*service.Service{}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				if len(data) != 0 {
					t.Errorf("expected empty list, got %d items", len(data))
				}
			},
		},
		{
			name: "error - repository error",
			mockSetup: func(m *mockServiceRepository) {
				m.findAllFunc = func(_ context.Context) ([]*service.Service, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  utils.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewServiceHandler(mockRepo)
			req := httptest.NewRequest(http.MethodGet, "/services", nil)
			w := httptest.NewRecorder()

			handler.GetAllServices(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assertErrorResponse(t, response, tt.expectedError)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, response)
			}
		})
	}
}

func TestServiceHandler_GetServiceByID(t *testing.T) {
	validID := utils.GenerateUUIDv7()

	tests := []struct {
		name           string
		serviceID      string
		mockSetup      func(*mockServiceRepository)
		expectedStatus int
		expectedError  string
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:      "success - get service by ID",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.ReconstructService(
						id,
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					return svc, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				if data["name"] != "Troca de Óleo" {
					t.Errorf("expected name 'Troca de Óleo', got %v", data["name"])
				}
			},
		},
		{
			name:           "error - invalid UUID format",
			serviceID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeInvalidUUID,
		},
		{
			name:      "error - service not found",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, service.ErrServiceNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  utils.ErrCodeNotFound,
		},
		{
			name:      "error - repository error",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  utils.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewServiceHandler(mockRepo)
			req := httptest.NewRequest(http.MethodGet, "/services/"+tt.serviceID, nil)
			req.SetPathValue("id", tt.serviceID)
			w := httptest.NewRecorder()

			handler.GetServiceByID(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assertErrorResponse(t, response, tt.expectedError)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, response)
			}
		})
	}
}

func TestServiceHandler_UpdateService(t *testing.T) {
	validID := utils.GenerateUUIDv7()
	name := "Troca de Óleo Atualizada"
	description := "Descrição atualizada do serviço"
	price := 18000

	tests := []struct {
		name           string
		serviceID      string
		requestBody    interface{}
		mockSetup      func(*mockServiceRepository)
		expectedStatus int
		expectedError  string
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:      "success - update all fields",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"name":        name,
				"description": description,
				"price":       price,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.ReconstructService(
						id,
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					return svc, nil
				}
				m.existsByNameFunc = func(_ context.Context, n string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				if data["message"] != "Service updated successfully" {
					t.Errorf("expected success message, got %v", data["message"])
				}
			},
		},
		{
			name:      "success - partial update (name only)",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"name": name,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.ReconstructService(
						id,
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					return svc, nil
				}
				m.existsByNameFunc = func(_ context.Context, n string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - invalid UUID format",
			serviceID:      "invalid-uuid",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeInvalidUUID,
		},
		{
			name:           "error - invalid JSON",
			serviceID:      validID,
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeInvalidRequest,
		},
		{
			name:      "error - name too short",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"name": "AB",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name:      "error - description too short",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"description": "Curto",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name:      "error - invalid price",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"price": 0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
		{
			name:      "error - service not found",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"name": name,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, service.ErrServiceNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  utils.ErrCodeNotFound,
		},
		{
			name:      "error - duplicate service name",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"name": name,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.ReconstructService(
						id,
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					return svc, nil
				}
				m.existsByNameFunc = func(_ context.Context, n string) (bool, error) {
					return true, nil
				}
			},
			expectedStatus: http.StatusConflict,
			expectedError:  utils.ErrCodeDuplicateResource,
		},
		{
			name:      "error - repository error",
			serviceID: validID,
			requestBody: map[string]interface{}{
				"name": name,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  utils.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewServiceHandler(mockRepo)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/services/"+tt.serviceID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tt.serviceID)
			w := httptest.NewRecorder()

			handler.UpdateService(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assertErrorResponse(t, response, tt.expectedError)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, response)
			}
		})
	}
}

func TestServiceHandler_DeleteService(t *testing.T) {
	validID := utils.GenerateUUIDv7()

	tests := []struct {
		name           string
		serviceID      string
		mockSetup      func(*mockServiceRepository)
		expectedStatus int
		expectedError  string
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:      "success - delete service",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.ReconstructService(
						id,
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					return svc, nil
				}
				m.deleteFunc = func(_ context.Context, id string) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				if data["message"] != "Service deleted successfully" {
					t.Errorf("expected success message, got %v", data["message"])
				}
			},
		},
		{
			name:           "error - invalid UUID format",
			serviceID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeInvalidUUID,
		},
		{
			name:      "error - service not found",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, service.ErrServiceNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  utils.ErrCodeNotFound,
		},
		{
			name:      "error - repository error on find",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  utils.ErrCodeNotFound,
		},
		{
			name:      "error - repository error on delete",
			serviceID: validID,
			mockSetup: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.ReconstructService(
						id,
						"Troca de Óleo",
						"Troca de óleo do motor",
						15000,
						time.Now(),
						time.Now(),
						nil,
					)
					return svc, nil
				}
				m.deleteFunc = func(_ context.Context, id string) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  utils.ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewServiceHandler(mockRepo)
			req := httptest.NewRequest(http.MethodDelete, "/services/"+tt.serviceID, nil)
			req.SetPathValue("id", tt.serviceID)
			w := httptest.NewRecorder()

			handler.DeleteService(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assertErrorResponse(t, response, tt.expectedError)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, response)
			}
		})
	}
}

// TestServiceHandler_CreateService_EdgeCases tests edge cases for CreateService
func TestServiceHandler_CreateService_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mockServiceRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "edge case - name at minimum length (3 chars)",
			requestBody: map[string]interface{}{
				"name":        "ABC",
				"description": "Descrição válida com mais de 10 caracteres",
				"price":       100,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					_ = s.SetID(utils.GenerateUUIDv7())
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "edge case - name at maximum length (100 chars)",
			requestBody: map[string]interface{}{
				"name":        strings.Repeat("A", 100),
				"description": "Descrição válida com mais de 10 caracteres",
				"price":       100,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					_ = s.SetID(utils.GenerateUUIDv7())
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "edge case - description at minimum length (10 chars)",
			requestBody: map[string]interface{}{
				"name":        "Serviço Teste",
				"description": "1234567890",
				"price":       100,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					_ = s.SetID(utils.GenerateUUIDv7())
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "edge case - description at maximum length (500 chars)",
			requestBody: map[string]interface{}{
				"name":        "Serviço Teste",
				"description": strings.Repeat("A", 500),
				"price":       100,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					_ = s.SetID(utils.GenerateUUIDv7())
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "edge case - price at minimum valid value (1)",
			requestBody: map[string]interface{}{
				"name":        "Serviço Teste",
				"description": "Descrição válida com mais de 10 caracteres",
				"price":       1,
			},
			mockSetup: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}
				m.saveFunc = func(_ context.Context, s *service.Service) error {
					_ = s.SetID(utils.GenerateUUIDv7())
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "edge case - negative price",
			requestBody: map[string]interface{}{
				"name":        "Serviço Teste",
				"description": "Descrição válida com mais de 10 caracteres",
				"price":       -100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  utils.ErrCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewServiceHandler(mockRepo)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateService(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assertErrorResponse(t, response, tt.expectedError)
			}
		})
	}
}
