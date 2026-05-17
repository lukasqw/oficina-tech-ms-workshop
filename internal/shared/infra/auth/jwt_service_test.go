package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signToken(t *testing.T, secret string, claims customClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return s
}

func setupJWTService(t *testing.T, secret string) *JWTServiceImpl {
	t.Helper()
	t.Setenv("JWT_SECRET_KEY", secret)
	return NewJWTService()
}

func TestNewJWTService_PanicsWithoutSecret(t *testing.T) {
	if err := os.Unsetenv("JWT_SECRET_KEY"); err != nil {
		t.Fatalf("failed to unset JWT_SECRET_KEY: %v", err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when JWT_SECRET_KEY is not set")
		}
	}()
	NewJWTService()
}

func TestNewJWTService_CreatesServiceWithSecret(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret")
	svc := NewJWTService()
	if svc == nil {
		t.Error("expected non-nil service")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	const secret = "test-secret-key"
	svc := setupJWTService(t, secret)

	claims := customClaims{
		UserID: "123e4567-e89b-12d3-a456-426614174000",
		Name:   "Test User",
		Email:  "test@example.com",
		Role:   "USER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tokenStr := signToken(t, secret, claims)
	result, err := svc.ValidateToken(tokenStr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != claims.UserID {
		t.Errorf("UserID = %s, want %s", result.UserID, claims.UserID)
	}
	if result.Email != claims.Email {
		t.Errorf("Email = %s, want %s", result.Email, claims.Email)
	}
	if result.Role != claims.Role {
		t.Errorf("Role = %s, want %s", result.Role, claims.Role)
	}
	if result.Name != claims.Name {
		t.Errorf("Name = %s, want %s", result.Name, claims.Name)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	const secret = "test-secret-key"
	svc := setupJWTService(t, secret)

	tokenStr := signToken(t, secret, customClaims{
		UserID: "user-1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})

	_, err := svc.ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc := setupJWTService(t, "correct-secret")

	tokenStr := signToken(t, "wrong-secret", customClaims{
		UserID: "user-1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	_, err := svc.ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	svc := setupJWTService(t, "some-secret")

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"plain string", "not.a.jwt"},
		{"only two parts", "header.payload"},
		{"garbage", "###invalid###"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tt.token)
			if err == nil {
				t.Errorf("expected error for token %q", tt.token)
			}
		})
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	svc := setupJWTService(t, "some-secret")

	// Craft a JWT whose header declares RS256 but is otherwise gibberish.
	// The keyfunc rejects any non-HMAC signing method before checking the signature.
	fakeRSAToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9.fakesignature"

	_, err := svc.ValidateToken(fakeRSAToken)
	if err == nil {
		t.Error("expected error for non-HMAC signing method")
	}
}
