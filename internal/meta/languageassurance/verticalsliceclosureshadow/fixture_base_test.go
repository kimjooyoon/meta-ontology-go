package verticalsliceclosureshadow

import (
	"encoding/json"
	"strings"
	"testing"
)

func exactInput(t *testing.T) Input {
	t.Helper()
	head := strings.Repeat("b", 40)
	syntax := syntaxFixture(head)
	semantic := semanticFixture(head, syntax)
	binding := bindingFixture(head, semantic)
	useCases := useCaseFixture(head, syntax)
	toolchain := toolchainFixture(head)
	release := releaseFixture(head)
	return Input{HeadSHA: head, Assurance: EmbeddedAssurance(), Artifacts: map[string][]byte{
		"syntax": syntax, "semantics": semantic, "binding": binding,
		"use-cases": useCases, "toolchain": toolchain, "release": release,
	}}
}

func encodeFixture(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mutateFixture(t *testing.T, raw []byte, mutate func(map[string]any)) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	mutate(value)
	return encodeFixture(t, value)
}

func fixtureDigest(value string) string {
	return "sha256:" + strings.Repeat(value, 64)
}
