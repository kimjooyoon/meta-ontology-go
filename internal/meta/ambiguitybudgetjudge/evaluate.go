package ambiguitybudgetjudge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func Evaluate(contractRaw, receiptRaw, source []byte) (Result, error) {
	var contract contract
	var report receipt
	if err := decode(contractRaw, &contract); err != nil {
		return Result{}, err
	}
	if err := decode(receiptRaw, &report); err != nil {
		return Result{}, err
	}
	result := Result{Schema: JudgeSchema, SubjectSHA: report.SubjectSHA, ContractID: report.ContractID,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
		ReportDigest: report.Digest, SourceDigest: digestBytes(source), FixedDenominator: Denominator,
		CasesSatisfied: report.Summary.CasesSatisfied, CasesTotal: report.Summary.CasesTotal,
		CoordinatesSatisfied: report.Summary.CoordinatesSatisfied, CoordinatesTotal: report.Summary.CoordinatesTotal,
		RepositoryWrites: report.Effects.RepositoryWrites, MutationAuthority: report.Effects.MutationAuthority}
	add := func(id string, passed bool, step string) {
		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		result.Checks = append(result.Checks, Check{ID: id, Status: status,
			Coordinate: coordinate{Stage: "ambiguity-budget-judge", Step: step, Reason: status}})
	}

	add("contract-identity", contract.Schema == ContractSchema && contract.ID != "" && contract.SourcePath != "" &&
		contract.SourcePackage != "" && contract.SourceNamespace != "" && contract.FixedDenominator == Denominator &&
		contract.Budget == (integerSet{InterpretationCandidates: 2, UnresolvedBranches: 1, EvidencePaths: 2}) && len(contract.Cases) == CaseTotal && len(contract.NotClaimed) == 4, "contract")

	sourceObservation := observeSource(contract.SourcePath, source)
	add("source-binding", sourceObservation.Digest == report.Source.Digest && reflect.DeepEqual(sourceObservation, report.Source) &&
		sourceObservation.Package == contract.SourcePackage && sourceObservation.Namespace == contract.SourceNamespace &&
		sourceObservation.Entities == contract.SourceEntities && sourceObservation.Activities == contract.SourceActivities, "source")
	add("receipt-identity", report.Schema == ReceiptSchema && report.ContractID == contract.ID && report.Producer == Producer &&
		report.Consumer == Consumer && report.MetaOperation == MetaOperation && report.ProofChoice == "FOUNDATION", "identity")
	add("fixed-denominator", report.Summary.FixedDenominator == Denominator && report.Summary.CoordinatesTotal == Denominator &&
		report.Summary.CasesTotal == CaseTotal && len(report.Indicators) == Denominator, "denominator")

	wantClaims := make([]transition, 0, len(contract.Cases))
	casesPass := len(report.Cases) == len(contract.Cases)
	for index, spec := range contract.Cases {
		wantDecision, wantResolution, wantReason := decisionFor(spec, contract.Budget)
		wantClaim := transition{CaseID: spec.ID, From: "AMBIGUITY_OBSERVED", To: wantResolution, Reason: wantReason}
		wantClaims = append(wantClaims, wantClaim)
		if !casesPass {
			continue
		}
		observed := report.Cases[index]
		casesPass = observed.ID == spec.ID && observed.Class == spec.Class && observed.InputState == spec.InputState &&
			observed.Counts == spec.Counts && observed.ExpectedDecision == wantDecision && observed.ExpectedResolution == wantResolution &&
			observed.ExpectedReason == wantReason && observed.Decision == wantDecision && observed.Resolution == wantResolution &&
			observed.Reason == wantReason && observed.Coordinate == spec.Coordinate && observed.Claim == wantClaim && observed.Status == "SATISFIED"
	}
	add("case-decisions", casesPass, "cases")
	add("claim-transitions", reflect.DeepEqual(report.Claims, wantClaims), "claims")

	indicatorsPass := len(report.Indicators) == Denominator
	for index, spec := range contract.Cases {
		values := []struct {
			id, dimension, proof string
			observed, expected   int
		}{
			{"gooo.metric.ambiguity-budget.candidate-count.v1", "interpretation_candidates", "FOUNDATION", spec.Counts.InterpretationCandidates, spec.Counts.InterpretationCandidates},
			{"gooo.metric.ambiguity-budget.unresolved-branches.v1", "unresolved_branches", "COHERENCE", spec.Counts.UnresolvedBranches, spec.Counts.UnresolvedBranches},
			{"gooo.metric.ambiguity-budget.evidence-paths.v1", "evidence_paths", "REGRESSION", spec.Counts.EvidencePaths, spec.Counts.EvidencePaths},
		}
		for offset, value := range values {
			position := index*3 + offset
			if position >= len(report.Indicators) {
				indicatorsPass = false
				continue
			}
			observed := report.Indicators[position]
			indicatorsPass = indicatorsPass && observed == (indicator{MetricID: value.id, CaseID: spec.ID, Dimension: value.dimension,
				Class: "DRIVER", ProofChoice: value.proof, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
				Observed: value.observed, Expected: value.expected, Satisfied: true})
		}
	}
	add("integer-set-indicators", indicatorsPass, "indicators")

	proofsPass := len(report.Proofs) == 3
	for index, choice := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		if index >= len(report.Proofs) {
			proofsPass = false
			continue
		}
		proof := report.Proofs[index]
		proofsPass = proofsPass && proof.Choice == choice && proof.Producer == Producer && proof.Consumer == Consumer &&
			proof.EvidenceDigest != "" && proof.Passed
	}
	add("proof-bindings", proofsPass, "proofs")

	wantSummary := summary{CasesSatisfied: CaseTotal, CasesTotal: CaseTotal, CoordinatesSatisfied: Denominator,
		CoordinatesTotal: Denominator, FixedDenominator: Denominator, ZeroAmbiguityCases: 1, BoundaryCases: 1,
		OverBudgetCases: 1, UnknownCases: 1, LowerResolutionCases: 2}
	summaryPass := report.Summary == wantSummary && report.NotClaimed != nil && reflect.DeepEqual(report.NotClaimed, contract.NotClaimed)
	add("summary-and-effects", summaryPass && report.Effects == (effects{}), "summary")

	digestPass := report.Digest == digestReport(report)
	add("receipt-digest", digestPass, "digest")

	result.Decision, result.Resolution, result.Reason = "PASS", "EXACT", "AMBIGUITY_RECEIPT_ACCEPTED"
	if !allChecksPass(result.Checks) {
		result.Decision, result.Resolution, result.Reason = "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_RECEIPT_REJECTED"
	}
	resultDigest := resultDigest(result)
	result.Digest = resultDigest
	return result, nil
}

