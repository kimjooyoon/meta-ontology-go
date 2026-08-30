package governancesnapshot

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"
)

type branchPayload struct {
	Name      string `json:"name"`
	Protected *bool  `json:"protected"`
	Commit    struct {
		SHA string `json:"sha"`
	} `json:"commit"`
	Protection *protectionPayload `json:"protection"`
}

type protectionPayload struct {
	RequiredStatusChecks *struct {
		EnforcementLevel string `json:"enforcement_level"`
		Contexts []string `json:"contexts"`
		Checks   []struct {
			Context string `json:"context"`
		} `json:"checks"`
	} `json:"required_status_checks"`
}

type repositoryPayload struct {
	DefaultBranch *string `json:"default_branch"`
}

type rulesetPayload struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Target      string `json:"target"`
	Enforcement string `json:"enforcement"`
	Authority   bool   `json:"authority"`
}

type parsedSnapshot struct {
	DefaultBranch string
	Branches      map[string]BranchEvidence
	Rulesets      []RulesetEvidence
	SourceReason  string
	SourceUnknown *Unknown
}

func Evaluate(input LoadedSnapshot, contract Contract, graph RawGraph) Report {
	report := evaluateCore(input, contract)
	graphValue, graphReason := graphEvidence(graph, contract)
	report.Graph = graphValue
	if graphReason != "" {
		report.Cells = appendGraphRefutation(report.Cells, contract, graphReason)
	}
	report.Cases = canonicalCases(input, contract)
	report.Replay = ReplayEvidence{InputDigest: digestJSON(input.Snapshot), ProjectionDigest: digestJSON(report), ReplayEqual: true}
	finishReport(&report, contract)
	return report
}

func evaluateCore(input LoadedSnapshot, contract Contract) Report {
	parsed := parseSnapshot(input, contract)
	report := Report{Schema: ReportSchema, ContractID: contract.ID, Repository: input.Repository,
		HeadSHA: input.HeadSHA, DefaultBranch: parsed.DefaultBranch, Source: sourceEvidence(input, contract),
		RepositoryWrites: input.RepositoryWrites, BranchSettingWrites: input.BranchSettingWrites,
		LocalTestExecutions: input.LocalTestExecutions, CrossProjectGates: input.CrossProjectGates,
		Improvement: "UNKNOWN", PromotionAuthorized: false,
		Branches: orderedBranches(parsed.Branches),
		Rulesets: append([]RulesetEvidence(nil), parsed.Rulesets...)}
	report.Cells = evaluateCells(report, parsed, contract)
	return report
}

func sourceEvidence(input LoadedSnapshot, contract Contract) SourceEvidence {
	evidence := SourceEvidence{Documentation: append([]string(nil), contract.Source.Documentation...),
		APIVersions: cloneMap(contract.Source.APIVersions), PayloadDigestModel: contract.Source.PayloadDigestModel,
		Requests: append([]RequestObservation(nil), input.Requests...)}
	for _, request := range input.Requests {
		evidence.Payloads = append(evidence.Payloads, PayloadEvidence{ID: request.ID, State: request.State, Digest: request.PayloadDigest})
	}
	return evidence
}

func parseSnapshot(input LoadedSnapshot, contract Contract) parsedSnapshot {
	parsed := parsedSnapshot{Branches: map[string]BranchEvidence{}}
	if issue := validateRequests(input, contract); issue != "" {
		parsed.SourceReason = issue
	}
	parsed.DefaultBranch = parseRepository(input, &parsed)
	for _, branch := range []string{"dev", "main"} {
		observed, branchReason := parseBranch(input, branch, &parsed)
		if branchReason != "" && parsed.SourceReason == "" {
			parsed.SourceReason = branchReason
		}
		parsed.Branches[branch] = observed
	}
	rulesets, rulesetReason := parseRulesets(input, &parsed)
	parsed.Rulesets = rulesets
	if rulesetReason != "" && parsed.SourceReason == "" {
		parsed.SourceReason = rulesetReason
	}
	return parsed
}

func parseRepository(input LoadedSnapshot, parsed *parsedSnapshot) string {
	raw, state := payload(input, "repository")
	if state != "PRESENT" {
		parsed.SourceUnknown = directMissing("PUBLIC_SNAPSHOT_MISSING")
		return ""
	}
	var value repositoryPayload
	if err := json.Unmarshal(raw, &value); err != nil {
		parsed.SourceReason = "MALFORMED_PUBLIC_PAYLOAD"
		return ""
	}
	if value.DefaultBranch == nil || *value.DefaultBranch == "" {
		parsed.SourceUnknown = directMissing("PUBLIC_SNAPSHOT_FIELD_MISSING")
		return ""
	}
	return *value.DefaultBranch
}

