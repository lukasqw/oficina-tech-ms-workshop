package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondSuccess(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
	}{
		{
			name:   "success with map data",
			status: http.StatusOK,
			data:   map[string]string{"id": "123", "name": "Test"},
		},
		{
			name:   "success with struct data",
			status: http.StatusCreated,
			data:   struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{ID: "456", Name: "Test User"},
		},
		{
			name:   "success with nil data",
			status: http.StatusNoContent,
			data:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondSuccess(w, tt.status, tt.data)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var envelope Envelope
			err := json.Unmarshal(w.Body.Bytes(), &envelope)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if len(envelope.Errors) != 0 {
				t.Errorf("Expected empty errors array, got %d errors", len(envelope.Errors))
			}
		})
	}
}

func TestRespondErrorEnvelope(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		code           string
		message        string
	}{
		{
			name:    "not found error",
			status:  http.StatusNotFound,
			code:    ErrCodeNotFound,
			message: "Resource not found",
		},
		{
			name:    "validation error",
			status:  http.StatusBadRequest,
			code:    ErrCodeValidationFailed,
			message: "Validation failed",
		},
		{
			name:    "internal error",
			status:  http.StatusInternalServerError,
			code:    ErrCodeInternalError,
			message: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondErrorEnvelope(w, tt.status, tt.code, tt.message)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var envelope Envelope
			err := json.Unmarshal(w.Body.Bytes(), &envelope)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if envelope.Data != nil {
				t.Error("Expected data to be nil")
			}

			if len(envelope.Errors) != 1 {
				t.Fatalf("Expected 1 error, got %d", len(envelope.Errors))
			}

			if envelope.Errors[0].Code != tt.code {
				t.Errorf("Expected code %s, got %s", tt.code, envelope.Errors[0].Code)
			}

			if envelope.Errors[0].Message != tt.message {
				t.Errorf("Expected message %s, got %s", tt.message, envelope.Errors[0].Message)
			}

			if envelope.Errors[0].Field != nil {
				t.Error("Expected field to be nil")
			}
		})
	}
}

func TestRespondValidationError(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		message string
	}{
		{
			name:    "email validation error",
			field:   "email",
			message: "Email is required",
		},
		{
			name:    "name validation error",
			field:   "name",
			message: "Name must be at least 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondValidationError(w, tt.field, tt.message)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var envelope Envelope
			err := json.Unmarshal(w.Body.Bytes(), &envelope)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if envelope.Data != nil {
				t.Error("Expected data to be nil")
			}

			if len(envelope.Errors) != 1 {
				t.Fatalf("Expected 1 error, got %d", len(envelope.Errors))
			}

			if envelope.Errors[0].Code != ErrCodeValidationFailed {
				t.Errorf("Expected code %s, got %s", ErrCodeValidationFailed, envelope.Errors[0].Code)
			}

			if envelope.Errors[0].Message != tt.message {
				t.Errorf("Expected message %s, got %s", tt.message, envelope.Errors[0].Message)
			}

			if envelope.Errors[0].Field == nil {
				t.Fatal("Expected field to be present")
			}

			if *envelope.Errors[0].Field != tt.field {
				t.Errorf("Expected field %s, got %s", tt.field, *envelope.Errors[0].Field)
			}
		})
	}
}

func TestRespondMultipleErrors(t *testing.T) {
	field1 := "email"
	field2 := "name"
	errors := []ErrorDetail{
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
	}

	w := httptest.NewRecorder()
	RespondMultipleErrors(w, http.StatusBadRequest, errors)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var envelope Envelope
	err := json.Unmarshal(w.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if envelope.Data != nil {
		t.Error("Expected data to be nil")
	}

	if len(envelope.Errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(envelope.Errors))
	}

	if envelope.Errors[0].Field == nil || *envelope.Errors[0].Field != "email" {
		t.Error("Expected first error field to be 'email'")
	}

	if envelope.Errors[1].Field == nil || *envelope.Errors[1].Field != "name" {
		t.Error("Expected second error field to be 'name'")
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-empty string",
			input: "test",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "string with spaces",
			input: "  test  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPtr(tt.input)

			if result == nil {
				t.Fatal("Expected non-nil pointer")
			}

			if *result != tt.input {
				t.Errorf("Expected %s, got %s", tt.input, *result)
			}
		})
	}
}
