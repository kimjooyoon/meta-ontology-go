package toolchainrelease

import (
	"strconv"
	"strings"
	"testing"
)

type computationMutation struct {
	path   string
	value  any
	remove bool
}

func mutateComputationReceipt(t *testing.T, f computationFixture, mutation computationMutation) {
	t.Helper()
	parts := strings.Split(mutation.path, ".")
	var node any = f.Receipt
	for _, part := range parts[:len(parts)-1] {
		switch x := node.(type) {
		case map[string]any:
			node = x[part]
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(x) {
				t.Fatalf("invalid test path %s", mutation.path)
			}
			node = x[i]
		default:
			t.Fatalf("unresolved test path %s", mutation.path)
		}
	}
	x, ok := node.(map[string]any)
	if !ok {
		t.Fatalf("test path %s is not a field", mutation.path)
	}
	key := parts[len(parts)-1]
	if mutation.remove {
		delete(x, key)
	} else {
		x[key] = mutation.value
	}
}

func rejectComputationMutations(t *testing.T, profile packageProfile, mutations []computationMutation) {
	t.Helper()
	f := newComputationFixture(t, profile)
	if _, err := f.validate(false, boundedTextSmokeBinding{}); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range mutations {
		t.Run(mutation.path, func(t *testing.T) {
			f := newComputationFixture(t, profile)
			mutateComputationReceipt(t, f, mutation)
			if _, err := f.validate(false, boundedTextSmokeBinding{}); err == nil {
				t.Fatal("incomplete or inconsistent synthetic receipt accepted")
			}
		})
	}
}
