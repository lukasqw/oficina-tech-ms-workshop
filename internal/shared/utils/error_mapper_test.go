package utils

import (
	"errors"
	"net/http"
	"testing"
)

func TestMapDomainError_NotFoundErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedCode string
		expectedStatus int
	}{
		{
			name:        "customer not found",
			err:         errors.New("customer not found"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "vehicle not found",
			err:         errors.New("vehicle not found"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "service not found",
			err:         errors.New("service not found"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "product not found",
			err:         errors.New("product not found"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "user not found",
			err:         errors.New("user not found"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "ordem de serviço não encontrada",
			err:         errors.New("ordem de serviço não encontrada"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "produto não encontrado",
			err:         errors.New("produto não encontrado"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "serviço não encontrado",
			err:         errors.New("serviço não encontrado"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "item não encontrado na ordem",
			err:         errors.New("item não encontrado na ordem"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "inventário do produto não encontrado",
			err:         errors.New("inventário do produto não encontrado"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "ID do cliente inválido",
			err:         errors.New("ID do cliente inválido"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "ID do veículo inválido",
			err:         errors.New("ID do veículo inválido"),
			expectedCode: ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapDomainError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
		})
	}
}

func TestMapDomainError_DuplicateErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedCode string
		expectedStatus int
	}{
		{
			name:        "email already exists",
			err:         errors.New("email already exists"),
			expectedCode: ErrCodeDuplicateResource,
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "user already exists",
			err:         errors.New("user already exists"),
			expectedCode: ErrCodeDuplicateResource,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapDomainError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
		})
	}
}

func TestMapDomainError_AuthenticationErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedCode string
		expectedStatus int
	}{
		{
			name:        "invalid credentials",
			err:         errors.New("invalid credentials"),
			expectedCode: ErrCodeInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "credenciais inválidas",
			err:         errors.New("credenciais inválidas"),
			expectedCode: ErrCodeInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapDomainError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
		})
	}
}

func TestMapDomainError_AuthorizationErrors(t *testing.T) {
	err := errors.New("acesso não autorizado a esta ordem de serviço")
	result := MapDomainError(err)

	if result.Code != ErrCodeUnauthorized {
		t.Errorf("Expected code %s, got %s", ErrCodeUnauthorized, result.Code)
	}

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, result.StatusCode)
	}
}

func TestMapDomainError_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedCode string
		expectedStatus int
	}{
		{
			name:        "veículo não pertence ao cliente",
			err:         errors.New("veículo não pertence ao cliente da ordem de serviço"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "estoque reservado insuficiente",
			err:         errors.New("estoque reservado insuficiente"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "operação de estoque falhou",
			err:         errors.New("operação de estoque falhou"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid format",
			err:         errors.New("invalid format"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "field is required",
			err:         errors.New("field is required"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "must be at least 3 characters",
			err:         errors.New("must be at least 3 characters"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "cannot delete resource",
			err:         errors.New("cannot delete resource"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapDomainError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
		})
	}
}

func TestMapDomainError_ConflictErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedCode string
		expectedStatus int
	}{
		{
			name:        "não há próximo status disponível",
			err:         errors.New("não há próximo status disponível"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "não é possível modificar ordem fechada",
			err:         errors.New("não é possível modificar ordem fechada"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "não é possível modificar itens após status PENDING_AUTHORIZATION",
			err:         errors.New("não é possível modificar itens após status PENDING_AUTHORIZATION"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "ordem de serviço não está aguardando autorização",
			err:         errors.New("ordem de serviço não está aguardando autorização"),
			expectedCode: ErrCodeValidationFailed,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapDomainError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
		})
	}
}

func TestMapDomainError_NilError(t *testing.T) {
	result := MapDomainError(nil)

	if result.Code != ErrCodeInternalError {
		t.Errorf("Expected code %s, got %s", ErrCodeInternalError, result.Code)
	}

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestMapDomainError_UnknownError(t *testing.T) {
	err := errors.New("some unknown error")
	result := MapDomainError(err)

	if result.Code != ErrCodeInternalError {
		t.Errorf("Expected code %s, got %s", ErrCodeInternalError, result.Code)
	}

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}
