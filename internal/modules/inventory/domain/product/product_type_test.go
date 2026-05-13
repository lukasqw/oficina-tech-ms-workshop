package product

import (
	"encoding/json"
	"testing"
)

func TestNewProductType(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		wantErr     bool
		expectedVal string
	}{
		{
			name:        "valid CONSUMABLE type",
			value:       ProductTypeConsumable,
			wantErr:     false,
			expectedVal: ProductTypeConsumable,
		},
		{
			name:        "valid SIMPLE type",
			value:       ProductTypeSimple,
			wantErr:     false,
			expectedVal: ProductTypeSimple,
		},
		{
			name:        "valid PARTS type",
			value:       ProductTypeParts,
			wantErr:     false,
			expectedVal: ProductTypeParts,
		},
		{
			name:    "invalid type - lowercase",
			value:   "consumable",
			wantErr: true,
		},
		{
			name:    "invalid type - empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid type - random value",
			value:   "INVALID_TYPE",
			wantErr: true,
		},
		{
			name:    "invalid type - numeric",
			value:   "123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProductType(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProductType() error = nil, wantErr %v", tt.wantErr)
				}
				if err != ErrInvalidProductType {
					t.Errorf("NewProductType() error = %v, want %v", err, ErrInvalidProductType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProductType() unexpected error = %v", err)
				return
			}

			if got.Value() != tt.expectedVal {
				t.Errorf("NewProductType().Value() = %v, want %v", got.Value(), tt.expectedVal)
			}
		})
	}
}

func TestProductType_String(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "CONSUMABLE type string representation",
			value:    ProductTypeConsumable,
			expected: ProductTypeConsumable,
		},
		{
			name:     "SIMPLE type string representation",
			value:    ProductTypeSimple,
			expected: ProductTypeSimple,
		},
		{
			name:     "PARTS type string representation",
			value:    ProductTypeParts,
			expected: ProductTypeParts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := NewProductType(tt.value)
			if err != nil {
				t.Fatalf("NewProductType() unexpected error = %v", err)
			}

			if got := pt.String(); got != tt.expected {
				t.Errorf("ProductType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProductType_Value(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "CONSUMABLE type value",
			value:    ProductTypeConsumable,
			expected: ProductTypeConsumable,
		},
		{
			name:     "SIMPLE type value",
			value:    ProductTypeSimple,
			expected: ProductTypeSimple,
		},
		{
			name:     "PARTS type value",
			value:    ProductTypeParts,
			expected: ProductTypeParts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := NewProductType(tt.value)
			if err != nil {
				t.Fatalf("NewProductType() unexpected error = %v", err)
			}

			if got := pt.Value(); got != tt.expected {
				t.Errorf("ProductType.Value() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProductType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "marshal CONSUMABLE type",
			value:    ProductTypeConsumable,
			expected: `"CONSUMABLE"`,
		},
		{
			name:     "marshal SIMPLE type",
			value:    ProductTypeSimple,
			expected: `"SIMPLE"`,
		},
		{
			name:     "marshal PARTS type",
			value:    ProductTypeParts,
			expected: `"PARTS"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := NewProductType(tt.value)
			if err != nil {
				t.Fatalf("NewProductType() unexpected error = %v", err)
			}

			got, err := pt.MarshalJSON()
			if err != nil {
				t.Errorf("ProductType.MarshalJSON() error = %v", err)
				return
			}

			if string(got) != tt.expected {
				t.Errorf("ProductType.MarshalJSON() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestProductType_JSONIntegration(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "JSON marshal/unmarshal CONSUMABLE",
			value:    ProductTypeConsumable,
			expected: `"CONSUMABLE"`,
		},
		{
			name:     "JSON marshal/unmarshal SIMPLE",
			value:    ProductTypeSimple,
			expected: `"SIMPLE"`,
		},
		{
			name:     "JSON marshal/unmarshal PARTS",
			value:    ProductTypeParts,
			expected: `"PARTS"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := NewProductType(tt.value)
			if err != nil {
				t.Fatalf("NewProductType() unexpected error = %v", err)
			}

			// Test marshaling
			jsonData, err := json.Marshal(pt)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			if string(jsonData) != tt.expected {
				t.Errorf("json.Marshal() = %v, want %v", string(jsonData), tt.expected)
			}
		})
	}
}
