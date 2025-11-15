package crdp

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("192.168.0.231", 32082, "P03", 10)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.policy != "P03" {
		t.Errorf("expected policy P03, got %s", client.policy)
	}
}

func TestIsSuccess(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, true},
		{201, true},
		{299, true},
		{400, false},
		{500, false},
		{0, false},
	}

	for _, tt := range tests {
		resp := &APIResponse{StatusCode: tt.statusCode}
		if resp.IsSuccess() != tt.expected {
			t.Errorf("IsSuccess(%d) = %v, want %v", tt.statusCode, resp.IsSuccess(), tt.expected)
		}
	}
}

func TestExtractProtectedFromProtectResponse(t *testing.T) {
	client := NewClient("192.168.0.231", 32082, "P03", 10)

	tests := []struct {
		name     string
		body     interface{}
		expected string
	}{
		{
			name: "protected_data key",
			body: map[string]interface{}{"protected_data": "token123"},
			expected: "token123",
		},
		{
			name: "protected key",
			body: map[string]interface{}{"protected": "token456"},
			expected: "token456",
		},
		{
			name: "token key",
			body: map[string]interface{}{"token": "token789"},
			expected: "token789",
		},
		{
			name: "nil body",
			body: nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		resp := &APIResponse{Body: tt.body}
		result := client.ExtractProtectedFromProtectResponse(resp)
		if result != tt.expected {
			t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, result)
		}
	}
}

func TestExtractProtectedListFromProtectResponse(t *testing.T) {
	client := NewClient("192.168.0.231", 32082, "P03", 10)

	tests := []struct {
		name     string
		body     interface{}
		expected []string
	}{
		{
			name: "direct list",
			body: []interface{}{"token1", "token2", "token3"},
			expected: []string{"token1", "token2", "token3"},
		},
		{
			name: "protected_data dict",
			body: map[string]interface{}{
				"protected_data": []interface{}{"token1", "token2"},
			},
			expected: []string{"token1", "token2"},
		},
		{
			name: "protected_data_array",
			body: map[string]interface{}{
				"protected_data_array": []interface{}{
					map[string]interface{}{"protected_data": "token1"},
					map[string]interface{}{"protected_data": "token2"},
				},
			},
			expected: []string{"token1", "token2"},
		},
		{
			name: "nil body",
			body: nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		resp := &APIResponse{Body: tt.body}
		result := client.ExtractProtectedListFromProtectResponse(resp)
		if len(result) != len(tt.expected) {
			t.Errorf("%s: expected length %d, got %d", tt.name, len(tt.expected), len(result))
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("%s: expected %s at index %d, got %s", tt.name, tt.expected[i], i, v)
			}
		}
	}
}
