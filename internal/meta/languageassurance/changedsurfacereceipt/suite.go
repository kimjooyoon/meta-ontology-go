package changedsurfacereceipt

type CaseSpec struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
}
type CaseResult struct {
	Definition CaseSpec `json:"definition"`
	Passed     bool     `json:"passed"`
	Report     Report   `json:"report"`
}
type Suite struct {
	Schema            string       `json:"schema"`
	SubjectSHA        string       `json:"subject_sha"`
	DenominatorID     string       `json:"denominator_id"`
	DenominatorDigest string       `json:"denominator_digest"`
	Decision          string       `json:"decision"`
	Resolution        string       `json:"resolution"`
	Cases             []CaseResult `json:"cases"`
	CasesTotal        int          `json:"cases_total"`
	CasesPassed       int          `json:"cases_passed"`
	CoverageBPS       int          `json:"coverage_bps"`
	SuiteDigest       string       `json:"suite_digest"`
}

func Denominator() []CaseSpec {
	return []CaseSpec{
		{ID: "exact", ExpectedDecision: DecisionFixedPoint, ExpectedResolution: ResolutionExact, ExpectedReason: ReasonTotal},
		{ID: "zero-change", ExpectedDecision: DecisionFixedPoint, ExpectedResolution: ResolutionExact, ExpectedReason: ReasonTotal},
		{ID: "missing", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedReason: ReasonMissing},
		{ID: "orphan", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedReason: ReasonOrphan},
		{ID: "duplicate", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedReason: ReasonDuplicate},
		{ID: "unknown-top", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedReason: ReasonUnknown},
	}
}

func EvaluateSuite(subjectSHA string) Suite {
	definitions := Denominator()
	cases, passed := make([]CaseResult, 0, len(definitions)), 0
	for _, definition := range definitions {
		report := Evaluate(CaseInput(definition.ID, subjectSHA))
		ok := report.Decision == definition.ExpectedDecision && report.Resolution == definition.ExpectedResolution && report.Reason == definition.ExpectedReason
		if ok {
			passed++
		}
		cases = append(cases, CaseResult{Definition: definition, Passed: ok, Report: report})
	}
	decision, resolution := DecisionFailClosed, ResolutionInvariant
	if passed == len(definitions) {
		decision, resolution = DecisionFixedPoint, ResolutionExact
	}
	suite := Suite{Schema: SuiteSchema, SubjectSHA: subjectSHA, DenominatorID: DenominatorID,
		DenominatorDigest: digestValue(definitions), Decision: decision, Resolution: resolution,
		Cases: cases, CasesTotal: len(definitions), CasesPassed: passed, CoverageBPS: ratio(passed, len(definitions))}
	copy := suite
	copy.SuiteDigest = ""
	suite.SuiteDigest = digestValue(copy)
	return suite
}
