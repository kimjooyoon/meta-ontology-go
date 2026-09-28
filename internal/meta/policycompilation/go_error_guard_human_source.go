package policycompilation

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func proposeGoHumanOutputGuard(report GoErrorGuardProposal, profile goErrorGuardProgram, source []byte, set *token.FileSet, file *ast.File) (GoErrorGuardProposal, error) {
	matches, functions := findGoHumanOutputGuards(file, profile)
	if functions > 1 || len(matches) > 1 {
		return declineGoErrorGuard(report, "UNKNOWN", "GO_GUARD_TARGET_AMBIGUOUS", "AMBIGUOUS", "SELECT_UNAMBIGUOUS_SOURCE")
	}
	if functions == 0 {
		return declineGoErrorGuard(report, "UNKNOWN", "GO_GUARD_FUNCTION_MISSING", "DIRECT_MISSING", "SUPPLY_DECLARED_FUNCTION")
	}
	if len(matches) != 1 {
		return declineGoErrorGuard(report, "UNKNOWN", "GO_HUMAN_GUARD_SHAPE_UNSUPPORTED", "UNBOUNDED", "SUPPLY_DECLARED_HUMAN_BRANCH")
	}
	match := matches[0]
	if len(match.calls) != profile.writeCount {
		return declineGoErrorGuard(report, "REFUTED", "GO_HUMAN_GUARD_WRITE_COUNT_MISMATCH", "", "")
	}
	body := source[set.Position(match.handler.Pos()).Offset:set.Position(match.handler.End()).Offset]
	report.HandlerDigest = DigestBytes(body)
	if report.HandlerDigest != profile.handler {
		return declineGoErrorGuard(report, "REFUTED", "GO_HANDLER_PIN_MISMATCH", "", "")
	}
	report.EditStart = set.Position(match.calls[0].Pos()).Offset
	report.EditEnd = set.Position(match.calls[len(match.calls)-1].End()).Offset
	var candidate strings.Builder
	cursor := 0
	for _, call := range match.calls {
		start, end := set.Position(call.Pos()).Offset, set.Position(call.End()).Offset
		candidate.Write(source[cursor:start])
		candidate.WriteString(fmt.Sprintf("if _, %s := %s; %s != nil %s", match.errorName, source[start:end], match.errorName, body))
		cursor = end
	}
	candidate.Write(source[cursor:])
	report.CandidateSource = candidate.String()
	report.CandidateDigest = DigestBytes([]byte(report.CandidateSource))
	report.GuardedCalls = len(match.calls)
	report.State, report.Reason = "PROPOSED", "GOOO_BOUND_HUMAN_OUTPUT_ERROR_GUARDS"
	return report, nil
}
