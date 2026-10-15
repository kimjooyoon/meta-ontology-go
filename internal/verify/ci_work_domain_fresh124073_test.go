package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The original command population is a source origin, not a timing estimate.
func TestSharedWorkClockRetainsCommandPopulation124073(t *testing.T) {
	raw, err := os.ReadFile("../../docs/observations/ci-command-domain/predecessor-command-population.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Jobs []struct {
			Job     string   `json:"job"`
			Digests []string `json:"original_command_body_sha256"`
		} `json:"jobs"`
	}
	if err = json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(raw)
	for _, job := range source.Jobs {
		_, body, ok := strings.Cut(workflow, "\n  "+job.Job+":\n")
		if !ok {
			t.Fatal("required job disappeared", job.Job)
		}
		if end := regexp.MustCompile(`(?m)^  [a-z][a-z0-9_-]*:\n`).FindStringIndex(body); end != nil {
			body = body[:end[0]]
		}
		if !strings.Contains(body, "    timeout-minutes: 40\n") ||
			!strings.Contains(body, "uses: ./.github/actions/work-domain") ||
			!strings.Contains(body, "uses: actions/cache/save@v5") ||
			!strings.Contains(body, "Preserve original command work origins") {
			t.Fatal("work clock or origin transport missing", job.Job)
		}
		commands := workCommands124073(body)
		if len(commands) != len(job.Digests) {
			t.Fatal("required command population differs", job.Job, len(commands), len(job.Digests))
		}
		for i, command := range commands {
			digest := sha256.Sum256([]byte(command))
			if hex.EncodeToString(digest[:]) != job.Digests[i] {
				t.Fatal("required command body differs", job.Job, i)
			}
		}
	}
	if strings.Count(workflow, "go install github.com/kimjooyoon/gooo-ci-domain/") != 1 ||
		!strings.Contains(workflow, "shell: bash ./scripts/ci-work-shell.sh {0}") {
		t.Fatal("tool preparation duplicated or command adapter absent")
	}
}

func workCommands124073(body string) []string {
	lines := strings.Split(body, "\n")
	commands := []string{}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		command, ok := strings.CutPrefix(line, "        run: ")
		if !ok {
			command, ok = strings.CutPrefix(line, "      - run: ")
		}
		if !ok {
			continue
		}
		if command != "|" {
			commands = append(commands, command)
			continue
		}
		block := []string{}
		for i+1 < len(lines) {
			next := lines[i+1]
			if next != "" && !strings.HasPrefix(next, "          ") {
				break
			}
			i++
			block = append(block, strings.TrimPrefix(next, "          "))
		}
		commands = append(commands, strings.TrimRight(strings.Join(block, "\n"), "\n")+"\n")
	}
	return commands
}
