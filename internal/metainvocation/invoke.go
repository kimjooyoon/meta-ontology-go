package metainvocation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

type rule struct {
	operation string
	checkID   string
	command   string
	activity  string
}

var rules = []rule{
	{operation: operationDocsRule, checkID: "docs-check", command: "gooo check docs", activity: "SelectDocsCheck"},
	{operation: operationGoRule, checkID: "go-test", command: "go test ./...", activity: "SelectGoCheck"},
	{operation: operationYAMLRule, checkID: "yaml-check", command: "gooo check yaml", activity: "SelectYAMLCheck"},
}

func Invoke(program Program, entry string, rawInput []byte) (Report, error) {
	if err := validateProgram(program); err != nil {
		return Report{}, err
	}
	bound, ok := program.Operations[entry]
	if !ok || bound.Program != operationPlan {
		return Report{}, fmt.Errorf("entry %q is not bound to %q", entry, operationPlan)
	}
	if program.Operations["VerifyCIPlan"].Program != operationVerify {
		return Report{}, fmt.Errorf("verification operation is not bound")
	}
	changeSet, decodeReason := decodeChangeSet(rawInput)
	inputDigest := bytesDigest(rawInput)
	if decodeReason != "" {
		return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionClosed, ResolutionExact, nil, nil, decodeReason), nil
	}
	if reason, file := validateChangeSet(changeSet); reason != "" {
		return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionClosed, ResolutionExact, nil, nil, reasonWithFile(reason, file)), nil
	}
	checks, unknowns := selectChecks(program, changeSet)
	if len(unknowns) != 0 {
		return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionUnknown, ResolutionLower, nil, unknowns, ""), nil
	}
	return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionPass, ResolutionExact, checks, nil, ""), nil
}

func decodeChangeSet(raw []byte) (ChangeSet, string) {
	changeSet := ChangeSet{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&changeSet); err != nil {
		return changeSet, "INPUT_DECODE:decode-change-set:" + err.Error()
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return changeSet, "INPUT_DECODE:reject-trailing-content"
	}
	return changeSet, ""
}

func validateChangeSet(changeSet ChangeSet) (string, string) {
	if changeSet.Schema != InputSchema {
		return "INPUT_SCHEMA_MISMATCH", ""
	}
	if changeSet.CaseID == "" {
		return "CASE_ID_EMPTY", ""
	}
	if len(changeSet.Files) == 0 {
		return "CHANGE_SET_EMPTY", ""
	}
	seen := map[string]struct{}{}
	for _, file := range changeSet.Files {
		if file == "" {
			return "CHANGE_PATH_EMPTY", file
		}
		if path.IsAbs(file) || strings.HasPrefix(file, "\\") {
			return "CHANGE_PATH_ABSOLUTE", file
		}
		if strings.Contains(file, "\\") || path.Clean(file) != file || file == "." || strings.HasPrefix(file, "../") {
			return "CHANGE_PATH_NON_CANONICAL", file
		}
		if _, exists := seen[file]; exists {
			return "CHANGE_PATH_DUPLICATE", file
		}
		seen[file] = struct{}{}
	}
	return "", ""
}

func reasonWithFile(reason, file string) string {
	if file == "" {
		return reason
	}
	return reason + ":" + file
}

func selectChecks(program Program, changeSet ChangeSet) ([]PlannedCheck, []UnknownCause) {
	filesByRule := map[string][]string{}
	unknowns := []UnknownCause{}
	for _, file := range changeSet.Files {
		matched := false
		for _, candidate := range rules {
			if ruleMatches(candidate.operation, file) {
				filesByRule[candidate.operation] = append(filesByRule[candidate.operation], file)
				matched = true
			}
		}
		if !matched {
			unknowns = append(unknowns, UnknownCause{Stage: "RULE_SELECTION", Step: "classify-change", Reason: "NO_REGISTERED_RULE", File: file})
		}
	}
	if len(unknowns) != 0 {
		return nil, unknowns
	}
	checks := make([]PlannedCheck, 0, len(filesByRule))
	for _, candidate := range rules {
		files := filesByRule[candidate.operation]
		if len(files) == 0 {
			continue
		}
		sort.Strings(files)
		bound := program.Operations[candidate.activity]
		reasons := make([]RuleEvidence, 0, len(files))
		for _, file := range files {
			reasons = append(reasons, RuleEvidence{
				ID: candidate.operation + ":" + file, Operation: candidate.operation, File: file,
				SpecDigest: bound.SpecDigest, Source: bound.Source,
			})
		}
		checks = append(checks, PlannedCheck{ID: candidate.checkID, Command: candidate.command, Files: files, Reasons: reasons})
	}
	return checks, nil
}

