package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func partitionNativeFiles(t *testing.T) (fstest.MapFS, sourcebundle.Partition) {
	t.Helper()
	repository, partition := nativePartitionFixture(t, "examples/entity-fields-v1", []string{"main.gooo"},
		"main.gooo", "", "activity.gooo", syntax.EntityFieldsV1Support())
	inputs, err := readPartitionNativeInputs(repository)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := fs.ReadFile(repository, "examples/entity-fields-v1/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	inputs["examples/entity-fields-v1/main.gooo"] = raw
	files := fstest.MapFS{}
	for name, raw := range inputs {
		files[name] = &fstest.MapFile{Data: bytes.Clone(raw)}
	}
	return files, partition
}

func TestPartitionNativeOwnershipFollowsDeclarationsAcrossFiles(t *testing.T) {
	files, partition := partitionNativeFiles(t)
	old := partitionNativeFieldOwner(t, files, "CaseSources", "CaseSourceSet", "explicit-consumer-continuation")
	renamed := syntaxRoot + "physical-owners.go"
	files[renamed] = files[old]
	delete(files, old)
	const incidental = "\n// keep incidental origin\nfunc unrelatedPartitionMembers() []string { return []string{\"keep.gooo\"} }\n"
	files[renamed].Data = append(files[renamed].Data, []byte(incidental)...)
	candidate, err := PreparePartitionRegistrationCandidate(files, "examples/entity-fields-v1", partition)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.Outputs[renamed] == nil || candidate.Outputs[old] != nil ||
		!bytes.Contains(candidate.Outputs[renamed], []byte("// keep incidental origin")) ||
		!bytes.Contains(candidate.Outputs[renamed], []byte(`[]string{"keep.gooo"}`)) {
		t.Fatal("native ownership was inferred from a filename or incidental literal")
	}
	files[syntaxRoot+"extra.go"] = &fstest.MapFile{Data: []byte("package languagesyntax\nfunc incidentalOwner() {}\n")}
	if err := ValidatePartitionRegistrationCandidate(files, "examples/entity-fields-v1", partition, candidate); err == nil {
		t.Fatal("native file population changed without invalidating the candidate")
	}
}

func TestPartitionNativeRefusesDivergentAndUnboundedOwners(t *testing.T) {
	for _, item := range []struct{ name, from, to string }{
		{"case", "", ""},
		{"family", "explicit-consumer-continuation", "wrong-family"},
		{"meta", "examples/agent-tool-inventory/main.gooo", "examples/unknown/main.gooo"},
		{"package-entry", `Entry: "PayOrder"`, `Entry: "Other"`},
		{"package-members", "examples/billing-package/entities.gooo", "examples/billing-package/other.gooo"},
		{"package-return", "packageUnit, symbolicUnit", "packageUnit, packageUnit"},
		{"package-mutation", "return []PackageDefinition", `packageUnit.Entry = "PayOrder"; return []PackageDefinition`},
		{"package-import", "internal/meta/languagepackageexecution", "internal/otherexecution"},
		{"package-import-shadow", "packageUnit", "languagepackageexecution"},
		{"package-type-shadow", "packageUnit", "PackageDefinition"},
		{"projection", "internal/completeness/receipt.gooo", "examples/entity-fields-v1/main.gooo"},
	} {
		t.Run(item.name, func(t *testing.T) {
			files, partition := partitionNativeFiles(t)
			if item.name == "case" {
				// Resolve the actual case owner rather than assuming its filename.
				inputs := map[string][]byte{}
				for name, file := range files {
					inputs[name] = file.Data
				}
				source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := readProfileNativeCases(source)
				if err != nil {
					t.Fatal(err)
				}
				name, _ := source.location(rows[0].literal.Pos())
				files[name].Data = bytes.Replace(files[name].Data, []byte(rows[0].definition.ID), []byte("divergent-case"), 1)
			} else {
				name := partitionNativeCounterexampleOwner(t, files, item.name, item.from)
				if !bytes.Contains(files[name].Data, []byte(item.from)) {
					t.Fatal("counterexample did not match an actual owner")
				}
				files[name].Data = bytes.ReplaceAll(files[name].Data, []byte(item.from), []byte(item.to))
			}
			result, err := PreparePartitionRegistrationCandidate(files, "examples/entity-fields-v1", partition)
			if err == nil || !reflect.DeepEqual(result, PartitionRegistrationCandidate{}) {
				t.Fatal("divergent native owner yielded a candidate", err)
			}
		})
	}
}

func TestPartitionNativeAppendsToASequenceOfLocalFamilyDeclarations(t *testing.T) {
	files, partition := partitionNativeFiles(t)
	inputs := map[string][]byte{}
	for name, file := range files {
		inputs[name] = file.Data
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	function, err := source.function("expectedCaseSources")
	if err != nil {
		t.Fatal(err)
	}
	name, _ := source.location(function.Pos())
	old := source.text(function)
	local := strings.Replace(old, "return []CaseSourceSet{{", "owned := CaseSourceSet{", 1)
	local = strings.Replace(local, "}}}\n}", "}}\nreturn []CaseSourceSet{owned}\n}", 1)
	if local == old || !strings.Contains(local, "return []CaseSourceSet{owned}") {
		t.Fatal("local-owner fixture was not constructed")
	}
	files[name].Data = bytes.Replace(files[name].Data, []byte(old), []byte(local), 1)
	candidate, err := PreparePartitionRegistrationCandidate(files, "examples/entity-fields-v1", partition)
	if err != nil {
		t.Fatal(err)
	}
	assertCompiledPartitionRegistry(t, candidate)
	var altered languagesyntax.Registry
	if err := json.Unmarshal(candidate.Projection.Corpus, &altered); err != nil {
		t.Fatal(err)
	}
	altered.Cases[0].ID = "wrong-oracle-expectation"
	candidate.Projection.Corpus, err = json.Marshal(altered)
	if err != nil {
		t.Fatal(err)
	}
	if err := compiledPartitionRegistry(t, candidate); err == nil ||
		!strings.Contains(err.Error(), "compiled registry differs from projected corpus") {
		t.Fatal("independent compiler oracle did not reject the wrong registry value", err)
	}
}
