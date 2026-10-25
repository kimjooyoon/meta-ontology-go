package packageruntime

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func sharedSourceFixture() Manifest {
	manifest := splitFixture()
	other := manifest.Packages[0]
	other.Path = "split/other"
	other.Sources = append([]Source(nil), other.Sources...)
	manifest.Packages = append(manifest.Packages, other)
	return manifest
}

func TestSharedSourcePartitionKeepsBothPackageConsumers(t *testing.T) {
	manifest := sharedSourceFixture()
	before, _ := json.Marshal(manifest)
	result, err := SplitSharedSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "moved.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if result.Schema != "gooo/package-source-split/v2" ||
		!reflect.DeepEqual(result.AffectedPackages, []string{"split/ledger", "split/other"}) {
		t.Fatal(result)
	}
	for _, pkg := range result.Manifest.Packages {
		if len(pkg.Sources) != 2 || pkg.Sources[0].Content != result.Files[0].Source ||
			pkg.Sources[1].Filename != "moved.gooo" || pkg.Sources[1].Content != result.Files[1].Source {
			t.Fatal("shared package lost a source member", pkg.Path)
		}
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("shared partition mutated caller source storage")
	}
}

func TestSharedSourceIdentityDoesNotReplacePackageLocalIdentity(t *testing.T) {
	manifest := sharedSourceFixture()
	other := &manifest.Packages[1].Sources[0]
	other.Content = strings.Replace(other.Content, "input+17", "input+19", 1)
	if result, err := SplitSharedSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "moved.gooo"); err == nil || !reflect.DeepEqual(result, SourceSplit{}) {
		t.Fatal("conflicting shared occurrence accepted", err)
	}
	legacy, err := SplitSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "moved.gooo")
	if err != nil || legacy.Schema != "gooo/package-source-split/v1" || legacy.AffectedPackages != nil ||
		len(legacy.Manifest.Packages[1].Sources) != 1 || legacy.Manifest.Packages[1].Sources[0] != *other {
		t.Fatal("package-local source semantics changed", err)
	}
	other.Content = manifest.Packages[0].Sources[0].Content
	other.Filename = "separate.gooo"
	distinct, err := SplitSharedSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "moved.gooo")
	if err != nil || distinct.AffectedPackages != nil || len(distinct.Manifest.Packages[1].Sources) != 1 {
		t.Fatal("identical content at a different address was coupled", err)
	}
}