func ruleMatches(operation, file string) bool {
	switch operation {
	case operationDocsRule:
		return strings.HasPrefix(file, "docs/")
	case operationGoRule:
		return path.Ext(file) == ".go"
	case operationYAMLRule:
		extension := path.Ext(file)
		return extension == ".yaml" || extension == ".yml"
	default:
		return false
	}
}

func buildReport(program Program, entry, caseID, inputDigest, decision, resolution string, checks []PlannedCheck, unknowns []UnknownCause, failureReason string) Report {
	if checks == nil {
		checks = []PlannedCheck{}
	}
	if unknowns == nil {
		unknowns = []UnknownCause{}
	}
	plan := sealPlan(CheckPlan{Schema: PlanSchema, CaseID: caseID, InputDigest: inputDigest, Checks: checks})
	evidenceDigests := []string{program.SourceDigest}
	for _, check := range checks {
		for _, reason := range check.Reasons {
			evidenceDigests = append(evidenceDigests, digest(reason))
		}
	}
	receipt := sealReceipt(VerificationReceipt{
		Schema: ReceiptSchema, SubjectDigest: plan.Digest, Decision: decision, Resolution: resolution,
		EvidenceDigests: evidenceDigests, Unknowns: unknowns,
	})
	report := Report{
		Schema: ReportSchema, Decision: decision, Resolution: resolution, CaseID: caseID, Entry: entry,
		SourceDigest: program.SourceDigest, InputDigest: inputDigest, Plan: plan, Receipt: receipt,
		Unknowns: unknowns, Claims: claimsFor(decision, program.SourceDigest, checks, unknowns, failureReason),
		Effects: Effects{}, NotClaimed: []string{
			"check-execution", "full-language-semantic-correctness", "general-build-planning",
			"comparative-performance", "production-or-external-effects",
		},
	}
	return sealReport(report)
}

func claimsFor(decision, sourceDigest string, checks []PlannedCheck, unknowns []UnknownCause, failureReason string) []Claim {
	sourceClaim := Claim{
		ID: "source-program-binding", Statement: "the invoked entry is bound to the declared Gooo meta program",
		Status: ClaimDischarged, Stage: "SOURCE_BINDING", Step: "compile-meta-program", Reason: "SOURCE_BINDING_PROVED",
		Evidence: []string{sourceDigest}, DependsOn: []string{},
	}
	ruleClaim := Claim{
		ID: "rule-evidence-completeness", Statement: "every changed path has registered rule evidence",
		Status: ClaimDischarged, Stage: "RULE_SELECTION", Step: "classify-change", Reason: "RULE_EVIDENCE_COMPLETE",
		Evidence: []string{}, DependsOn: []string{"source-program-binding"},
	}
	for _, check := range checks {
		for _, reason := range check.Reasons {
			ruleClaim.Evidence = append(ruleClaim.Evidence, reason.ID)
		}
	}
	planClaim := Claim{
		ID: "ci-plan-decision", Statement: "the selected checks are authorized by complete rule evidence",
		Status: ClaimDischarged, Stage: "PLAN_AUTHORIZATION", Step: "require-rule-evidence", Reason: "PLAN_AUTHORIZED",
		Evidence: []string{}, DependsOn: []string{"rule-evidence-completeness"},
	}
	if decision == DecisionPass {
		for _, check := range checks {
			planClaim.Evidence = append(planClaim.Evidence, check.ID)
		}
	}
	if decision == DecisionClosed {
		ruleClaim.Status = ClaimRefuted
		ruleClaim.Stage = "INPUT_VALIDATION"
		ruleClaim.Step = "validate-change-set"
		ruleClaim.Reason = failureReason
		planClaim.Status = ClaimOpen
		planClaim.Reason = "DEPENDENCY_BLOCKED"
	}
	if decision == DecisionUnknown {
		ruleClaim.Status = ClaimOpen
		ruleClaim.Reason = "UNKNOWN_AT_RULE_SELECTION"
		for _, unknown := range unknowns {
			ruleClaim.Evidence = append(ruleClaim.Evidence, unknown.Stage+":"+unknown.Step+":"+unknown.Reason+":"+unknown.File)
		}
		planClaim.Status = ClaimOpen
		planClaim.Reason = "DEPENDENCY_BLOCKED"
	}
	return []Claim{sourceClaim, ruleClaim, planClaim}
}
