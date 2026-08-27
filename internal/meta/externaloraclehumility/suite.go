package externaloraclehumility

import "encoding/json"

func RunCase(caseID string, input Input) Report {
	caseInput := input
	caseInput.Evidence = cloneEvidence(input.Evidence)
	switch caseID {
	case "reference-agreement":
	case "reference-mismatch":
		if len(caseInput.Evidence.References) > 1 {
			caseInput.Evidence.References[1].Relation = "SEMANTIC_AUTHORITY"
			caseInput.Evidence.References[1].Authority = "EXTERNAL_REFERENCE"
		}
	case "reference-absence":
		if len(caseInput.Evidence.References) > 2 {
			caseInput.Evidence.References[2].Available = false
		}
	default:
		return Report{Schema: ReportSchema, SubjectSHA: input.Subject, Decision: "FAIL_CLOSED",
			Resolution: "UNKNOWN", Reason: "UNKNOWN_CASE", ReferenceAgreement: agreementUnknown,
			SemanticAuthority: "GOOO_SOURCE_INTENT", AuthorityGrant: "NONE", EnforcementEffect: "BLOCK"}
	}
	return Judge(caseInput)
}

func RunSuite(contract Contract, input Input) Suite {
	suite := Suite{Schema: SuiteSchema, SubjectSHA: input.Subject,
		CaseDenominatorVersion: CaseDenominatorVersion, CasesTotal: len(contract.Cases),
		OfficialMutations: 0, RepositoryWrites: 0, PromotionCount: 0}
	for _, expected := range contract.Cases {
		report := RunCase(expected.ID, input)
		passed := report.Decision == expected.ExpectedDecision &&
			report.Resolution == expected.ExpectedResolution &&
			report.SemanticAuthority == expected.ExpectedAuthority &&
			report.EnforcementEffect == expected.ExpectedEffect && report.Decision != "PASS"
		suite.Cases = append(suite.Cases, SuiteCase{ID: expected.ID,
			ExpectedDecision: expected.ExpectedDecision, ExpectedResolution: expected.ExpectedResolution,
			ActualDecision: report.Decision, ActualResolution: report.Resolution,
			Authority: report.SemanticAuthority, Effect: report.EnforcementEffect, Passed: passed})
		if passed {
			suite.CasesSatisfied++
		}
	}
	if suite.CasesTotal > 0 {
		suite.CoverageBPS = suite.CasesSatisfied * 10000 / suite.CasesTotal
	}
	if suite.CasesSatisfied == suite.CasesTotal {
		suite.Decision, suite.Resolution, suite.Reason = "HUMILITY_MODEL_BOUND", "EXACT", "ALL_BOUNDARY_CASES_REPLAYED"
	} else {
		suite.Decision, suite.Resolution, suite.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", "BOUNDARY_CASE_MISMATCH"
	}
	suite.SuiteDigest = ""
	suite.SuiteDigest = Digest(suite)
	return suite
}

func cloneEvidence(source ReferenceEvidenceSet) ReferenceEvidenceSet {
	clone := source
	clone.References = append([]ReferenceEvidence(nil), source.References...)
	return clone
}

func Encode(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
