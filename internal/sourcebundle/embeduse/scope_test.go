package embeduse_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle/embeduse"
)

func TestBuildContextAndDirectoryAliasRemainExplicit(t *testing.T) {
	root := fixture(t)
	alias := filepath.Join(t.TempDir(), "linked")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFLAGS", "-tags=gooo_alt")
	report, err := embeduse.Inspect(context.Background(), alias, partition())
	if err != nil {
		t.Fatal(err)
	}
	physical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Directory != physical || report.RequestedDirectory != alias || report.Build["GOFLAGS"] != "-tags=gooo_alt" || len(report.Consumers) != 5 {
		t.Fatalf("build/address context was erased: %+v", report)
	}
	for _, consumer := range report.Consumers {
		if consumer.Name == "family" && len(consumer.Members) != 2 {
			t.Fatal("alias made the overlay silently ineffective")
		}
	}
	assertOriginalFiles(t, root)
}

func TestInvalidOrUnavailableInputsReturnNoPartialObservation(t *testing.T) {
	root := fixture(t)
	for name, projection := range map[string][]sourcebundle.Source{
		"outside":   {{Filename: "../contract.gooo", Content: first}},
		"duplicate": {{Filename: "contract.gooo", Content: first}, {Filename: "contract.gooo", Content: first}},
		"oversize":  {{Filename: "contract.gooo", Content: strings.Repeat("x", sourcebundle.MaxSourceBytes+1)}},
	} {
		t.Run(name, func(t *testing.T) {
			report, err := embeduse.Inspect(context.Background(), root, projection)
			if err == nil || !reflect.DeepEqual(report, embeduse.Report{}) {
				t.Fatalf("partial observation on failure: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if report, err := embeduse.Inspect(ctx, root, nil); err == nil || !reflect.DeepEqual(report, embeduse.Report{}) {
		t.Fatal("cancellation returned a report")
	}
	if err := os.Remove(filepath.Join(root, "contract.gooo")); err != nil {
		t.Fatal(err)
	}
	if report, err := embeduse.Inspect(context.Background(), root, nil); err == nil || !reflect.DeepEqual(report, embeduse.Report{}) {
		t.Fatal("missing embed inherited a partial result")
	}
}

func TestUnselectedProjectionMembersRemainVisible(t *testing.T) {
	root := fixture(t)
	sources := append(partition(), sourcebundle.Source{Filename: "unselected.gooo", Content: first})
	report, err := embeduse.Inspect(context.Background(), root, sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Projection) != 3 {
		t.Fatal("unselected projected source disappeared")
	}
	index := report.Projection[2]
	if report.Files[index].Filename != "unselected.gooo" || !report.Files[index].Projected {
		t.Fatal("projection identity changed")
	}
	for _, consumer := range report.Consumers {
		for _, member := range consumer.Members {
			if member == index {
				t.Fatal("unselected member was invented as a consumer input")
			}
		}
	}
}

func TestExistingGoFlagsOverlayIsNotSilentlyReplaced(t *testing.T) {
	root := fixture(t)
	overlay := filepath.Join(root, "existing.json")
	if err := os.WriteFile(overlay, []byte(`{"Replace":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFLAGS", "-overlay="+overlay)
	report, err := embeduse.Inspect(context.Background(), root, partition())
	if err == nil || !strings.Contains(err.Error(), "GOFLAGS") || !reflect.DeepEqual(report, embeduse.Report{}) {
		t.Fatalf("an existing overlay was silently discarded: %v", err)
	}
}
