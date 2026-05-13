package database

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockEnvReader é um mock para testes
type MockEnvReader struct {
	values map[string]string
}

func NewMockEnvReader(values map[string]string) *MockEnvReader {
	return &MockEnvReader{values: values}
}

func (m *MockEnvReader) Getenv(key string) string {
	return m.values[key]
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "DSN with all environment variables",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_PORT":     "5432",
				"DB_SSLMODE":  "require",
			},
			expected: "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=require",
		},
		{
			name: "DSN with default sslmode when not provided",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_PORT":     "5432",
			},
			expected: "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable",
		},
		{
			name: "DSN with production settings",
			envVars: map[string]string{
				"DB_HOST":     "prod-db.example.com",
				"DB_USER":     "produser",
				"DB_PASSWORD": "prodpass123",
				"DB_NAME":     "oficina_prod",
				"DB_PORT":     "5433",
				"DB_SSLMODE":  "verify-full",
			},
			expected: "host=prod-db.example.com user=produser password=prodpass123 dbname=oficina_prod port=5433 sslmode=verify-full",
		},
		{
			name: "DSN with empty password",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "",
				"DB_NAME":     "testdb",
				"DB_PORT":     "5432",
				"DB_SSLMODE":  "disable",
			},
			expected: "host=localhost user=postgres password= dbname=testdb port=5432 sslmode=disable",
		},
		{
			name: "DSN with special characters in password",
			envVars: map[string]string{
				"DB_HOST":     "db.example.com",
				"DB_USER":     "admin",
				"DB_PASSWORD": "p@ss!w0rd#123",
				"DB_NAME":     "production",
				"DB_PORT":     "5433",
				"DB_SSLMODE":  "require",
			},
			expected: "host=db.example.com user=admin password=p@ss!w0rd#123 dbname=production port=5433 sslmode=require",
		},
		{
			name: "DSN with empty sslmode defaults to disable",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "pass",
				"DB_NAME":     "db",
				"DB_PORT":     "5432",
				"DB_SSLMODE":  "",
			},
			expected: "host=localhost user=user password=pass dbname=db port=5432 sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnv := NewMockEnvReader(tt.envVars)
			result := BuildDSN(mockEnv)

			if result != tt.expected {
				t.Errorf("Expected DSN:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestBuildDSN_WithRealEnv(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		shouldContain []string
	}{
		{
			name: "Real environment variables",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_PORT":     "5432",
				"DB_SSLMODE":  "require",
			},
			shouldContain: []string{
				"host=localhost",
				"user=testuser",
				"password=testpass",
				"dbname=testdb",
				"port=5432",
				"sslmode=require",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Set environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", key, err)
				}
			}

			// Cleanup: Restore environment after test
			defer func() {
				for key := range tt.envVars {
					if err := os.Unsetenv(key); err != nil {
						t.Errorf("Failed to unset environment variable %s: %v", key, err)
					}
				}
			}()

			// Execute with real OSEnvReader
			envReader := &OSEnvReader{}
			dsn := BuildDSN(envReader)

			// Assert: Check that DSN contains expected parts
			for _, expected := range tt.shouldContain {
				if !strings.Contains(dsn, expected) {
					t.Errorf("DSN should contain '%s', but got: %s", expected, dsn)
				}
			}
		})
	}
}

func TestOSEnvReader_Getenv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "Read existing environment variable",
			key:      "TEST_VAR_1",
			value:    "test_value_1",
			expected: "test_value_1",
		},
		{
			name:     "Read non-existing environment variable",
			key:      "NON_EXISTING_VAR",
			value:    "",
			expected: "",
		},
		{
			name:     "Read environment variable with special characters",
			key:      "TEST_VAR_2",
			value:    "value@with#special$chars",
			expected: "value@with#special$chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatalf("Failed to set environment variable: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Errorf("Failed to unset environment variable: %v", err)
					}
				}()
			}

			reader := &OSEnvReader{}
			result := reader.Getenv(tt.key)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetModelsToMigrate(t *testing.T) {
	models := GetModelsToMigrate()

	// MS3 owns service_catalog + inventory (product, inventory, saga_operation).
	expectedCount := 4
	if len(models) != expectedCount {
		t.Errorf("Expected %d models, got %d", expectedCount, len(models))
	}

	// Verifica se nenhum modelo é nil
	for i, model := range models {
		if model == nil {
			t.Errorf("Model at index %d is nil", i)
		}
	}
}

