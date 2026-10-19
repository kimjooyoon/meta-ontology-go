package syntaxregistration

import (
	"context"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestProfileHintUsesPinnedReplayAndGoooChoice(t *testing.T) {
	repository := osProjectFS()
	for _, row := range []struct {
		path, want, status string
		matches            int
	}{
		{"examples/billing/main.gooo", "ordinary", "AMBIGUOUS", 8},
		{"examples/collection-filter/main.gooo", "gooo.entityfields.go-projection.v5", "AMBIGUOUS", 2},
		{"examples/scope-fragments/main.gooo", "gooo.entityfields.go-projection.v6", "HINT", 1},
	} {
		t.Run(row.path, func(t *testing.T) {
			report := InspectProfileHint(context.Background(), repository, row.path)
			if report.Status != row.status || report.SuggestedProfile != row.want ||
				report.MatchCount != row.matches ||
				report.MatchCount != len(report.CompatibleProfiles) ||
				!slices.Contains(report.CompatibleProfiles, row.want) ||
				report.Authority != "ADVISORY_ONLY" || len(report.Profiles) != len(hintProfiles) ||
				report.ActivityID != "gooo://syntax-registration/suggest-profile" ||
				report.ActivitySource != digest(contractSource) || report.SourceDigest == "" ||
				report.ProjectionDigest == "" {
				t.Fatalf("unexpected profile hint: %+v", report)
			}
		})
	}
	path := "examples/collection-filter/main.gooo"
	changed := strings.Replace(string(contractSource), "} else if input.v5 {",
		"} else if false && input.v5 {", 1)
	if changed == string(contractSource) {
		t.Fatal("Gooo intervention did not change the profile decision")
	}
	other := inspectProfileHintSource(context.Background(), repository, path, []byte(changed))
	if other.Status != "AMBIGUOUS" || other.SuggestedProfile != "gooo.entityfields.go-projection.v6" ||
		other.ActivitySource != digest([]byte(changed)) {
		t.Fatalf("Gooo source did not own the advisory choice: %+v", other)
	}
	wrongCount := strings.Replace(string(contractSource),
		"if input.v5 { matches = matches + 1 }", "if input.v5 { matches = matches + 2 }", 1)
	if wrongCount == string(contractSource) {
		t.Fatal("Gooo intervention did not change the match count")
	}
	invalid := inspectProfileHintSource(context.Background(), repository, path, []byte(wrongCount))
	if invalid.Status != "UNKNOWN" || invalid.Reason != "HINT_CONTRACT_MISMATCH" {
		t.Fatalf("mismatched Gooo count was trusted: %+v", invalid)
	}
}

func TestProfileHintBadSourceIsAdvisory(t *testing.T) {
	repository := fstest.MapFS{"bad.gooo": &fstest.MapFile{Data: []byte("package broken\nentity")}}
	bad := InspectProfileHint(context.Background(), repository, "bad.gooo")
	if bad.Status != "NO_MATCH" || bad.SuggestedProfile != "" ||
		bad.MatchCount != 0 || len(bad.CompatibleProfiles) != 0 ||
		len(bad.Profiles) != len(hintProfiles) {
		t.Fatalf("invalid source received a profile: %+v", bad)
	}
	missing := InspectProfileHint(context.Background(), repository, "missing.gooo")
	if missing.Status != "UNKNOWN" || missing.Reason != "SOURCE_UNREADABLE" ||
		len(missing.Profiles) != 0 {
		t.Fatalf("missing source should remain advisory unknown: %+v", missing)
	}
	invalidPath := InspectProfileHint(context.Background(), repository, "../bad.gooo")
	if invalidPath.Status != "UNKNOWN" || invalidPath.Reason != "INVALID_SOURCE_PATH" {
		t.Fatalf("invalid path should remain advisory unknown: %+v", invalidPath)
	}
}

func osProjectFS() fs.FS { return os.DirFS("../../..") }
