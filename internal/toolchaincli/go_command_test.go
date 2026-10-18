package toolchaincli

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestGeneratedTestGoInputsWireRepresentation(t *testing.T) {
	digest := strings.Repeat("a", 64)
	inputs := GeneratedTestGoInputs{Profile: GeneratedTestGoProfile, Path: "/sdk/bin/go",
		ExecutableDigest: digest, Version: "go1.27.2", Platform: "darwin/arm64", EnvironmentDigest: digest}
	want := `{"profile":"gooo/generated-test-native-go-command/v1","path":"/sdk/bin/go","executable_digest":"` + digest +
		`","version":"go1.27.2","platform":"darwin/arm64","environment_digest":"` + digest + `"}`
	raw, err := json.Marshal(inputs)
	if err != nil || string(raw) != want || inputs.Validate() != nil {
		t.Fatal("shared observer changed the existing receipt input representation", err)
	}
}

func TestGoCommandEnvironmentIsFrozenAndCanonical(t *testing.T) {
	a := generatedTestGoEnvironment([]string{"HOME=/home/a", "GOCACHE=/cache", "PATH=/bin", "GOFLAGS=-race", "GOENV=/custom", "TOKEN=private"}, "/sdk")
	b := generatedTestGoEnvironment([]string{"TOKEN=other", "PATH=/bin", "GOCACHE=/cache", "HOME=/home/a", "GOFLAGS=-tags=other"}, "/sdk")
	if !slices.Equal(a, b) || !slices.IsSorted(a) {
		t.Fatal("ambient Go settings or credentials changed the selected profile")
	}
	for _, entry := range []string{"GOENV=off", "GOFLAGS=-mod=readonly", "GOTOOLCHAIN=local", "GOWORK=off", "GOROOT=/sdk", "GOPROXY=off", "CGO_ENABLED=0"} {
		if !slices.Contains(a, entry) {
			t.Fatal("missing controlled setting", entry)
		}
	}
	c := generatedTestGoEnvironment([]string{"PATH=/other", "HOME=/home/a", "GOCACHE=/cache"}, "/sdk")
	if slices.Equal(a, c) {
		t.Fatal("actual retained environment change was discarded")
	}
}
