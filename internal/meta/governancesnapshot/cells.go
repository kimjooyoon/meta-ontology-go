package governancesnapshot

import "fmt"

func evaluateCells(report Report, parsed parsedSnapshot, contract Contract) []CellObservation {
	byID := map[string]CellSpec{}
	for _, spec := range contract.Cells {
		byID[spec.ID] = spec
	}
	cells := make([]CellObservation, 0, len(contract.Cells))
	for _, spec := range contract.Cells {
		cell := baseCell(spec)
		switch spec.ID {
		case "OBSERVATION_SOURCE_PIN":
			cell = sourcePinCell(cell, parsed)
		case "CONTRACT_MANIFEST_PIN":
			cell = closeCell(cell, "contract="+contract.ID, "checked-in contract")
		case "DEV_BRANCH_PROTECTED", "MAIN_BRANCH_PROTECTED":
			cell = branchProtectedCell(cell, parsed.Branches[branchForCell(spec.ID)])
		case "DEV_STATUS_ENFORCEMENT", "MAIN_STATUS_ENFORCEMENT":
			branch := branchForCell(spec.ID)
			cell = compareStatusCell(cell, parsed, branch, expectedBranch(contract, branch))
		case "DEV_CONTEXT_SET", "MAIN_CONTEXT_SET":
			branch := branchForCell(spec.ID)
			cell = compareContextsCell(cell, parsed, branch, expectedBranch(contract, branch))
		case "RULESET_INVENTORY":
			cell = rulesetInventoryCell(cell, parsed, contract)
		case "DISABLED_RULESET_AUTHORITY":
			cell = disabledAuthorityCell(cell, parsed, contract)
		case "UNKNOWN_CAUSALITY":
			cell = unknownCausalityCell(cell, parsed, cells)
		case "HUMAN_DRIFT_REPORT":
			cell = closeCell(cell, "published=governance-drift-report", "human-readable drift report")
		}
		cell.EvidenceDigest = digestJSON(struct {
			ID, Observed, Expected, Decision, Resolution, Reason string
		}{cell.ID, cell.Observed, cell.Expected, cell.Decision, cell.Resolution, cell.Reason})
		cells = append(cells, cell)
	}
	return cells
}

func baseCell(spec CellSpec) CellObservation {
	return CellObservation{ID: spec.ID, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
		Indicator: spec.Indicator, Activity: spec.Activity, InputID: spec.InputID, OutputID: spec.OutputID}
}

func sourcePinCell(cell CellObservation, parsed parsedSnapshot) CellObservation {
	if parsed.SourceReason != "" {
		return refutedCell(cell, "source-payload:"+parsed.SourceReason, parsed.SourceReason, "valid public REST payload")
	}
	if parsed.SourceUnknown != nil {
		return unknownCell(cell, parsed.SourceUnknown, "public snapshot unavailable", "complete public REST snapshot")
	}
	return closeCell(cell, "requests=6;normalized-digests=checked", "six exact public REST requests")
}

func branchProtectedCell(cell CellObservation, branch BranchEvidence) CellObservation {
	cell.Observed = fmt.Sprintf("%s protected=%t", branch.Branch, branch.Protected)
	cell.Expected = "protected=true"
	if branch.Branch == "" {
		return unknownCell(cell, directMissing("PUBLIC_SNAPSHOT_FIELD_MISSING"), cell.Observed, cell.Expected)
	}
	if !branch.Protected {
		return refutedCell(cell, branch.Branch+":protected=false", "BRANCH_NOT_PROTECTED", cell.Expected)
	}
	return closeCell(cell, cell.Observed, cell.Expected)
}

func compareStatusCell(cell CellObservation, parsed parsedSnapshot, branch string, expected BranchExpectation) CellObservation {
	observed := parsed.Branches[branch].Enforcement
	cell.Observed, cell.Expected = observed, expected.Enforcement
	if observed == "" {
		return unknownCell(cell, parsed.SourceUnknown, "status enforcement unavailable", cell.Expected)
	}
	if observed != expected.Enforcement {
		return refutedCell(cell, branch+":status-enforcement", "STATUS_ENFORCEMENT_MISMATCH", cell.Expected)
	}
	return closeCell(cell, cell.Observed, cell.Expected)
}

func compareContextsCell(cell CellObservation, parsed parsedSnapshot, branch string, expected BranchExpectation) CellObservation {
	observed := parsed.Branches[branch].Contexts
	cell.Observed, cell.Expected = formatContexts(observed), formatContexts(expected.Contexts)
	if parsed.Branches[branch].Enforcement == "" {
		return unknownCell(cell, parsed.SourceUnknown, "required contexts unavailable", cell.Expected)
	}
	if !sameStrings(observed, expected.Contexts) {
		return refutedCell(cell, branch+":required-contexts", "REQUIRED_CONTEXT_SET_MISMATCH", cell.Expected)
	}
	return closeCell(cell, cell.Observed, cell.Expected)
}

