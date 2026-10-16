package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestProfileMigrationPlanFresh1848(t *testing.T) {
	files, request, replacement := profilePlanFixture1848(t)
	before := digestValue(files)
	plan, err := CompileProfileMigration(files, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	replacement[0] = '!'
	candidate, err := plan.Generate(files)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateCandidate(files, candidate); err != nil {
		t.Fatal(err)
	}
	if candidate.CasePopulation != 1 || candidate.RequiredMembers != 3 || candidate.RepositoryWrites != 0 ||
		candidate.ApplyAuthorized || candidate.PromotionAllowed || candidate.State != "PROPOSAL_ONLY" ||
		candidate.Admission != "UNASSESSED" || candidate.BeforeCase != request.Case ||
		candidate.AfterCase.ID != request.Case.ID || candidate.AfterCase.Path != request.Case.Path ||
		candidate.AfterCase.EntityFields || candidate.AfterCase.EntityFieldsProfile != request.TargetProfile ||
		candidate.ContractDigest == "" || candidate.SemanticDigest == "" || digestValue(files) != before {
		t.Fatal("migration lost identity, source or authority boundaries")
	}
	for _, member := range candidate.Members {
		if member.BeforeDigest != digest(files[member.Path].Data) || member.BeforeDigest != digest(member.BeforeContent) ||
			member.AfterDigest != digest(member.Content) || member.ActivityID == "" || member.OutputID == "" ||
			strings.Contains(member.Path, "denominator") {
			t.Fatalf("member lost exact source role: %s", member.Path)
		}
	}
	var registry languagesyntax.Registry
	if err := json.Unmarshal(candidate.Members[0].Content, &registry); err != nil || len(registry.Cases) != 1 ||
		registry.Cases[0] != candidate.AfterCase || !bytes.Contains(candidate.Members[1].Content, []byte("ignored1848")) {
		t.Fatalf("migration changed population or incidental source: %v", err)
	}
}

func TestProfileMigrationCandidateRefusalsFresh1849(t *testing.T) {
	files, request, replacement := profilePlanFixture1848(t)
	plan, err := CompileProfileMigration(files, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"authority", "population", "origin-bytes", "resealed-output", "semantic-role"} {
		t.Run(name, func(t *testing.T) {
			candidate, err := plan.Generate(files)
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "authority":
				candidate.ApplyAuthorized = true
			case "population":
				candidate.CasePopulation++
			case "origin-bytes":
				candidate.Members[0].BeforeContent[0] = '!'
			case "resealed-output":
				candidate.Members[2].Content[0] = '!'
				candidate.Members[2].AfterDigest = digest(candidate.Members[2].Content)
			case "semantic-role":
				candidate.Members[2].OutputID = candidate.Members[0].OutputID
			}
			if err := plan.ValidateCandidate(files, candidate); err == nil ||
				err.Error() != "REFUTED/PROFILE_CANDIDATE_DIVERGED" {
				t.Fatalf("candidate mutation accepted: %v", err)
			}
		})
	}
}

func TestProfileMigrationRequestRefusalsFresh1850(t *testing.T) {
	files, original, replacement := profilePlanFixture1848(t)
	for _, name := range []string{"unknown-profile", "stale-source", "stale-replacement", "mixed-profile", "invalid-after"} {
		t.Run(name, func(t *testing.T) {
			request, after := original, bytes.Clone(replacement)
			switch name {
			case "unknown-profile":
				request.TargetProfile = "future-profile"
			case "stale-source":
				request.SourceDigest = digest(nil)
			case "stale-replacement":
				after[0] = '!'
			case "mixed-profile":
				request.Case.EntityFieldsProfile = original.TargetProfile
			case "invalid-after":
				after[0] = '!'
				request.ReplacementDigest = digest(after)
			}
			if _, err := CompileProfileMigration(files, request, after); err == nil {
				t.Fatal("invalid migration request accepted")
			}
		})
	}
}

func TestProfileMigrationSnapshotFresh1851(t *testing.T) {
	files, request, replacement := profilePlanFixture1848(t)
	plan, err := CompileProfileMigration(files, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	files[denominatorPath(93)].Data = []byte("changed historical denominator")
	if _, err := plan.Generate(files); err == nil || err.Error() != "UNKNOWN/PROFILE_MIGRATION_INPUT_STALE" {
		t.Fatalf("stale historical source accepted: %v", err)
	}
}
