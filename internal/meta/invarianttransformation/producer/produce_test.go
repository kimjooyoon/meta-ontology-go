package producer

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const testHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestBuildUsesAllFourInvariantValues(t *testing.T) {
	receipt, err := Build([]byte("package fixture\n"), testHead, "preserved-translation")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != model.DecisionAllowed || receipt.Resolution != model.ResolutionExact || len(receipt.Claims) != 4 || len(receipt.Values) != 4 {
		t.Fatalf("receipt=%+v", receipt)
	}
	for index, claim := range receipt.Claims {
		if claim.Status != model.StatusDischarged || len(claim.Transitions) != 1 || claim.Transitions[0].From != model.StatusOpen || claim.Transitions[0].To != model.StatusDischarged {
			t.Fatalf("claim[%d]=%+v", index, claim)
		}
	}
}

func TestBuildRecordsApprovedArtifactWithoutWrite(t *testing.T) {
	receipt, err := Build([]byte("package fixture\n"), testHead, "approved-artifact")
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Effects) != 1 || receipt.Effects[0].Kind != model.EffectApproved || receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		t.Fatalf("approved effect=%+v writes=%d authority=%t", receipt.Effects, receipt.RepositoryWrites, receipt.MutationAuthority)
	}
}
