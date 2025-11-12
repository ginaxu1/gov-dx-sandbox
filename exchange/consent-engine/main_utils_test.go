package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("WithValue", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "test-value")
		defer os.Unsetenv("TEST_ENV_VAR")
		result := getEnvOrDefault("TEST_ENV_VAR", "default")
		assert.Equal(t, "test-value", result)
	})

	t.Run("WithoutValue", func(t *testing.T) {
		os.Unsetenv("TEST_ENV_VAR")
		result := getEnvOrDefault("TEST_ENV_VAR", "default-value")
		assert.Equal(t, "default-value", result)
	})

	t.Run("EmptyString", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "")
		defer os.Unsetenv("TEST_ENV_VAR")
		result := getEnvOrDefault("TEST_ENV_VAR", "default")
		assert.Equal(t, "default", result)
	})
}

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a < b", 1, 2, 1},
		{"a > b", 5, 3, 3},
		{"a == b", 4, 4, 4},
		{"negative values", -5, -3, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetOwnerEmailByID(t *testing.T) {
	t.Run("ValidOwnerID", func(t *testing.T) {
		email, err := getOwnerEmailByID("user@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "user@example.com", email)
	})

	t.Run("EmptyOwnerID", func(t *testing.T) {
		email, err := getOwnerEmailByID("")
		assert.Error(t, err)
		assert.Equal(t, "", email)
		assert.Contains(t, err.Error(), "owner_id cannot be empty")
	})
}
