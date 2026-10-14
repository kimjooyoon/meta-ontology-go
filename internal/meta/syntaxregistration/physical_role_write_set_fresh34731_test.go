package syntaxregistration

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
	"testing/fstest"
)

func freshAssertRoleWriteSet34731(t *testing.T, candidate Candidate) {
	t.Helper()
	owned := map[string]map[string]bool{}
	roles := map[string]bool{}
	for _, artifact := range candidate.Artifacts {
		if artifact.ActivityID == "" || artifact.OutputID == "" || roles[artifact.OutputID] || len(artifact.Paths) == 0 {
			t.Fatal("missing/repeated semantic role", artifact)
		}
		roles[artifact.OutputID] = true
		seen := map[string]bool{}
		for _, path := range artifact.Paths {
			if seen[path] {
				t.Fatal("role repeats physical source", path)
			}
			seen[path] = true
			if owned[path] == nil {
				owned[path] = map[string]bool{}
			}
			owned[path][artifact.ActivityID] = true
		}
	}
	if len(roles) != 9 || len(owned) != candidate.Required || len(candidate.Members) != len(owned) {
		t.Fatal("complete source-owned union differs from write set")
	}
	seen := map[string]bool{}
	for _, member := range candidate.Members {
		if seen[member.Path] || owned[member.Path] == nil || len(member.Content) == 0 || member.AfterDigest != digest(member.Content) || len(member.ActivityIDs) != len(owned[member.Path]) {
			t.Fatal("missing/extra/changed source member", member.Path)
		}
		seen[member.Path] = true
		activitySeen := map[string]bool{}
		for _, activity := range member.ActivityIDs {
			if activitySeen[activity] {
				t.Fatal("member repeats generating activity", member.Path, activity)
			}
			activitySeen[activity] = true
			if !owned[member.Path][activity] {
				t.Fatal("unowned physical write", member.Path, activity)
			}
		}
	}
}

func TestMultiplePhysicalRoleOrigins34731(t *testing.T) {
	data, request := fixture(t)
	inputs := map[string][]byte{}
	for name, file := range data {
		inputs[name] = file.Data
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://role-origins34731/layout"
	expectedPaths := map[string]bool{}
	wanted := map[string]bool{"totalCases": true, "validCases": true, "FixedCapabilityTotal": true}
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, raw := range group.Specs {
			spec, ok := raw.(*ast.ValueSpec)
			if !ok || len(spec.Names) != 1 || !wanted[spec.Names[0].Name] {
				continue
			}
			name := spec.Names[0].Name
			path := syntaxRoot + "separate_origin34731_" + name + ".go"
			data[path] = &fstest.MapFile{Data: []byte("package languagesyntax\nconst " + source.text(spec) + "\n")}
			expectedPaths[path] = true
			if len(group.Specs) == 1 {
				source.replace(group, "")
			} else {
				source.replace(spec, "")
			}
		}
	}
	if len(expectedPaths) != 3 {
		t.Fatal("three distinct model source owners not constructed")
	}
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	for path, raw := range changes {
		data[path] = &fstest.MapFile{Data: raw}
	}
	delete(data, request.Case.Path)
	request.Case.ID = "multiple-physical-origin34731"
	request.Case.Path = "examples/multiple-physical-origin34731/main.gooo"
	data[request.Case.Path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(fixtureSource, "registrationfixture", "multiplephysical34731"))}
	pin(t, data, &request)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateCandidate(data, candidate); err != nil {
		t.Fatal(err)
	}
	freshAssertRoleWriteSet34731(t, candidate)
	model := candidate.Artifacts[2]
	if len(model.Paths) != 3 {
		t.Fatal("one semantic role lost its three source owners", model.Paths)
	}
	for _, path := range model.Paths {
		if !expectedPaths[path] {
			t.Fatal("model source path was guessed", path)
		}
	}
	t.Log("new3-file model role;all9 semantic roles;exact complete physical union;registry/source/toolchain identity preserved;producer/model0")
}