func rulesetInventoryCell(cell CellObservation, parsed parsedSnapshot, contract Contract) CellObservation {
	cell.Observed = formatRulesets(parsed.Rulesets)
	cell.Expected = expectedRulesets(contract)
	if parsed.Rulesets == nil {
		return unknownCell(cell, parsed.SourceUnknown, "rulesets unavailable", cell.Expected)
	}
	if len(parsed.Rulesets) != len(contract.Expected.Rulesets) {
		return refutedCell(cell, "rulesets:count", "RULESET_INVENTORY_MISMATCH", cell.Expected)
	}
	for _, expected := range contract.Expected.Rulesets {
		if !containsRuleset(parsed.Rulesets, expected) {
			return refutedCell(cell, "rulesets:"+expected.Branch, "RULESET_INVENTORY_MISMATCH", cell.Expected)
		}
	}
	return closeCell(cell, cell.Observed, cell.Expected)
}

func disabledAuthorityCell(cell CellObservation, parsed parsedSnapshot, contract Contract) CellObservation {
	cell.Observed = formatRulesets(parsed.Rulesets)
	cell.Expected = "all expected rulesets enforcement=disabled; disabled is not authority"
	if parsed.Rulesets == nil {
		return unknownCell(cell, parsed.SourceUnknown, "ruleset authority unavailable", cell.Expected)
	}
	for _, value := range parsed.Rulesets {
		if value.Enforcement != contract.Expected.RequiredRulesetState {
			return refutedCell(cell, "ruleset:"+value.Name, "DISABLED_RULESET_AUTHORITY", cell.Expected)
		}
	}
	return closeCell(cell, cell.Observed, cell.Expected)
}

func unknownCausalityCell(cell CellObservation, parsed parsedSnapshot, prior []CellObservation) CellObservation {
	cell.Observed, cell.Expected = "typed-unknown-frontier", "all unavailable observations retain six fields"
	if parsed.SourceUnknown != nil || hasUnknown(prior) {
		return unknownCell(cell, firstUnknown(parsed.SourceUnknown, prior), cell.Observed, cell.Expected)
	}
	return closeCell(cell, cell.Observed, cell.Expected)
}

func firstUnknown(source *Unknown, cells []CellObservation) *Unknown {
	if source != nil {
		return source
	}
	for _, cell := range cells {
		if cell.Unknown != nil {
			return cell.Unknown
		}
	}
	return directMissing("PUBLIC_SNAPSHOT_FIELD_MISSING")
}

func expectedBranch(contract Contract, branch string) BranchExpectation {
	for _, expected := range contract.Expected.StatusChecks {
		if expected.Branch == branch {
			return expected
		}
	}
	return BranchExpectation{Branch: branch, Enforcement: "everyone", Contexts: ExpectedContexts(), Protected: true}
}

func expectedRulesets(contract Contract) string {
	values := make([]RulesetEvidence, 0, len(contract.Expected.Rulesets))
	for _, expected := range contract.Expected.Rulesets {
		values = append(values, RulesetEvidence{ID: expected.ID, Name: expected.Name, Enforcement: expected.Enforcement})
	}
	return formatRulesets(values)
}

func containsRuleset(values []RulesetEvidence, expected RulesetExpectation) bool {
	for _, value := range values {
		if value.ID == expected.ID && value.Name == expected.Name && value.Enforcement == expected.Enforcement {
			return true
		}
	}
	return false
}

func branchForCell(id string) string {
	if len(id) >= 4 && id[:4] == "DEV_" {
		return "dev"
	}
	return "main"
}

func hasUnknown(cells []CellObservation) bool {
	for _, cell := range cells {
		if cell.Decision == DecisionUnknown {
			return true
		}
	}
	return false
}

func closeCell(cell CellObservation, observed, expected string) CellObservation {
	cell.Observed, cell.Expected, cell.Decision, cell.Resolution = observed, expected, DecisionClosed, ResolutionExact
	return cell
}

func refutedCell(cell CellObservation, counterexample, reason, expected string) CellObservation {
	cell.Decision, cell.Resolution, cell.Reason = DecisionRefuted, ResolutionExact, reason
	cell.Counterexample, cell.Expected = counterexample, expected
	return cell
}

func unknownCell(cell CellObservation, unknown *Unknown, observed, expected string) CellObservation {
	cell.Observed, cell.Expected, cell.Decision, cell.Resolution = observed, expected, DecisionUnknown, ResolutionLower
	if unknown == nil {
		unknown = directMissing("PUBLIC_SNAPSHOT_FIELD_MISSING")
	}
	cell.Unknown = unknown
	cell.Reason = unknown.Reason
	return cell
}