func TestBuildDSN_SSLModeVariations(t *testing.T) {
	tests := []struct {
		name            string
		sslmode         string
		expectedSSLMode string
	}{
		{
			name:            "sslmode disable",
			sslmode:         "disable",
			expectedSSLMode: "sslmode=disable",
		},
		{
			name:            "sslmode require",
			sslmode:         "require",
			expectedSSLMode: "sslmode=require",
		},
		{
			name:            "sslmode verify-ca",
			sslmode:         "verify-ca",
			expectedSSLMode: "sslmode=verify-ca",
		},
		{
			name:            "sslmode verify-full",
			sslmode:         "verify-full",
			expectedSSLMode: "sslmode=verify-full",
		},
		{
			name:            "sslmode prefer",
			sslmode:         "prefer",
			expectedSSLMode: "sslmode=prefer",
		},
		{
			name:            "empty sslmode defaults to disable",
			sslmode:         "",
			expectedSSLMode: "sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnv := NewMockEnvReader(map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "pass",
				"DB_NAME":     "db",
				"DB_PORT":     "5432",
				"DB_SSLMODE":  tt.sslmode,
			})

			dsn := BuildDSN(mockEnv)

			if !strings.Contains(dsn, tt.expectedSSLMode) {
				t.Errorf("Expected DSN to contain '%s', got: %s", tt.expectedSSLMode, dsn)
			}
		})
	}
}

func TestBuildDSN_AllParameters(t *testing.T) {
	mockEnv := NewMockEnvReader(map[string]string{
		"DB_HOST":     "testhost",
		"DB_USER":     "testuser",
		"DB_PASSWORD": "testpass",
		"DB_NAME":     "testdb",
		"DB_PORT":     "5432",
		"DB_SSLMODE":  "require",
	})

	dsn := BuildDSN(mockEnv)

	// Verifica cada parâmetro individualmente
	requiredParams := []string{
		"host=testhost",
		"user=testuser",
		"password=testpass",
		"dbname=testdb",
		"port=5432",
		"sslmode=require",
	}

	for _, param := range requiredParams {
		if !strings.Contains(dsn, param) {
			t.Errorf("DSN missing parameter: %s\nFull DSN: %s", param, dsn)
		}
	}
}

func TestMockEnvReader(t *testing.T) {
	values := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "",
	}

	mock := NewMockEnvReader(values)

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "Get existing key",
			key:      "KEY1",
			expected: "value1",
		},
		{
			name:     "Get another existing key",
			key:      "KEY2",
			expected: "value2",
		},
		{
			name:     "Get key with empty value",
			key:      "KEY3",
			expected: "",
		},
		{
			name:     "Get non-existing key",
			key:      "NON_EXISTING",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mock.Getenv(tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// MockDBOpener é um mock para testes de conexão
type MockDBOpener struct {
	shouldFail bool
	db         *gorm.DB
}

func (m *MockDBOpener) Open(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock connection error")
	}

	if m.db != nil {
		return m.db, nil
	}

	// Retorna erro se não houver DB configurado
	return nil, fmt.Errorf("no mock db configured")
}

func TestConnectWithDependencies_ConnectionError(t *testing.T) {
	mockEnv := NewMockEnvReader(map[string]string{
		"DB_HOST":     "localhost",
		"DB_USER":     "testuser",
		"DB_PASSWORD": "testpass",
		"DB_NAME":     "testdb",
		"DB_PORT":     "5432",
		"DB_SSLMODE":  "disable",
	})

	mockOpener := &MockDBOpener{
		shouldFail: true,
	}

	db, err := ConnectWithDependencies(mockEnv, mockOpener)

	if err == nil {
		t.Error("Expected error when connection fails")
	}

	if db != nil {
		t.Error("Expected db to be nil when connection fails")
	}

	if !strings.Contains(err.Error(), "failed to connect to database") {
		t.Errorf("Expected connection error message, got: %v", err)
	}
}

func TestConnectWithDependencies_DSNConstruction(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "Standard configuration",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "pass",
				"DB_NAME":     "db",
				"DB_PORT":     "5432",
				"DB_SSLMODE":  "disable",
			},
		},
		{
			name: "Production configuration",
			envVars: map[string]string{
				"DB_HOST":     "prod.example.com",
				"DB_USER":     "produser",
				"DB_PASSWORD": "prodpass",
				"DB_NAME":     "proddb",
				"DB_PORT":     "5433",
				"DB_SSLMODE":  "require",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnv := NewMockEnvReader(tt.envVars)
			mockOpener := &MockDBOpener{shouldFail: true}

			// Tenta conectar (vai falhar, mas testa a construção da DSN)
			_, err := ConnectWithDependencies(mockEnv, mockOpener)

			if err == nil {
				t.Error("Expected error from mock opener")
			}

			// Verifica se a DSN foi construída corretamente
			dsn := BuildDSN(mockEnv)
			for key, value := range tt.envVars {
				if key == "DB_SSLMODE" && value == "" {
					continue
				}
				// Verifica se os valores estão na DSN (exceto a chave)
				if value != "" && !strings.Contains(dsn, value) {
					t.Errorf("DSN should contain value '%s' from %s", value, key)
				}
			}
		})
	}
}

func TestGormDBOpener_Open(t *testing.T) {
	opener := &GormDBOpener{}

	// Testa com uma configuração inválida (deve falhar ou retornar db)
	db, err := opener.Open(postgres.Open("invalid dsn"), &gorm.Config{})

	// GORM pode não retornar erro imediatamente, mas o DB não deve ser utilizável
	// O importante é que o método Open foi chamado corretamente
	if err == nil && db == nil {
		t.Error("Expected either error or db to be returned")
	}
}
