package toolchain

import "testing"

func TestParseRequirementUsesCapturedBytes(t *testing.T) {
	for _, test := range []struct {
		name      string
		source    string
		language  string
		toolchain string
		fails     bool
	}{
		{"language fallback", "module example.org/one\n\ngo 1.27.3\n", "go1.27.3", "go1.27.3", false},
		{"explicit toolchain", "module example.org/one\n\ngo 1.27.3\ntoolchain go1.27.4\n", "go1.27.3", "go1.27.4", false},
		{"missing directive", "module example.org/one\n", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseRequirement([]byte(test.source))
			if (err != nil) != test.fails {
				t.Fatalf("unexpected error: %v", err)
			}
			if !test.fails && (actual.Language != test.language || actual.Toolchain != test.toolchain) {
				t.Fatalf("unexpected requirement: %+v", actual)
			}
		})
	}
}
