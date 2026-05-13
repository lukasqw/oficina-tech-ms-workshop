package utils

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateUUIDv7(t *testing.T) {
	// Generate multiple UUIDs to ensure uniqueness
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateUUIDv7()

		// Check if UUID is valid format
		_, err := uuid.Parse(id)
		if err != nil {
			t.Errorf("Generated invalid UUID: %s, error: %v", id, err)
		}

		// Check for uniqueness
		if uuids[id] {
			t.Errorf("Generated duplicate UUID: %s", id)
		}
		uuids[id] = true

		// Check format (8-4-4-4-12)
		parts := strings.Split(id, "-")
		if len(parts) != 5 {
			t.Errorf("UUID has wrong format: %s", id)
		}
	}
}

func TestValidateUUID_Valid(t *testing.T) {
	tests := []struct {
		name string
		uuid string
	}{
		{
			name: "valid UUID v4",
			uuid: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "valid UUID v7",
			uuid: GenerateUUIDv7(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if err != nil {
				t.Errorf("Expected valid UUID, got error: %v", err)
			}
		})
	}
}

func TestValidateUUID_Invalid(t *testing.T) {
	tests := []struct {
		name string
		uuid string
	}{
		{
			name: "empty string",
			uuid: "",
		},
		{
			name: "invalid format",
			uuid: "not-a-uuid",
		},
		{
			name: "wrong length",
			uuid: "550e8400-e29b-41d4-a716",
		},
		{
			name: "invalid characters",
			uuid: "550e8400-e29b-41d4-a716-44665544000g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if err == nil {
				t.Error("Expected error for invalid UUID, got none")
			}
		})
	}
}

func TestParseUUID_Valid(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	
	result, err := ParseUUID(validUUID)
	if err != nil {
		t.Errorf("Expected valid UUID, got error: %v", err)
	}

	if result.String() != validUUID {
		t.Errorf("Expected %s, got %s", validUUID, result.String())
	}
}

func TestParseUUID_Invalid(t *testing.T) {
	tests := []struct {
		name string
		uuid string
	}{
		{
			name: "empty string",
			uuid: "",
		},
		{
			name: "invalid format",
			uuid: "not-a-uuid",
		},
		{
			name: "invalid characters",
			uuid: "550e8400-e29b-41d4-a716-44665544000g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseUUID(tt.uuid)
			if err == nil {
				t.Error("Expected error for invalid UUID, got none")
			}
		})
	}
}
