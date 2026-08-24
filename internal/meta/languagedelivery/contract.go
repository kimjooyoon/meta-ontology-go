package languagedelivery

func CanonicalContract() Contract {
	obligations := append(userObligations(), toolObligations()...)
	obligations = append(obligations, governorObligations()...)
	return Contract{
		Schema: ContractSchema, ContractID: ContractID, Version: 1,
		Scope: "Gooo v0.1 observable delivery; not universal language completeness",
		AudienceOrder: append([]Audience(nil), audienceOrder...),
		Obligations: obligations,
		NotClaimed: []string{
			"universal programming-language completeness",
			"production readiness",
			"historical improvement without a comparable predecessor",
			"runtime capability without an executable receipt",
		},
		References: []Reference{
			{ID: "go-command", URL: "https://pkg.go.dev/cmd/go", Authority: "SCOPE_REFERENCE_ONLY"},
			{ID: "go-1.27", URL: "https://go.dev/doc/go1.27", Authority: "SCOPE_REFERENCE_ONLY"},
			{ID: "gopls-features", URL: "https://go.dev/gopls/features/", Authority: "SCOPE_REFERENCE_ONLY"},
			{ID: "gomacro", URL: "https://github.com/cosmos72/gomacro", Authority: "SCOPE_REFERENCE_ONLY"},
		},
	}
}

func obligation(id string, audience Audience, class IndicatorClass, outcome string, evidence EvidenceRule, operation string, proof ProofChoice) Obligation {
	return Obligation{ID: id, Audience: audience, Class: class, Outcome: outcome,
		Evidence: evidence, MetaOperation: operation, ProofChoice: proof}
}

func rule(source SourceName, kind EvidenceKind, id, counter string, target int) EvidenceRule {
	return EvidenceRule{Source: source, Kind: kind, ID: id, Counter: counter, Target: target}
}

func missing() EvidenceRule {
	return rule(SourceNone, EvidenceUnimplemented, "", "", 1)
}
