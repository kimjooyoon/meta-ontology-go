package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func profileFamilyFixture(t *testing.T) (fstest.MapFS, ProfileMigrationRequest, []byte, string) {
	t.Helper()
	files, request, replacement := profilePlanFixture1848(t)
	secondary := "examples/profile1848/echo.gooo"
	before := files[request.Case.Path].Data
	cut := bytes.Index(before, []byte("activity Echo"))
	files[secondary] = &fstest.MapFile{Data: append([]byte("package profile1848\nnamespace profile1848\n"), before[cut:]...)}
	files[request.Case.Path].Data = bytes.Clone(before[:cut])
	replacement = replacement[:bytes.Index(replacement, []byte("activity Echo"))]
	var registry languagesyntax.Registry
	if err := json.Unmarshal(files[corpusPath].Data, &registry); err != nil {
		t.Fatal(err)
	}
	registry.CaseSources = []languagesyntax.CaseSourceSet{{CaseID: request.Case.ID,
		Members: []string{request.Case.Path, secondary}}}
	files[corpusPath].Data, _ = json.Marshal(registry)
	native := files[syntaxRoot+"owner1848.go"]
	native.Data = bytes.Replace(native.Data, []byte("Cases: owned1848()"),
		[]byte("CaseSources: expectedCaseSources(), Cases: owned1848()"), 1)
	files[syntaxRoot+"family1848.go"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`package languagesyntax
func expectedCaseSources() []CaseSourceSet { return []CaseSourceSet{{CaseID: %q, Members: []string{%q, %q}}} }
`, request.Case.ID, request.Case.Path, secondary))}
	repinProfileFamily(t, files, &request, replacement)
	return files, request, replacement, secondary
}

func repinProfileFamily(t *testing.T, files fstest.MapFS, request *ProfileMigrationRequest, replacement []byte) {
	t.Helper()
	var err error
	request.SnapshotDigest, request.SourceDigest, request.ReplacementDigest, err =
		InspectProfileMigrationInputs(files, *request, replacement)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProfileMigrationReplaysTheWholeSourceFamily(t *testing.T) {
	files, request, replacement, secondary := profileFamilyFixture(t)
	plan, err := CompileProfileMigration(files, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(files)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range candidate.SourceChecks {
		if len(check.SourceMembers) != 2 || check.SourceDigestKind != "gooo/source-bundle/v1" ||
			check.SourceMembers[1].Path != secondary || check.SourceMembers[1].SourceDigest != digest(files[secondary].Data) {
			t.Fatal("migration omitted the secondary member", check)
		}
	}
	if candidate.RequiredMembers != 3 || candidate.CasePopulation != 1 || candidate.RepositoryWrites != 0 {
		t.Fatal("an unchanged family member became a write or a new capability")
	}
	candidate.SourceChecks[0].SourceMembers[1].SourceDigest = "changed"
	if err := plan.ValidateCandidate(files, candidate); err == nil {
		t.Fatal("caller mutation rewrote the plan's source observation")
	}
}

func TestProfileMigrationPinsSecondaryMembers(t *testing.T) {
	files, request, replacement, secondary := profileFamilyFixture(t)
	plan, err := CompileProfileMigration(files, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	files[secondary].Data = append(files[secondary].Data, []byte("\n// later origin\n")...)
	if _, err := CompileProfileMigration(files, request, replacement); err == nil {
		t.Fatal("a changed secondary member did not stale the input")
	}
	if _, err := plan.Generate(files); err == nil {
		t.Fatal("generation ignored a changed secondary member")
	}
}

func TestProfileMigrationRefusesInvalidSecondarySource(t *testing.T) {
	files, request, replacement, secondary := profileFamilyFixture(t)
	files[secondary].Data = bytes.ReplaceAll(files[secondary].Data, []byte("Observation"), []byte("MissingType"))
	repinProfileFamily(t, files, &request, replacement)
	if _, err := CompileProfileMigration(files, request, replacement); err == nil {
		t.Fatal("primary replay hid unresolved types in the secondary member")
	}
}

func TestSourceFamilyFragmentCannotBecomeAnotherCapability(t *testing.T) {
	files, request := fixture(t)
	delete(files, request.Case.Path)
	request.Case.Path = "internal/meta/executioncontinuation/policy/source_recovery.gooo"
	request.PromoteMetaSource = true
	files[request.Case.Path] = &fstest.MapFile{Data: []byte(fixtureSource)}
	pin(t, files, &request)
	if _, err := Compile(files, request); err == nil || !strings.Contains(err.Error(), "REGISTRATION_SOURCE_ALREADY_REGISTERED") {
		t.Fatal("physical family membership was promoted to another capability", err)
	}
}

func TestProfileMigrationBindsNativeFamilyMembership(t *testing.T) {
	for _, mode := range []string{"missing-field", "changed-member", "duplicate-field", "unknown-field", "moved-unit"} {
		t.Run(mode, func(t *testing.T) {
			files, request, replacement, secondary := profileFamilyFixture(t)
			native, family := files[syntaxRoot+"owner1848.go"], files[syntaxRoot+"family1848.go"]
			switch mode {
			case "missing-field":
				native.Data = bytes.Replace(native.Data, []byte("CaseSources: expectedCaseSources(), "), nil, 1)
			case "changed-member":
				family.Data = bytes.Replace(family.Data, []byte(secondary), []byte("examples/other.gooo"), 1)
			case "duplicate-field":
				family.Data = bytes.Replace(family.Data, []byte("CaseID: "), []byte("CaseID: \"extra\", CaseID: "), 1)
			case "unknown-field":
				family.Data = bytes.Replace(family.Data, []byte("CaseID: "), []byte("Other: \"extra\", CaseID: "), 1)
			case "moved-unit":
				files[syntaxRoot+"renamed-family.go"] = family
				delete(files, syntaxRoot+"family1848.go")
			}
			repinProfileFamily(t, files, &request, replacement)
			_, err := CompileProfileMigration(files, request, replacement)
			if (err == nil) != (mode == "moved-unit") {
				t.Fatalf("native membership must follow symbols and preserve exact source membership: %v", err)
			}
		})
	}
}

func TestProfileMigrationCannotNarrowAwaySecondaryFields(t *testing.T) {
	files, request, _, secondary := profileFamilyFixture(t)
	const v1 = "gooo.entityfields.go-projection.v1"
	const v2 = "gooo.entityfields.go-projection.v2"
	request.Case.EntityFields, request.Case.EntityFieldsProfile = false, v2
	request.TargetProfile = v1
	var registry languagesyntax.Registry
	if err := json.Unmarshal(files[corpusPath].Data, &registry); err != nil {
		t.Fatal(err)
	}
	registry.Cases[0] = request.Case
	files[corpusPath].Data, _ = json.Marshal(registry)
	native := files[syntaxRoot+"owner1848.go"]
	native.Data = bytes.Replace(native.Data, []byte("EntityFields: true"), []byte(fmt.Sprintf("EntityFieldsProfile: %q", v2)), 1)
	files[secondary].Data = append(files[secondary].Data, []byte(`
entity Flag id "urn:family:flag" fields { field value id "urn:family:flag-value" type boolean required one }
`)...)
	replacement := bytes.Clone(files[request.Case.Path].Data)
	repinProfileFamily(t, files, &request, replacement)
	if _, err := CompileProfileMigration(files, request, replacement); err == nil ||
		!strings.Contains(err.Error(), "PROFILE_SOURCE_CONFORMANCE_FAILED") {
		t.Fatal("target profile ignored an unsupported secondary declaration", err)
	}
}
