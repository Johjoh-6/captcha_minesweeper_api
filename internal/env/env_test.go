package env

import (
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "returns env value when exists",
			key:          "TEST_STRING_KEY",
			defaultValue: "default",
			envValue:     "environment_value",
			setEnv:       true,
			expected:     "environment_value",
		},
		{
			name:         "returns default when env not set",
			key:          "MISSING_STRING_KEY",
			defaultValue: "default",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "returns default when env is empty",
			key:          "EMPTY_STRING_KEY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "handles empty default value",
			key:          "TEST_EMPTY_DEFAULT",
			defaultValue: "",
			envValue:     "test_value",
			setEnv:       true,
			expected:     "test_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := GetString(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetString() = %v, want %v", result, tt.expected)
			}

			// Clean up after test
			os.Unsetenv(tt.key)
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		expected     int
		shouldPanic  bool
	}{
		{
			name:         "returns env value when valid int",
			key:          "TEST_INT_KEY",
			defaultValue: 42,
			envValue:     "123",
			setEnv:       true,
			expected:     123,
		},
		{
			name:         "returns default when env not set",
			key:          "MISSING_INT_KEY",
			defaultValue: 42,
			setEnv:       false,
			expected:     42,
		},
		{
			name:         "returns default when env is empty",
			key:          "EMPTY_INT_KEY",
			defaultValue: 42,
			envValue:     "",
			setEnv:       true,
			expected:     42,
		},
		{
			name:         "handles zero value",
			key:          "ZERO_INT_KEY",
			defaultValue: 42,
			envValue:     "0",
			setEnv:       true,
			expected:     0,
		},
		{
			name:         "handles negative value",
			key:          "NEG_INT_KEY",
			defaultValue: 42,
			envValue:     "-100",
			setEnv:       true,
			expected:     -100,
		},
		{
			name:         "panics on invalid int",
			key:          "INVALID_INT_KEY",
			defaultValue: 42,
			envValue:     "not_a_number",
			setEnv:       true,
			shouldPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("GetInt() should have panicked")
					}
				}()
			}

			result := GetInt(tt.key, tt.defaultValue)
			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("GetInt() = %v, want %v", result, tt.expected)
			}

			// Clean up after test
			os.Unsetenv(tt.key)
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
		shouldPanic  bool
	}{
		{
			name:         "returns true for 'true'",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false for 'false'",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns true for '1'",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false for '0'",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns default when env not set",
			key:          "MISSING_BOOL_KEY",
			defaultValue: true,
			setEnv:       false,
			expected:     true,
		},
		{
			name:         "returns default when env is empty",
			key:          "EMPTY_BOOL_KEY",
			defaultValue: true,
			envValue:     "",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "handles case insensitive true",
			key:          "TEST_BOOL_CASE",
			defaultValue: false,
			envValue:     "TRUE",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "panics on invalid bool",
			key:          "INVALID_BOOL_KEY",
			defaultValue: false,
			envValue:     "maybe",
			setEnv:       true,
			shouldPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("GetBool() should have panicked")
					}
				}()
			}

			result := GetBool(tt.key, tt.defaultValue)
			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("GetBool() = %v, want %v", result, tt.expected)
			}

			// Clean up after test
			os.Unsetenv(tt.key)
		})
	}
}
