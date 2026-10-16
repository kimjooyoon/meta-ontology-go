package syntaxregistration

import (
	"bytes"
	"testing"
)

func TestProfileMigrationObservationFresh1852(t *testing.T) {
	files, request, replacement := profilePlanFixture1848(t)
	plan, err := CompileProfileMigration(files, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(files)
	if err != nil {
		t.Fatal(err)
	}
	before, after := candidate.SourceChecks[0], candidate.SourceChecks[1]
	if before.SourceDigest != request.SourceDigest || after.SourceDigest != request.ReplacementDigest ||
		after.EntityFieldsProfile != request.TargetProfile || len(after.ProfileDigest) != 71 ||
		before.EntityFieldsProfile != "" || before.SemanticDigest == after.SemanticDigest ||
		!before.GetPut || !before.PutGet || !after.GetPut || !after.PutGet {
		t.Fatal("actual before/after parser origins or BX laws missing")
	}
	candidate.SourceChecks[1].ProfileDigest = candidate.SourceChecks[0].SourceDigest
	if err := plan.ValidateCandidate(files, candidate); err == nil {
		t.Fatal("forged kernel-profile origin accepted")
	}
	request.Case.ProofChoice = "invented-proof"
	if _, err := CompileProfileMigration(files, request, replacement); err == nil {
		t.Fatal("invalid proof family accepted")
	}
}

func TestProfileMigrationNativeBoundariesFresh1853(t *testing.T) {
	files, request, replacement := profilePlanFixture1848(t)
	path := syntaxRoot + "owner1848.go"
	for _, name := range []string{"native-membership", "unrelated-format"} {
		t.Run(name, func(t *testing.T) {
			original := files[path].Data
			defer func() { files[path].Data = original }()
			if name == "native-membership" {
				files[path].Data = bytes.Replace(original, []byte(profileContractPath), []byte("other.gooo"), 1)
			} else {
				files[path].Data = append([]byte("\n"), original...)
			}
			var err error
			request.SnapshotDigest, request.SourceDigest, request.ReplacementDigest, err =
				InspectProfileMigrationInputs(files, request, replacement)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := CompileProfileMigration(files, request, replacement)
			if name == "unrelated-format" && err == nil {
				_, err = plan.Generate(files)
			}
			if err == nil {
				t.Fatal("source membership or unrelated byte change accepted")
			}
		})
	}
}