func parseBranch(input LoadedSnapshot, branch string, parsed *parsedSnapshot) (BranchEvidence, string) {
	raw, state := payload(input, branch+"-branch")
	if state != "PRESENT" {
		parsed.SourceUnknown = directMissing("PUBLIC_SNAPSHOT_MISSING")
		return BranchEvidence{Branch: branch}, ""
	}
	var value branchPayload
	if err := json.Unmarshal(raw, &value); err != nil {
		return BranchEvidence{Branch: branch}, "MALFORMED_PUBLIC_PAYLOAD"
	}
	if value.Protected == nil || value.Name == "" || value.Commit.SHA == "" {
		parsed.SourceUnknown = directMissing("PUBLIC_SNAPSHOT_FIELD_MISSING")
		return BranchEvidence{Branch: branch, CommitSHA: value.Commit.SHA}, ""
	}
	protected := *value.Protected
	status, statusReason, statusSource := parseBranchStatus(input, branch, protected, value.Protection, parsed)
	if statusReason != "" {
		return BranchEvidence{Branch: branch, CommitSHA: value.Commit.SHA, Protected: protected}, statusReason
	}
	return BranchEvidence{Branch: branch, CommitSHA: value.Commit.SHA, Available: true, Protected: protected, StatusSource: statusSource,
		Enforcement: status.enforcement, Contexts: status.contexts}, ""
}

type protectionObservation struct {
	enforcement string
	contexts    []string
}

func parseBranchStatus(input LoadedSnapshot, branch string, protected bool, inline *protectionPayload, parsed *parsedSnapshot) (protectionObservation, string, string) {
	if inline == nil {
		parsed.SourceUnknown = dependencyBlocked(branch)
		return protectionObservation{}, "", "branch-summary"
	}
	status, reason := parseProtectionValue(protected, inline)
	return status, reason, "branch-summary"
}

func parseProtectionValue(protected bool, value *protectionPayload) (protectionObservation, string) {
	if !protected {
		return protectionObservation{enforcement: "off"}, ""
	}
	if value == nil || value.RequiredStatusChecks == nil {
		return protectionObservation{enforcement: "off", contexts: []string{}}, ""
	}
	enforcement := value.RequiredStatusChecks.EnforcementLevel
	if enforcement == "" {
		return protectionObservation{}, "MALFORMED_PUBLIC_PAYLOAD"
	}
	contexts := append([]string{}, value.RequiredStatusChecks.Contexts...)
	if hasDuplicate(contexts) {
		return protectionObservation{}, "DUPLICATE_REQUIRED_CONTEXT"
	}
	sort.Strings(contexts)
	return protectionObservation{enforcement: enforcement, contexts: contexts}, ""
}

func parseRulesets(input LoadedSnapshot, parsed *parsedSnapshot) ([]RulesetEvidence, string) {
	raw, state := payload(input, "rulesets")
	if state != "PRESENT" {
		parsed.SourceUnknown = directMissing("PUBLIC_SNAPSHOT_MISSING")
		return nil, ""
	}
	var values []rulesetPayload
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, "MALFORMED_PUBLIC_PAYLOAD"
	}
	result := make([]RulesetEvidence, 0, len(values))
	seen := map[int64]bool{}
	for _, value := range values {
		if value.ID <= 0 || value.Name == "" || value.Target == "" || value.Enforcement == "" || seen[value.ID] {
			return nil, "RULESET_PAYLOAD_CONTRADICTION"
		}
		if value.Authority && value.Enforcement == "disabled" {
			return nil, "DISABLED_RULESET_AUTHORITY"
		}
		seen[value.ID] = true
		result = append(result, RulesetEvidence{ID: value.ID, Name: value.Name, Target: value.Target, Enforcement: value.Enforcement})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, ""
}

func payload(input LoadedSnapshot, id string) ([]byte, string) {
	for _, request := range input.Requests {
		if request.ID == id {
			return input.Payloads[id], request.State
		}
	}
	return nil, "MISSING"
}

func orderedBranches(branches map[string]BranchEvidence) []BranchEvidence {
	return []BranchEvidence{branches["dev"], branches["main"]}
}

func cloneMap(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	maps.Copy(result, values)
	return result
}

func hasDuplicate(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func directMissing(reason string) *Unknown {
	return &Unknown{Stage: "capture", Step: "read-public-rest-snapshot", Reason: reason,
		UnknownClass: "DIRECT_MISSING", NextOperation: "CAPTURE_GITHUB_PUBLIC_SNAPSHOT", BlockedBy: []string{}}
}

func dependencyBlocked(branch string) *Unknown {
	return &Unknown{Stage: "evaluate-governance", Step: "compare-required-contexts", Reason: "BRANCH_PUBLIC_PROTECTION_FIELDS_UNAVAILABLE",
		UnknownClass: "DEPENDENCY_BLOCKED", NextOperation: "CAPTURE_BRANCH_PUBLIC_SNAPSHOT",
		BlockedBy: []string{"live-governance-snapshot:" + branch + "-branch"}}
}

func formatContexts(values []string) string {
	data, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func formatRulesets(values []RulesetEvidence) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%s#%d:%s", value.Name, value.ID, value.Enforcement))
	}
	return strings.Join(parts, ",")
}
