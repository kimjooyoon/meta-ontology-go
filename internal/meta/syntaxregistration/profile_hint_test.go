package syntaxregistration

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestProfileHintUsesPinnedReplayAndGoooChoice(t *testing.T) {
	repository := osProjectFS()
	for _, row := range []struct{ path, want string }{
		{"examples/billing/main.gooo", "ordinary"},
		{"examples/collection-filter/main.gooo", "gooo.entityfields.go-projection.v5"},
		{"examples/scope-fragments/main.gooo", "gooo.entityfields.go-projection.v6"},
	} {
		t.Run(row.path, func(t *testing.T) {
			report := InspectProfileHint(context.Background(), repository, row.path)
			if report.Status != "HINT" || report.SuggestedProfile != row.want ||
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
	if other.Status != "HINT" || other.SuggestedProfile != "gooo.entityfields.go-projection.v6" ||
		other.ActivitySource != digest([]byte(changed)) {
		t.Fatalf("Gooo source did not own the advisory choice: %+v", other)
	}
}

func TestProfileHintBadSourceIsAdvisory(t *testing.T) {
	repository := fstest.MapFS{"bad.gooo": &fstest.MapFile{Data: []byte("package broken\nentity")}}
	bad := InspectProfileHint(context.Background(), repository, "bad.gooo")
	if bad.Status != "NO_MATCH" || bad.SuggestedProfile != "" || len(bad.Profiles) != len(hintProfiles) {
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