func decisionFor(spec caseSpec, budget integerSet) (string, string, string) {
	if spec.InputState == "UNKNOWN" {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_INPUT_UNKNOWN"
	}
	if spec.Counts.InterpretationCandidates < 1 || spec.Counts.UnresolvedBranches < 0 || spec.Counts.EvidencePaths < 1 {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COUNT_UNKNOWN"
	}
	if spec.Counts.InterpretationCandidates > budget.InterpretationCandidates || spec.Counts.UnresolvedBranches > budget.UnresolvedBranches || spec.Counts.EvidencePaths > budget.EvidencePaths {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED"
	}
	return "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"
}

func observeSource(path string, raw []byte) sourceObservation {
	observation := sourceObservation{Path: path, Digest: digestBytes(raw)}
	for _, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		fields := strings.Fields(line)
		switch {
		case len(fields) == 2 && fields[0] == "package":
			observation.Package = fields[1]
		case len(fields) == 2 && fields[0] == "namespace":
			observation.Namespace = fields[1]
		case len(fields) == 4 && fields[0] == "entity" && fields[2] == "id":
			observation.Entities++
		case strings.HasPrefix(line, "activity "):
			observation.Activities++
		}
	}
	return observation
}

func allChecksPass(checks []Check) bool {
	for _, check := range checks {
		if check.Status != "PASS" {
			return false
		}
	}
	return true
}

func decode(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestReport(report receipt) string {
	report.Digest = ""
	return digestValue(report)
}

func resultDigest(result Result) string {
	result.Digest = ""
	return digestValue(result)
}

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}
