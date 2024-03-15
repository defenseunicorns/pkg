package main

import (
	"testing"
)

func TestGetTypeOfChange(t *testing.T) {
	tests := []struct {
		name     string
		commit  string
		expected string
	}{
		{
			name:     "Feature commit",
			commit:  "feat(something): add login feature",
			expected: minor,
		},
		{
			name:     "Breaking change commit, then feature",
			commit:  "fix!: fix critical bug",
			expected: major,
		},
		{
			name:     "Multiple commits, last is feat",
			commit:  "chore(scope): update dependencies",
			expected: patch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeOfChange(tt.commit)
			if result != tt.expected {
				t.Errorf("getTypeOfChange(%v) = %v, want %v", tt.commit, result, tt.expected)
			}
		})
	}
}
