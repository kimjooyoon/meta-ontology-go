package workspacesplit_test

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func TestJointSplitRejectsChangingSharedDependency(t *testing.T) {
	files, owners := sharedOwners()
	files["dependency.gooo"] = &fstest.MapFile{Data: []byte("package shared\nnamespace shared\n" +
		"activity Side(Integer) -> Integer id \"urn:shared:side\" computes \"return input+11\"\n")}
	owners[0].Definition.Packages[0].Sources = append(owners[0].Definition.Packages[0].Sources, "dependency.gooo")
	dependencyReads := 0
	reader := changingReader{read: func(path string) ([]byte, error) {
		raw, err := files.ReadFile(path)
		if path == "dependency.gooo" {
			dependencyReads++
			if dependencyReads == 2 {
				raw = []byte(strings.Replace(string(raw), "input+11", "input+12", 1))
			}
		}
		return raw, err
	}}
	plan, err := workspacesplit.Prepare(reader, owners)
	if !errors.Is(err, workspacesplit.ErrSourceChanged) || dependencyReads != 2 ||
		!reflect.DeepEqual(plan, workspacesplit.Plan{}) {
		t.Fatalf("incompatible shared dependency produced %d owner proposals (reads=%d): %v", len(plan.Owners), dependencyReads, err)
	}
}

func TestJointOriginsRetainEqualContentAtDistinctAddresses(t *testing.T) {
	files, owners := sharedOwners()
	for _, name := range []string{"dep-a.gooo", "dep-b.gooo"} {
		files[name] = &fstest.MapFile{Data: []byte("package shared\nnamespace shared\n")}
		owners[0].Definition.Packages[0].Sources = append(owners[0].Definition.Packages[0].Sources, name)
	}
	plan, err := workspacesplit.Prepare(files, owners)
	if err != nil || len(plan.Inputs) != 3 {
		t.Fatal("source occurrences disappeared", err, plan.Inputs)
	}
	first, _ := filepath.Abs("first.workspace.json")
	second, _ := filepath.Abs("second.workspace.json")
	for i, name := range []string{"dep-a.gooo", "dep-b.gooo", "shared.gooo"} {
		address, _ := filepath.Abs(name)
		raw := files[name].Data
		input := plan.Inputs[i]
		if input.Path != address || input.SHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) ||
			input.ByteLength != len(raw) || !reflect.DeepEqual(input.Consumers, []string{first, second}) {
			t.Fatal("source occurrence lost byte identity or consumers", input)
		}
	}
	owners[0], owners[1] = owners[1], owners[0]
	replay, err := workspacesplit.Prepare(files, owners)
	if err != nil || !reflect.DeepEqual(plan.Inputs, replay.Inputs) {
		t.Fatal("input observations depend on consumer enumeration order", err)
	}
}
