package syntaxregistration

import (
	"fmt"
	"path"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRegistrationFindsRolesAfterArbitrarySourceFileRelocation(t *testing.T) {
	data, request := fixture(t)
	for index, name := range sortedPaths(data) {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		suffix := ".go"
		if strings.HasSuffix(name, "_test.go") {
			suffix = "_test.go"
		}
		renamed := path.Join(path.Dir(name), fmt.Sprintf("opaque_unit_%04d%s", index, suffix))
		data[renamed] = data[name]
		delete(data, name)
	}
	pin(t, data, &request)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidate.Artifacts) != 9 || candidate.RequiredArtifacts != 9 ||
		candidate.Required != len(candidate.Members) || candidate.Emitted != candidate.Required {
		t.Fatalf("source relocation lost an obligation: %#v", candidate)
	}
	for _, member := range candidate.Members {
		if strings.HasSuffix(member.Path, ".go") && !strings.HasPrefix(path.Base(member.Path), "opaque_unit_") {
			t.Fatalf("generator guessed an old physical filename: %s", member.Path)
		}
	}
	if err := plan.ValidateCandidate(data, candidate); err != nil {
		t.Fatal(err)
	}
}

func TestRegistrationRejectsAmbiguousSymbolOwners(t *testing.T) {
	data, request := fixture(t)
	inputs, err := readInputs(data, request)
	if err != nil {
		t.Fatal(err)
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	function, err := source.function("expectedRegistry")
	if err != nil {
		t.Fatal(err)
	}
	data[syntaxRoot+"ambiguous_owner.go"] = &fstest.MapFile{
		Data: []byte("package languagesyntax\n" + source.text(function) + "\n")}
	pin(t, data, &request)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	requireFailure(t, err, "UNKNOWN", "AMBIGUOUS")
	if candidate.Emitted != 0 || candidate.ApplyAuthorized {
		t.Fatal("ambiguous ownership emitted a candidate")
	}
}

func TestRegistrationPinsAddedSourceUnits(t *testing.T) {
	data, request := fixture(t)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	data[syntaxRoot+"new_source_unit.go"] = &fstest.MapFile{Data: []byte("package languagesyntax\n")}
	candidate, err := plan.Generate(data)
	requireFailure(t, err, "UNKNOWN", "STALE")
	if candidate.Emitted != 0 {
		t.Fatal("an unpinned source unit was accepted")
	}
}

func TestSourceUnitLocationsIgnoreLineDirectiveAliases(t *testing.T) {
	name := syntaxRoot + "actual_owner.go"
	inputs := map[string][]byte{name: []byte("package languagesyntax\n//line outside.go:1\nfunc expectedRegistry() {}\n")}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	function, err := source.function("expectedRegistry")
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "source-bound-test-activity"
	source.insert(function.Body.Rbrace, "\n")
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[name] == nil {
		t.Fatalf("line directive forged source ownership: %v", sortedPaths(changes))
	}
}
