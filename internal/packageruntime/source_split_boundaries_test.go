package packageruntime

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestSourceSplitBoundariesRetainResolvedIDsAndExactSizes(t *testing.T) {
	manifest := splitFixture()
	manifest.Packages = append(manifest.Packages, PackageSpec{Path: "a/other", Name: "other", Sources: []Source{
		{Filename: "other.gooo", Content: "package other\nnamespace other\nentity A id \"urn:a\"\n"},
	}})
	for _, reverse := range []bool{false, true} {
		if reverse {
			slices.Reverse(manifest.Packages)
		}
		observed, err := InspectSourceSplit(manifest, "split/ledger", "source.gooo")
		if err != nil || len(observed.Candidates) != 2 || observed.SemanticAdmission != "UNASSESSED" ||
			observed.Candidates[0].BeforeID != "urn:split:adjust" || observed.Candidates[1].BeforeID != "urn:split:total" {
			t.Fatal("boundary identity or scope lost", observed, err)
		}
		for _, candidate := range observed.Candidates {
			plan, err := SplitSharedSource(manifest, "split/ledger", "source.gooo", candidate.BeforeID, "new.gooo")
			if err != nil || plan.InputSHA256 != observed.InputSHA256 || plan.InputImageDigest != observed.InputImageDigest ||
				len(plan.Files[0].Source) != candidate.Left.Bytes || len(plan.Files[1].Source) != candidate.Right.Bytes ||
				plan.Files[1].Origins[1].InputStart != candidate.StartByte {
				t.Fatal("observed candidate did not reach the existing split operation", candidate, err)
			}
		}
		replay, err := InspectSourceSplit(manifest, "split/ledger", "source.gooo")
		if err != nil || !reflect.DeepEqual(observed, replay) {
			t.Fatal("boundary observation did not replay", err)
		}
	}
}

func TestSourceSplitBoundaryIsNotSharedConsumerAdmission(t *testing.T) {
	manifest := splitFixture()
	manifest.Packages = append(manifest.Packages, PackageSpec{Path: "other", Name: "other", Sources: []Source{
		{Filename: "source.gooo", Content: "package other\nnamespace other\nentity A id \"urn:a\"\n"},
	}})
	observed, err := InspectSourceSplit(manifest, "split/ledger", "source.gooo")
	if err != nil || len(observed.Candidates) != 2 || observed.SemanticAdmission != "UNASSESSED" {
		t.Fatal("original workspace observation failed", err)
	}
	plan, err := SplitSharedSource(manifest, "split/ledger", "source.gooo", observed.Candidates[0].BeforeID, "new.gooo")
	if err == nil || len(plan.Files) != 0 {
		t.Fatal("observed cut bypassed contradictory shared source bytes", err)
	}
}

func TestSourceSplitBoundariesRequireTheWholeOriginalWorkspace(t *testing.T) {
	for _, mode := range []string{"missing-package", "missing-file", "invalid-other-source"} {
		manifest, pkg, source := splitFixture(), "split/ledger", "source.gooo"
		switch mode {
		case "missing-package":
			pkg = "absent"
		case "missing-file":
			source = "absent.gooo"
		case "invalid-other-source":
			manifest.Packages = append(manifest.Packages, PackageSpec{Path: "other", Sources: []Source{
				{Filename: "bad.gooo", Content: "invalid"}}})
		}
		observed, err := InspectSourceSplit(manifest, pkg, source)
		if err == nil || !reflect.DeepEqual(observed, SourceSplitBoundaries{}) {
			t.Fatal("invalid scope returned candidate prefix", mode, observed, err)
		}
	}
	manifest := splitFixture()
	before, _ := InspectSourceSplit(manifest, "split/ledger", "source.gooo")
	manifest.Packages[0].Sources[0].Content = strings.ReplaceAll(manifest.Packages[0].Sources[0].Content, "Total", "합계")
	after, err := InspectSourceSplit(manifest, "split/ledger", "source.gooo")
	if err == nil || len(after.Candidates) != 0 || len(before.Candidates) != 2 {
		t.Fatal("stale entry must not be hidden by syntax-only observation", err)
	}
	manifest.Entry.Activity = "합계"
	after, err = InspectSourceSplit(manifest, "split/ledger", "source.gooo")
	if err != nil || before.Candidates[1].BeforeID != after.Candidates[1].BeforeID ||
		before.InputSHA256 == after.InputSHA256 || after.Candidates[1].Name != "합계" {
		t.Fatal("display rename lost stable boundary identity", err)
	}
}
