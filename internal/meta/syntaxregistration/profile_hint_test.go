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
		{"examples/billing/main.gooo", "ordinary", "AMBIGUOUS", 11},
		{"examples/collection-filter/main.gooo", "ordered-collections", "AMBIGUOUS", 4},
		{"examples/scope-fragments/main.gooo", "gooo.entityfields.go-projection.v6", "AMBIGUOUS", 2},
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
	changed := strings.Replace(string(contractSource), "if result.match_count == 0 {",
		`if (result.match_count == 0 || result.profile == "") && profile.profile != "ordered-collections" {`, 1)
	if changed == string(contractSource) {
		t.Fatal("Gooo intervention did not change the profile decision")
	}
	other := inspectProfileHintSource(context.Background(), repository, path, []byte(changed))
	if other.Status != "AMBIGUOUS" || other.SuggestedProfile != "gooo.entityfields.go-projection.v5" ||
		other.ActivitySource != digest([]byte(changed)) {
		t.Fatalf("Gooo source did not own the advisory choice: %+v", other)
	}
	wrongCount := strings.Replace(string(contractSource),
		"result.match_count = result.match_count + 1", "result.match_count = result.match_count + 2", 1)
	if wrongCount == string(contractSource) {
		t.Fatal("Gooo intervention did not change the match count")
	}
	invalid := inspectProfileHintSource(context.Background(), repository, path, []byte(wrongCount))
	if invalid.Status != "UNKNOWN" || invalid.Reason != "HINT_CONTRACT_MISMATCH" {
		t.Fatalf("mismatched Gooo count was trusted: %+v", invalid)
	}
}

func TestProfileHintUsesExplicitCapabilityWithoutChangingCaseIntent(t *testing.T) {
	repository := osProjectFS()
	for _, row := range []struct {
		path, requirement, want, status string
		matches                         int
	}{
		{"examples/billing/main.gooo", "entity-string", "gooo.entityfields.go-projection.v1", "MINIMAL_HINT", 7},
		{"examples/billing/main.gooo", "entity-boolean", "gooo.entityfields.go-projection.v2", "MINIMAL_HINT", 6},
		{"examples/billing/main.gooo", "entity-integer", "gooo.entityfields.go-projection.v3", "MINIMAL_HINT", 5},
		{"examples/billing/main.gooo", "entity-optional", "gooo.entityfields.go-projection.v4", "MINIMAL_HINT", 4},
		{"examples/collection-filter/main.gooo", "entity-many", "gooo.entityfields.go-projection.v5", "MINIMAL_HINT", 3},
		{"examples/billing/main.gooo", "entity-text-bound", "gooo.entityfields.go-projection.v6", "MINIMAL_HINT", 2},
		{"examples/billing/main.gooo", "ordinary", "ordinary", "HINT", 1},
		{"examples/billing/main.gooo", "text-fields", "text-fields", "HINT", 1},
		{"examples/billing/main.gooo", "implicit-ports", "implicit-ports", "HINT", 1},
		{"examples/collection-filter/main.gooo", "ordered-collections", "ordered-collections", "HINT", 1},
	} {
		t.Run(row.requirement, func(t *testing.T) {
			report := InspectProfileHintForCapability(context.Background(), repository, row.path, row.requirement)
			if report.RequiredCapability != row.requirement || report.Status != row.status ||
				report.SuggestedProfile != row.want || report.MatchCount != row.matches ||
				report.MatchCount != len(report.CompatibleProfiles) || len(report.Profiles) != len(hintProfiles) ||
				report.Authority != "ADVISORY_ONLY" || report.ActivitySource != digest(contractSource) {
				t.Fatalf("explicit capability did not constrain the Gooo hint: %+v", report)
			}
		})
	}
	changed := strings.Replace(string(contractSource), "if result.match_count == 0 {",
		`if (result.match_count == 0 || result.profile == "") && profile.profile != "gooo.entityfields.go-projection.v5" {`, 1)
	if changed == string(contractSource) {
		t.Fatal("Gooo minimum choice was not found")
	}
	invalid := inspectProfileHintSourceForCapability(context.Background(), repository,
		"examples/collection-filter/main.gooo", "entity-many", []byte(changed))
	if invalid.Status != "UNKNOWN" || invalid.Reason != "HINT_CONTRACT_MISMATCH" {
		t.Fatalf("nonminimum Gooo choice was trusted: %+v", invalid)
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
	invalidRequirement := InspectProfileHintForCapability(context.Background(), repository, "bad.gooo", "unknown")
	if invalidRequirement.Status != "UNKNOWN" || invalidRequirement.Reason != "INVALID_REQUIRED_CAPABILITY" ||
		len(invalidRequirement.Profiles) != 0 {
		t.Fatalf("unknown capability received an inference: %+v", invalidRequirement)
	}
	unsupported := InspectProfileHintForCapability(context.Background(), osProjectFS(),
		"examples/collection-filter/main.gooo", "ordinary")
	if unsupported.Status != "NO_MATCH" || unsupported.MatchCount != 0 ||
		unsupported.SuggestedProfile != "" || len(unsupported.Profiles) != len(hintProfiles) {
		t.Fatalf("unsupported requested route received a hint: %+v", unsupported)
	}
}

func osProjectFS() fs.FS { return os.DirFS("../../..") }
