package validators

import (
	"testing"
)

type TestStruct struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=18,max=100"`
	Role     string `json:"role" validate:"oneof=USER ADMIN MANAGER"`
	ID       string `json:"id" validate:"uuid"`
	Optional string `json:"optional" validate:"omitempty,min=5"`
}

func TestValidateStruct_Required(t *testing.T) {
	tests := []struct {
		name      string
		input     TestStruct
		shouldErr bool
	}{
		{
			name: "all required fields present",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
				Role:  "USER",
				ID:    "550e8400-e29b-41d4-a716-446655440000",
			},
			shouldErr: false,
		},
		{
			name: "missing name",
			input: TestStruct{
				Email: "john@example.com",
				Age:   25,
			},
			shouldErr: true,
		},
		{
			name: "missing email",
			input: TestStruct{
				Name: "John Doe",
				Age:  25,
			},
			shouldErr: true,
		},
		{
			name: "missing age",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_Min(t *testing.T) {
	tests := []struct {
		name      string
		input     TestStruct
		shouldErr bool
	}{
		{
			name: "name meets minimum length",
			input: TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   25,
			},
			shouldErr: false,
		},
		{
			name: "name below minimum length",
			input: TestStruct{
				Name:  "Jo",
				Email: "john@example.com",
				Age:   25,
			},
			shouldErr: true,
		},
		{
			name: "age meets minimum",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   18,
			},
			shouldErr: false,
		},
		{
			name: "age below minimum",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   17,
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_Max(t *testing.T) {
	tests := []struct {
		name      string
		input     TestStruct
		shouldErr bool
	}{
		{
			name: "name within maximum length",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
			},
			shouldErr: false,
		},
		{
			name: "name exceeds maximum length",
			input: TestStruct{
				Name:  "This is a very long name that exceeds the maximum allowed length",
				Email: "john@example.com",
				Age:   25,
			},
			shouldErr: true,
		},
		{
			name: "age within maximum",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   100,
			},
			shouldErr: false,
		},
		{
			name: "age exceeds maximum",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   101,
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_Email(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		shouldErr bool
	}{
		{
			name:      "valid email",
			email:     "john@example.com",
			shouldErr: false,
		},
		{
			name:      "valid email with subdomain",
			email:     "john@mail.example.com",
			shouldErr: false,
		},
		{
			name:      "valid email with plus",
			email:     "john+test@example.com",
			shouldErr: false,
		},
		{
			name:      "invalid email without @",
			email:     "johnexample.com",
			shouldErr: true,
		},
		{
			name:      "invalid email without domain",
			email:     "john@",
			shouldErr: true,
		},
		{
			name:      "invalid email without local part",
			email:     "@example.com",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				Name:  "John Doe",
				Email: tt.email,
				Age:   25,
			}
			err := ValidateStruct(input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_OneOf(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		shouldErr bool
	}{
		{
			name:      "valid role USER",
			role:      "USER",
			shouldErr: false,
		},
		{
			name:      "valid role ADMIN",
			role:      "ADMIN",
			shouldErr: false,
		},
		{
			name:      "valid role MANAGER",
			role:      "MANAGER",
			shouldErr: false,
		},
		{
			name:      "invalid role",
			role:      "INVALID",
			shouldErr: true,
		},
		{
			name:      "lowercase role",
			role:      "user",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
				Role:  tt.role,
			}
			err := ValidateStruct(input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_UUID(t *testing.T) {
	tests := []struct {
		name      string
		uuid      string
		shouldErr bool
	}{
		{
			name:      "valid UUID",
			uuid:      "550e8400-e29b-41d4-a716-446655440000",
			shouldErr: false,
		},
		{
			name:      "invalid UUID format",
			uuid:      "not-a-uuid",
			shouldErr: true,
		},
		{
			name:      "UUID without dashes",
			uuid:      "550e8400e29b41d4a716446655440000",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
				ID:    tt.uuid,
			}
			err := ValidateStruct(input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_OmitEmpty(t *testing.T) {
	tests := []struct {
		name      string
		optional  string
		shouldErr bool
	}{
		{
			name:      "optional field empty",
			optional:  "",
			shouldErr: false,
		},
		{
			name:      "optional field meets minimum",
			optional:  "valid",
			shouldErr: false,
		},
		{
			name:      "optional field below minimum",
			optional:  "test",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      25,
				Optional: tt.optional,
			}
			err := ValidateStruct(input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

type TestStructWithPointer struct {
	Name  *string `json:"name" validate:"required,min=3"`
	Email *string `json:"email" validate:"omitempty,email"`
}

func TestValidateStruct_Pointer(t *testing.T) {
	name := "John"
	email := "john@example.com"
	invalidEmail := "invalid"

	tests := []struct {
		name      string
		input     TestStructWithPointer
		shouldErr bool
	}{
		{
			name: "pointer field with valid value",
			input: TestStructWithPointer{
				Name:  &name,
				Email: &email,
			},
			shouldErr: false,
		},
		{
			name: "pointer field nil (omitempty)",
			input: TestStructWithPointer{
				Name:  &name,
				Email: nil,
			},
			shouldErr: false,
		},
		{
			name: "pointer field with invalid value",
			input: TestStructWithPointer{
				Name:  &name,
				Email: &invalidEmail,
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

type ItemStruct struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

type TestStructWithSlice struct {
	Name  string       `json:"name" validate:"required"`
	Items []ItemStruct `json:"items" validate:"required,dive"`
}

func TestValidateStruct_Dive(t *testing.T) {
	tests := []struct {
		name      string
		input     TestStructWithSlice
		shouldErr bool
	}{
		{
			name: "valid slice with valid items",
			input: TestStructWithSlice{
				Name: "Order",
				Items: []ItemStruct{
					{
						ProductID: "550e8400-e29b-41d4-a716-446655440000",
						Quantity:  5,
					},
					{
						ProductID: "550e8400-e29b-41d4-a716-446655440001",
						Quantity:  3,
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "slice with invalid item UUID",
			input: TestStructWithSlice{
				Name: "Order",
				Items: []ItemStruct{
					{
						ProductID: "invalid-uuid",
						Quantity:  5,
					},
				},
			},
			shouldErr: true,
		},
		{
			name: "slice with invalid item quantity",
			input: TestStructWithSlice{
				Name: "Order",
				Items: []ItemStruct{
					{
						ProductID: "550e8400-e29b-41d4-a716-446655440000",
						Quantity:  0,
					},
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_Gt(t *testing.T) {
	type TestGt struct {
		Quantity int `json:"quantity" validate:"required,gt=0"`
		Price    int `json:"price" validate:"required,gt=100"`
	}

	tests := []struct {
		name      string
		input     TestGt
		shouldErr bool
	}{
		{
			name: "values greater than threshold",
			input: TestGt{
				Quantity: 5,
				Price:    150,
			},
			shouldErr: false,
		},
		{
			name: "quantity equal to threshold",
			input: TestGt{
				Quantity: 0,
				Price:    150,
			},
			shouldErr: true,
		},
		{
			name: "price equal to threshold",
			input: TestGt{
				Quantity: 5,
				Price:    100,
			},
			shouldErr: true,
		},
		{
			name: "values below threshold",
			input: TestGt{
				Quantity: -1,
				Price:    50,
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  uint
		shouldErr bool
	}{
		{
			name:      "valid ID",
			input:     "123",
			expected:  123,
			shouldErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			shouldErr: true,
		},
		{
			name:      "zero ID",
			input:     "0",
			shouldErr: true,
		},
		{
			name:      "negative ID",
			input:     "-1",
			shouldErr: true,
		},
		{
			name:      "non-numeric ID",
			input:     "abc",
			shouldErr: true,
		},
		{
			name:      "whitespace only",
			input:     "   ",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateID(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestValidateQueryParams(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]string
		shouldErr bool
	}{
		{
			name: "valid limit and offset",
			params: map[string]string{
				"limit":  "10",
				"offset": "0",
			},
			shouldErr: false,
		},
		{
			name: "valid limit only",
			params: map[string]string{
				"limit": "50",
			},
			shouldErr: false,
		},
		{
			name: "valid offset only",
			params: map[string]string{
				"offset": "20",
			},
			shouldErr: false,
		},
		{
			name:      "empty params",
			params:    map[string]string{},
			shouldErr: false,
		},
		{
			name: "limit below minimum",
			params: map[string]string{
				"limit": "0",
			},
			shouldErr: true,
		},
		{
			name: "limit above maximum",
			params: map[string]string{
				"limit": "1001",
			},
			shouldErr: true,
		},
		{
			name: "negative offset",
			params: map[string]string{
				"offset": "-1",
			},
			shouldErr: true,
		},
		{
			name: "non-numeric limit",
			params: map[string]string{
				"limit": "abc",
			},
			shouldErr: true,
		},
		{
			name: "non-numeric offset",
			params: map[string]string{
				"offset": "xyz",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueryParams(tt.params)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_NonStruct(t *testing.T) {
	err := ValidateStruct("not a struct")
	if err == nil {
		t.Error("Expected error for non-struct input")
	}
}

func TestValidateStruct_NoValidationTags(t *testing.T) {
	type NoTags struct {
		Name  string
		Email string
	}

	input := NoTags{
		Name:  "",
		Email: "",
	}

	err := ValidateStruct(input)
	if err != nil {
		t.Errorf("Expected no error for struct without validation tags, got: %v", err)
	}
}
