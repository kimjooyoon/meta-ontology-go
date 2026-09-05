package transformationeffect

import "testing"

func TestFirstReplayDifferenceKeyUnionRegression(t *testing.T) {
	tests := []struct {
		name               string
		left, right        map[string]any
		wantPath           string
		wantExpectedValue  string
		wantObservedValue  string
	}{
		{
			name:              "nil left map",
			left:              nil,
			right:             map[string]any{"field": nil},
			wantPath:          "$.field",
			wantExpectedValue: "<missing>",
			wantObservedValue: "null",
		},
		{
			name: "overlapping equal key",
			left: map[string]any{"shared": nil}, right: map[string]any{"shared": nil},
		},
		{
			name:              "disjoint keys",
			left:              map[string]any{"left": nil},
			right:             map[string]any{"right": nil},
			wantPath:          "$.left",
			wantExpectedValue: "null",
			wantObservedValue: "<missing>",
		},
		{
			name:              "asymmetric maps with overlap",
			left:              map[string]any{"common": nil, "left": nil},
			right:             map[string]any{"common": nil},
			wantPath:          "$.left",
			wantExpectedValue: "null",
			wantObservedValue: "<missing>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, expected, observed := firstReplayDifference("$", test.left, test.right)
			if path != test.wantPath || expected != test.wantExpectedValue || observed != test.wantObservedValue {
				t.Fatalf("difference = %q, %q, %q", path, expected, observed)
			}
		})
	}
}
