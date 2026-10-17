package sourcerepack

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func TestCheckPlansObservesWithoutWriting(t *testing.T) {
	for _, name := range []string{"blocked", "planned", "missing", "unselected"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			source := "package fixture\nfunc First() {}\n"
			path := filepath.Join(root, "fixture.go")
			if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			if name == "planned" {
				if err := os.WriteFile(filepath.Join(root, "target.go"), []byte("package fixture\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var report metricevidence.Report
			report.Meta.Policy.MaxFileLines = 75
			cfg := config{root: root}
			subject := "fixture.go"
			if name == "missing" {
				subject = "missing.go"
			} else if name == "unselected" {
				cfg.subject = "other.go"
			}
			var output strings.Builder
			err := checkPlansTo(&output, cfg, report, []metricevidence.Indicator{{Subject: subject}})
			switch name {
			case "blocked":
				want := "source-repacker: blocked subject=\"fixture.go\" reason=\"no-build-domain-destination\"\n" +
					"source-repacker: actionable=1 planned=0 blocked=1 write=false\n"
				if err != nil || output.String() != want {
					t.Fatalf("blocked plan changed meaning: %q, %v", output.String(), err)
				}
			case "planned":
				if err != nil || output.String() != "source-repacker: actionable=1 planned=1 blocked=0 write=false\n" {
					t.Fatalf("available plan changed meaning: %q, %v", output.String(), err)
				}
			default:
				if err == nil || errors.Is(err, errRepackBlocked) || output.Len() != 0 {
					t.Fatalf("real error became a planning observation: %q, %v", output.String(), err)
				}
				want := "plan missing.go: resolve subject:"
				if name == "unselected" {
					want = "is not an actionable split indicator"
				}
				if !strings.Contains(err.Error(), want) {
					t.Fatal("actual failure reason differs", err)
				}
			}
			if data, err := os.ReadFile(path); err != nil || string(data) != source {
				t.Fatal("planning changed original source", err)
			}
			if name == "planned" {
				if data, err := os.ReadFile(filepath.Join(root, "target.go")); err != nil || string(data) != "package fixture\n" {
					t.Fatal("planning changed destination source", err)
				}
			}
		})
	}
}
