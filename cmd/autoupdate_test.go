package cmd

import (
	"reflect"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"v1.2.3", []int{1, 2, 3}},
		{"1.2.3", []int{1, 2, 3}},
		{"v0.1.0-draft", []int{0, 1, 0}},
		{"v2.0", []int{2, 0}},
		{"dev", []int{0}},
		{"", []int{}},
	}

	for _, test := range tests {
		result := parseVersion(test.input)
		if len(result) == 0 && len(test.expected) == 0 {
			continue
		}
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("parseVersion(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"v1.0.0", "v1.1.0", true},
		{"v1.0.0", "v2.0.0", true},
		{"v1.0.1", "v1.0.0", false},
		{"v2.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.0.0", false},
		{"dev", "v1.0.0", false},
		{"v1.0.0", "", false},
		{"", "v1.0.0", false},
		{"v0.1.0", "v0.1.1-draft", true},
		{"v0.1.1-draft", "v0.1.0", false},
	}

	for _, test := range tests {
		result := isNewerVersion(test.current, test.latest)
		if result != test.expected {
			t.Errorf("isNewerVersion(%q, %q) = %v; want %v", test.current, test.latest, result, test.expected)
		}
	}
}
