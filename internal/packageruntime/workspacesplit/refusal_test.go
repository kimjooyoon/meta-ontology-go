package workspacesplit_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func TestJointSplitRefusesPartialConflictingOrContentOnlyCoupling(t *testing.T) {
	for _, mode := range []string{"empty", "too many", "blank address", "duplicate owner", "address alias",
		"same content elsewhere", "different destination", "missing boundary", "missing dependency", "overlapping member"} {
		t.Run(mode, func(t *testing.T) {
			files, owners := sharedOwners(t)
			switch mode {
			case "empty":
				owners = nil
			case "too many":
				owners = make([]workspacesplit.Owner, workspacesplit.MaxOwners+1)
			case "blank address":
				owners[1].ManifestPath = ""
			case "duplicate owner":
				owners[1].ManifestPath = owners[0].ManifestPath
			case "address alias":
				owners[1].ManifestPath = "nested/../first.workspace.json"
			case "same content elsewhere":
				files["other.gooo"] = &fstest.MapFile{Data: append([]byte{}, files["shared.gooo"].Data...)}
				owners[1].Definition.Packages = append(owners[1].Definition.Packages[:0:0], owners[1].Definition.Packages...)
				owners[1].Definition.Packages[0].Sources = []string{"other.gooo"}
				owners[1].Request.Source = "other.gooo"
			case "different destination":
				owners[1].Request.Into = "different.gooo"
			case "missing boundary":
				owners[1].Request.BeforeID = "urn:absent"
			case "missing dependency":
				owners[1].Definition.Packages = append(owners[1].Definition.Packages[:0:0], owners[1].Definition.Packages...)
				owners[1].Definition.Packages[0].Sources = []string{"shared.gooo", "missing.gooo"}
			case "overlapping member":
				owners[1].ManifestPath = "run.gooo"
			}
			plan, err := workspacesplit.Prepare(files, owners)
			if err == nil || !reflect.DeepEqual(plan, workspacesplit.Plan{}) {
				t.Fatal("partial draft escaped", mode, plan, err)
			}
		})
	}
}

func TestJointSplitBoundsAggregateSourcePayload(t *testing.T) {
	files, owners := sharedOwners(t)
	note := []byte("package shared\nnamespace shared\n#" + strings.Repeat("x", 950000) + "\n")
	for i := range 9 {
		name := fmt.Sprintf("note%d.gooo", i)
		files[name] = &fstest.MapFile{Data: note}
		owners[0].Definition.Packages[0].Sources = append(owners[0].Definition.Packages[0].Sources, name)
	}
	plan, err := workspacesplit.Prepare(files, owners)
	if err == nil || !strings.Contains(err.Error(), "aggregate source byte limit") ||
		!reflect.DeepEqual(plan, workspacesplit.Plan{}) {
		t.Fatal("joint plans escaped the aggregate byte bound", plan, err)
	}
}
