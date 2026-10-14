package verify

import (
	"os"
	"strings"
	"testing"
)

func TestRacePackageOrderRetainsFullScope58201(t *testing.T) {
	path := os.Getenv("GOOO_RACE_ORDER_WORKFLOW58201")
	if path == "" {
		path = "../../.github/workflows/ci.yml"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, body, found := strings.Cut(string(raw), "\n  race:\n")
	if !found {
		t.Fatal("race job missing")
	}
	body, _, found = strings.Cut(body, "\n  semantic:\n")
	if !found || !strings.Contains(body, "    name: go test -race\n") ||
		!strings.Contains(body, "    timeout-minutes: 20\n") ||
		!strings.Contains(body, "          cache: false\n") {
		t.Fatal("race identity, time budget or fresh-build boundary changed")
	}
	commands := []string{}
	for line := range strings.SplitSeq(body, "\n") {
		if command, found := strings.CutPrefix(line, "      - run: "); found {
			commands = append(commands, command)
		}
	}
	if len(commands) != 1 || !racePackageOrderCommand58201(commands[0]) {
		t.Fatalf("race command changed its complete scope: %v", commands)
	}
}

func TestRacePackageOrderRejectsNarrowedCommands58201(t *testing.T) {
	commands := []string{
		"go test -race -json ./...",
		"go test -race -json ./internal/meta/repositoryprojection/extractor ./... ./...",
		"go test -race -json ./internal/meta/repositoryprojection/extractor",
		"go test -json ./internal/meta/repositoryprojection/extractor ./...",
		"go test -race -json -short ./internal/meta/repositoryprojection/extractor ./...",
		"go test -race -json -run=TestOne ./internal/meta/repositoryprojection/extractor ./...",
		"go test -race -json ./internal/meta/repositoryprojection/extractor ./... || true",
	}
	for _, command := range commands {
		if racePackageOrderCommand58201(command) {
			t.Fatalf("unexpected race schedule accepted: %s", command)
		}
	}
}

func racePackageOrderCommand58201(command string) bool {
	return strings.Join(strings.Fields(command), " ") ==
		"go test -race -json ./internal/meta/repositoryprojection/extractor ./..."
}
