package main

import (
	"testing"
)


func TestGetTypeOfChange(t *testing.T) {
	tests := []struct {
		name     string
		commits  []string
		expected string
	}{
		{
			name:     "Empty commits",
			commits:  []string{},
			expected: patch,
		},
		{
			name:     "Feature commit",
			commits:  []string{"feat(something): add login feature"},
			expected: minor,
		},
		{
			name:     "Multiple commits, last is feat",
			commits:  []string{"chore(scope): update dependencies", "feat: add new API endpoint"},
			expected: minor,
		},
		{
			name:     "Multiple commits, last is breaking",
			commits:  []string{"feat: add something", "fix!: critical fix"},
			expected: major,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeOfChange(tt.commits)
			if result != tt.expected {
				t.Errorf("getTypeOfChange(%v) = %v, want %v", tt.commits, result, tt.expected)
			}
		})
	}
}
