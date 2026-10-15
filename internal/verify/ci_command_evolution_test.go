package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

type commandEvolution struct {
	Schema         string                           `json:"schema"`
	BaselineSHA256 string                           `json:"baseline_sha256"`
	Jobs           map[string][]commandEvolutionRow `json:"jobs"`
}

type commandEvolutionRow struct {
	Original  int    `json:"original_index"`
	Before    string `json:"before_sha256"`
	After     string `json:"after_sha256"`
	Operation string `json:"operation"`
}

func readCommandEvolution(t *testing.T, baseline []byte) commandEvolution {
	t.Helper()
	raw, err := os.ReadFile("../../docs/observations/ci-source-join/command-evolution.json")
	if err != nil {
		t.Fatal(err)
	}
	var ledger commandEvolution
	if err := json.Unmarshal(raw, &ledger); err != nil || ledger.Schema != "gooo/ci-command-evolution/v1" ||
		ledger.BaselineSHA256 != commandSourceDigest(baseline) {
		t.Fatal("command evolution lost the immutable original population", err)
	}
	return ledger
}

func commandSourceDigest(source []byte) string {
	digest := sha256.Sum256(source)
	return hex.EncodeToString(digest[:])
}

func bindCommandEvolution(commands, original []string, rows []commandEvolutionRow) error {
	if len(commands) != len(rows) {
		return fmt.Errorf("current command population lacks explicit origins")
	}
	seen := make([]bool, len(original))
	for i, row := range rows {
		if row.Operation == "" || row.After != commandSourceDigest([]byte(commands[i])) {
			return fmt.Errorf("current command %d differs from its declared origin", i)
		}
		if row.Original == -1 {
			if row.Before != "" {
				return fmt.Errorf("new command claims a prior origin")
			}
			continue
		}
		if row.Original < 0 || row.Original >= len(original) || seen[row.Original] || row.Before != original[row.Original] {
			return fmt.Errorf("original command reference is missing, repeated or changed")
		}
		seen[row.Original] = true
	}
	for _, found := range seen {
		if !found {
			return fmt.Errorf("original command obligation has no descendant")
		}
	}
	return nil
}

func TestCommandEvolutionBindsSourceWithoutLosingAnOriginalObligation(t *testing.T) {
	original := []string{commandSourceDigest([]byte("old fixture")), commandSourceDigest([]byte("old proof"))}
	commands := []string{"bind tool", "fresh fixture", "fresh proof"}
	rows := []commandEvolutionRow{
		{Original: -1, After: commandSourceDigest([]byte(commands[0])), Operation: "bind"},
		{Original: 0, Before: original[0], After: commandSourceDigest([]byte(commands[1])), Operation: "reproduce"},
		{Original: 1, Before: original[1], After: commandSourceDigest([]byte(commands[2])), Operation: "prove"},
	}
	if err := bindCommandEvolution(commands, original, rows); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing", "duplicate", "old", "new", "added-prior"} {
		changed := append([]commandEvolutionRow(nil), rows...)
		switch mutation {
		case "missing":
			changed[2].Original, changed[2].Before = -1, ""
		case "duplicate":
			changed[2].Original, changed[2].Before = 0, original[0]
		case "old":
			changed[1].Before = strings.Repeat("0", 64)
		case "new":
			changed[1].After = strings.Repeat("0", 64)
		case "added-prior":
			changed[0].Before = original[0]
		}
		if err := bindCommandEvolution(commands, original, changed); err == nil {
			t.Fatal("invalid evolution accepted", mutation)
		}
	}
}
