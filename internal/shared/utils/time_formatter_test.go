package utils

import (
	"testing"
	"time"
)

func TestFormatTimeRFC3339(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "specific date and time",
			input:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			expected: "2024-01-15T10:30:45Z",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimeRFC3339(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseTimeRFC3339(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "valid RFC3339 format",
			input:     "2024-01-15T10:30:45Z",
			shouldErr: false,
		},
		{
			name:      "valid RFC3339 with timezone",
			input:     "2024-01-15T10:30:45-03:00",
			shouldErr: false,
		},
		{
			name:      "invalid format",
			input:     "2024-01-15 10:30:45",
			shouldErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeRFC3339(tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsZero() {
					t.Error("Expected non-zero time")
				}
			}
		})
	}
}

func TestFormatAndParse_RoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	
	formatted := FormatTimeRFC3339(original)
	parsed, err := ParseTimeRFC3339(formatted)
	
	if err != nil {
		t.Fatalf("Failed to parse formatted time: %v", err)
	}
	
	if !parsed.Equal(original) {
		t.Errorf("Round trip failed. Original: %v, Parsed: %v", original, parsed)
	}
}
