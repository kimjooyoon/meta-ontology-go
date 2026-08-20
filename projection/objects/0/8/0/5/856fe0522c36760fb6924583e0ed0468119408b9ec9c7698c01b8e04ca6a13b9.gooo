package bidir

import (
	"reflect"
	"testing"
)

func TestHostingContractsDoNotOverclaimFutureSelfHosting(t *testing.T) {
	current := InitialGoHostedContract()
	target := PlannedGoooHostedContract()
	if err := current.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := target.Validate(); err != nil {
		t.Fatal(err)
	}
	if current.Phase != HostPhaseGoHosted || target.Phase != HostPhaseGoooHosted {
		t.Fatalf("unexpected phase transition: %s -> %s", current.Phase, target.Phase)
	}
	if target.Verified() {
		t.Fatal("planned gooo-hosted phase was marked verified")
	}
	want := []string{"go-projection-reconcile", "gooo-hosted-self-hosting", "semantic-ir-lowering"}
	if got := target.UnverifiedChecks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("unverified checks = %#v, want %#v", got, want)
	}
}

func TestHostingComparisonReportsEvidenceGap(t *testing.T) {
	comparison, err := CompareHostingContracts(InitialGoHostedContract(), PlannedGoooHostedContract())
	if err != nil {
		t.Fatal(err)
	}
	if !comparison.HostChanged || comparison.From != HostPhaseGoHosted || comparison.To != HostPhaseGoooHosted {
		t.Fatalf("unexpected hosting comparison: %#v", comparison)
	}
	if len(comparison.AddedChecks) != 0 {
		t.Fatalf("existing checks were reported as new: %#v", comparison.AddedChecks)
	}
	if !reflect.DeepEqual(comparison.Remaining, []string{"go-projection-reconcile", "gooo-hosted-self-hosting", "semantic-ir-lowering"}) {
		t.Fatalf("remaining evidence = %#v", comparison.Remaining)
	}
}

func TestHostingContractRejectsDuplicateEvidence(t *testing.T) {
	contract := HostingContract{
		Phase:             HostPhaseGoHosted,
		HostLanguage:      "go",
		AuthoritativeView: ".gooo DSL",
		Evidence: []ContractEvidence{
			{Check: "delta", State: EvidenceVerified},
			{Check: " delta ", State: EvidencePlanned},
		},
	}
	if err := contract.Validate(); err == nil {
		t.Fatal("duplicate evidence was accepted")
	}
}
