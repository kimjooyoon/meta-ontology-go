package extractor

import (
	"os"
	"testing"
)

func TestCountedReplayCapacity18829(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{
		"internal/bodycodegen#compareSourceFillReplay",
		"internal/bodycodegen#compareSourceFillCompleteness",
	}
	requireNamedCapacities(t, root, targets)
	t.Log("required declaration population", len(targets), "unchanged75-line generated helper cap;package/receiver/function identity survives logical source splits;producer/model0")
}
