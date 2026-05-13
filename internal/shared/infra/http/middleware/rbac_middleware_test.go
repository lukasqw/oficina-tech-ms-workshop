package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRBACMiddleware_RequireRole_Success(t *testing.T) {
	tests := []struct {
		name         string
		userRole     string
		allowedRoles []string
		shouldPass   bool
	}{
		{
			name:         "USER can access USER endpoint",
			userRole:     RoleUser,
			allowedRoles: []string{RoleUser},
			shouldPass:   true,
		},
		{
			name:         "MANAGER can access USER endpoint (hierarchy)",
			userRole:     RoleManager,
			allowedRoles: []string{RoleUser},
			shouldPass:   true,
		},
		{
			name:         "ADMIN can access USER endpoint (hierarchy)",
			userRole:     RoleAdmin,
			allowedRoles: []string{RoleUser},
			shouldPass:   true,
		},
		{
			name:         "MANAGER can access MANAGER endpoint",
			userRole:     RoleManager,
			allowedRoles: []string{RoleManager},
			shouldPass:   true,
		},
		{
			name:         "ADMIN can access MANAGER endpoint (hierarchy)",
			userRole:     RoleAdmin,
			allowedRoles: []string{RoleManager},
			shouldPass:   true,
		},
		{
			name:         "ADMIN can access ADMIN endpoint",
			userRole:     RoleAdmin,
			allowedRoles: []string{RoleAdmin},
			shouldPass:   true,
		},
		{
			name:         "USER cannot access MANAGER endpoint",
			userRole:     RoleUser,
			allowedRoles: []string{RoleManager},
			shouldPass:   false,
		},
		{
			name:         "USER cannot access ADMIN endpoint",
			userRole:     RoleUser,
			allowedRoles: []string{RoleAdmin},
			shouldPass:   false,
		},
		{
			name:         "MANAGER cannot access ADMIN endpoint",
			userRole:     RoleManager,
			allowedRoles: []string{RoleAdmin},
			shouldPass:   false,
		},
		{
			name:         "Multiple allowed roles - USER matches",
			userRole:     RoleUser,
			allowedRoles: []string{RoleUser, RoleManager},
			shouldPass:   true,
		},
		{
			name:         "Multiple allowed roles - MANAGER matches",
			userRole:     RoleManager,
			allowedRoles: []string{RoleUser, RoleManager},
			shouldPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRBACMiddleware()

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequireRole(tt.allowedRoles...)(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), UserRoleKey, tt.userRole)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.shouldPass {
				if !handlerCalled {
					t.Error("Expected handler to be called")
				}
				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
			} else {
				if handlerCalled {
					t.Error("Expected handler not to be called")
				}
				if w.Code != http.StatusForbidden {
					t.Errorf("Expected status 403, got %d", w.Code)
				}
			}
		})
	}
}

func TestRBACMiddleware_RequireRole_MissingRole(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when role is missing")
	})

	handler := middleware.RequireRole(RoleUser)(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireRole_EmptyRole(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when role is empty")
	})

	handler := middleware.RequireRole(RoleUser)(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, "")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireRole_InvalidRole(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid role")
	})

	handler := middleware.RequireRole(RoleUser)(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, "INVALID_ROLE")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireOwnerOrRole_OwnerAccess(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequireOwnerOrRole(RoleAdmin)(testHandler)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("GET", "/users/"+userID, nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, UserRoleKey, RoleUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for owner access, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireOwnerOrRole_RoleAccess(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequireOwnerOrRole(RoleAdmin)(testHandler)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	resourceID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("GET", "/users/"+resourceID, nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, UserRoleKey, RoleAdmin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin access, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireOwnerOrRole_Forbidden(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware.RequireOwnerOrRole(RoleAdmin)(testHandler)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	resourceID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("GET", "/users/"+resourceID, nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, UserRoleKey, RoleUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireOwnerOrRole_MissingUserID(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware.RequireOwnerOrRole(RoleAdmin)(testHandler)

	req := httptest.NewRequest("GET", "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, RoleUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRBACMiddleware_RequireOwnerOrRole_MissingRole(t *testing.T) {
	middleware := NewRBACMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware.RequireOwnerOrRole(RoleAdmin)(testHandler)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("GET", "/users/"+userID, nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestExtractResourceID(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple path with UUID",
			path:     "/users/550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "path with api prefix",
			path:     "/api/users/550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "path with trailing slash",
			path:     "/users/550e8400-e29b-41d4-a716-446655440000/",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "nested path with UUID",
			path:     "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/profile",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "path without UUID",
			path:     "/users/list",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "",
		},
		{
			name:     "path with invalid UUID",
			path:     "/users/not-a-uuid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractResourceID(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHasRequiredRole(t *testing.T) {
	tests := []struct {
		name         string
		userRole     string
		allowedRoles []string
		expected     bool
	}{
		{
			name:         "exact match USER",
			userRole:     RoleUser,
			allowedRoles: []string{RoleUser},
			expected:     true,
		},
		{
			name:         "exact match MANAGER",
			userRole:     RoleManager,
			allowedRoles: []string{RoleManager},
			expected:     true,
		},
		{
			name:         "exact match ADMIN",
			userRole:     RoleAdmin,
			allowedRoles: []string{RoleAdmin},
			expected:     true,
		},
		{
			name:         "hierarchy MANAGER >= USER",
			userRole:     RoleManager,
			allowedRoles: []string{RoleUser},
			expected:     true,
		},
		{
			name:         "hierarchy ADMIN >= USER",
			userRole:     RoleAdmin,
			allowedRoles: []string{RoleUser},
			expected:     true,
		},
		{
			name:         "hierarchy ADMIN >= MANAGER",
			userRole:     RoleAdmin,
			allowedRoles: []string{RoleManager},
			expected:     true,
		},
		{
			name:         "insufficient USER < MANAGER",
			userRole:     RoleUser,
			allowedRoles: []string{RoleManager},
			expected:     false,
		},
		{
			name:         "insufficient USER < ADMIN",
			userRole:     RoleUser,
			allowedRoles: []string{RoleAdmin},
			expected:     false,
		},
		{
			name:         "insufficient MANAGER < ADMIN",
			userRole:     RoleManager,
			allowedRoles: []string{RoleAdmin},
			expected:     false,
		},
		{
			name:         "invalid user role",
			userRole:     "INVALID",
			allowedRoles: []string{RoleUser},
			expected:     false,
		},
		{
			name:         "invalid allowed role",
			userRole:     RoleUser,
			allowedRoles: []string{"INVALID"},
			expected:     false,
		},
		{
			name:         "multiple allowed roles - matches first",
			userRole:     RoleUser,
			allowedRoles: []string{RoleUser, RoleManager},
			expected:     true,
		},
		{
			name:         "multiple allowed roles - matches second",
			userRole:     RoleManager,
			allowedRoles: []string{RoleUser, RoleManager},
			expected:     true,
		},
		{
			name:         "empty allowed roles",
			userRole:     RoleUser,
			allowedRoles: []string{},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRequiredRole(tt.userRole, tt.allowedRoles)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRoleHierarchy(t *testing.T) {
	// Verify role hierarchy levels are correctly defined
	if roleHierarchy[RoleUser] >= roleHierarchy[RoleManager] {
		t.Error("USER should have lower level than MANAGER")
	}

	if roleHierarchy[RoleManager] >= roleHierarchy[RoleAdmin] {
		t.Error("MANAGER should have lower level than ADMIN")
	}

	if roleHierarchy[RoleUser] >= roleHierarchy[RoleAdmin] {
		t.Error("USER should have lower level than ADMIN")
	}
}
