package utils

import (
	"encoding/json"
	"testing"
)

func TestEnvelope_SuccessResponse(t *testing.T) {
	data := map[string]string{"id": "123", "name": "Test"}
	envelope := Envelope{
		Data:   data,
		Errors: []ErrorDetail{},
	}

	jsonData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	var result Envelope
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if result.Data == nil {
		t.Error("Expected data to be present")
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected empty errors array, got %d errors", len(result.Errors))
	}
}

func TestEnvelope_ErrorResponse(t *testing.T) {
	field := "email"
	envelope := Envelope{
		Data: nil,
		Errors: []ErrorDetail{
			{
				Code:    ErrCodeValidationFailed,
				Message: "O campo email é obrigatório",
				Field:   &field,
			},
		},
	}

	jsonData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	var result Envelope
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if result.Data != nil {
		t.Error("Expected data to be nil")
	}

	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Code != ErrCodeValidationFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeValidationFailed, result.Errors[0].Code)
	}

	if result.Errors[0].Field == nil {
		t.Error("Expected field to be present")
	} else if *result.Errors[0].Field != "email" {
		t.Errorf("Expected field 'email', got '%s'", *result.Errors[0].Field)
	}
}

func TestErrorDetail_WithoutField(t *testing.T) {
	errorDetail := ErrorDetail{
		Code:    ErrCodeNotFound,
		Message: "Resource not found",
		Field:   nil,
	}

	jsonData, err := json.Marshal(errorDetail)
	if err != nil {
		t.Fatalf("Failed to marshal error detail: %v", err)
	}

	var result ErrorDetail
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal error detail: %v", err)
	}

	if result.Code != ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeNotFound, result.Code)
	}

	if result.Field != nil {
		t.Error("Expected field to be nil")
	}
}

func TestEnvelope_MultipleErrors(t *testing.T) {
	field1 := "email"
	field2 := "name"
	envelope := Envelope{
		Data: nil,
		Errors: []ErrorDetail{
			{
				Code:    ErrCodeValidationFailed,
				Message: "Email is required",
				Field:   &field1,
			},
			{
				Code:    ErrCodeValidationFailed,
				Message: "Name is required",
				Field:   &field2,
			},
		},
	}

	if len(envelope.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(envelope.Errors))
	}

	jsonData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	var result Envelope
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors after unmarshal, got %d", len(result.Errors))
	}
}
