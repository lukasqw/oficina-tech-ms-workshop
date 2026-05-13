package middleware

import (
	"net/http"
	"net/http/httptest"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/auth"
	"testing"
)

// mockJWTService implements auth.JWTService for testing
type mockJWTService struct {
	validateTokenFunc func(token string) (*auth.TokenClaims, error)
}

func (m *mockJWTService) GenerateToken(claims auth.TokenClaims) (string, error) {
	return "", nil
}

func (m *mockJWTService) ValidateToken(token string) (*auth.TokenClaims, error) {
	return m.validateTokenFunc(token)
}

// InvalidTokenError is a mock error for testing
type InvalidTokenError struct{}

func (e *InvalidTokenError) Error() string {
	return "invalid token"
}

func TestAuthMiddleware_Authenticate_Success(t *testing.T) {
	// Mock JWT service
	mockService := &mockJWTService{
		validateTokenFunc: func(token string) (*auth.TokenClaims, error) {
			return &auth.TokenClaims{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
				Email:  "test@example.com",
				Role:   "USER",
			}, nil
		},
	}

	middleware := NewAuthMiddleware(mockService)

	// Create test handler
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context values
		userID := r.Context().Value(UserIDKey)
		userEmail := r.Context().Value(UserEmailKey)
		userRole := r.Context().Value(UserRoleKey)

		if userID != "123e4567-e89b-12d3-a456-426614174000" {
			t.Errorf("Expected user ID '123e4567-e89b-12d3-a456-426614174000', got '%v'", userID)
		}

		if userEmail != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%v'", userEmail)
		}

		if userRole != "USER" {
			t.Errorf("Expected role 'USER', got '%v'", userRole)
		}

		w.WriteHeader(http.StatusOK)
	}))

	// Create request with Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_MissingAuthorizationHeader(t *testing.T) {
	mockService := &mockJWTService{
		validateTokenFunc: func(token string) (*auth.TokenClaims, error) {
			return nil, &InvalidTokenError{}
		},
	}
	middleware := NewAuthMiddleware(mockService)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_InvalidAuthorizationFormat(t *testing.T) {
	mockService := &mockJWTService{
		validateTokenFunc: func(token string) (*auth.TokenClaims, error) {
			return &auth.TokenClaims{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
				Email:  "test@example.com",
				Role:   "USER",
			}, nil
		},
	}
	middleware := NewAuthMiddleware(mockService)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	tests := []struct {
		name   string
		header string
	}{
		{"Missing Bearer prefix", "token-without-bearer"},
		{"Wrong prefix", "Basic token"},
		{"Empty token", "Bearer "},
		{"Only Bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", rec.Code)
			}
		})
	}
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	mockService := &mockJWTService{
		validateTokenFunc: func(token string) (*auth.TokenClaims, error) {
			return nil, &InvalidTokenError{}
		},
	}

	middleware := NewAuthMiddleware(mockService)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}
