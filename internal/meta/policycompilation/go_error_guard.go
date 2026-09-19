package policycompilation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type GoErrorGuardProposal struct {
	State              string                 `json:"state"`
	Reason             string                 `json:"reason"`
	ActivityID         string                 `json:"activity_id"`
	ProgramDigest      string                 `json:"program_digest"`
	SemanticDigest     string                 `json:"semantic_digest"`
	SourceDigest       string                 `json:"source_digest"`
	HandlerDigest      string                 `json:"handler_digest"`
	CandidateDigest    string                 `json:"candidate_digest,omitempty"`
	CandidateSource    string                 `json:"candidate_source,omitempty"`
	EditStart          int                    `json:"edit_start"`
	EditEnd            int                    `json:"edit_end"`
	GuardedCalls       int                    `json:"guarded_calls,omitempty"`
	Pending            *PolicyRevisionPending `json:"pending,omitempty"`
	Admission          PolicyRevisionPending  `json:"admission"`
	Improvement        string                 `json:"improvement"`
	MutationAuthority  int                    `json:"mutation_authority"`
	PromotionAuthority int                    `json:"promotion_authority"`
}

type goErrorGuardProgram struct {
	function    string
	writer      string
	diagnostic  string
	source      string
	handler     string
	writerType  string
	returnOnly  bool
	humanBranch bool
	mode        string
	handlerCall string
	writeCount  int
}

// ProposeGoErrorGuard implements an opt-in computes profile, not a policy
// decision revision or a semantics-preserving extraction. It never executes or applies
// the candidate. Edit offsets are zero-based byte offsets in the pinned source.
func ProposeGoErrorGuard(filename string, program, source []byte) (GoErrorGuardProposal, error) {
	report := GoErrorGuardProposal{
		State: "UNKNOWN", ProgramDigest: DigestBytes(program), SourceDigest: DigestBytes(source),
		Admission: revisionPending("INDEPENDENT_VALIDATION", "OBSERVE_GO_GUARD_CANDIDATE",
			"NATIVE_GO_GUARD_EVIDENCE_MISSING", "RUN_FROZEN_NATIVE_ORACLE"),
		Improvement: "UNKNOWN",
	}
	profile, activity, semanticDigest, err := compileGoErrorGuard(filename, program)
	if err != nil {
		report.State, report.Reason = "REFUTED", "GOOO_GUARD_PROFILE_INVALID"
		return report, fmt.Errorf("%s: %w", report.Reason, err)
	}
	report.ActivityID, report.SemanticDigest = activity, semanticDigest
	return proposeGoErrorGuardSource(report, profile, source)
}

func proposeGoErrorGuardSource(report GoErrorGuardProposal, profile goErrorGuardProgram, source []byte) (GoErrorGuardProposal, error) {
	if profile.source != report.SourceDigest {
		return declineGoErrorGuard(report, "REFUTED", "GO_SOURCE_PIN_MISMATCH", "", "")
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "source.go", source, parser.ParseComments)
	if err != nil {
		report.State, report.Reason = "REFUTED", "GO_SOURCE_SYNTAX_INVALID"
		return report, fmt.Errorf("%s: %w", report.Reason, err)
	}
	if profile.humanBranch {
		return proposeGoHumanOutputGuard(report, profile, source, set, file)
	}
	matches, functions := findGoErrorGuards(file, profile)
	if functions > 1 || len(matches) > 1 {
		return declineGoErrorGuard(report, "UNKNOWN", "GO_GUARD_TARGET_AMBIGUOUS", "AMBIGUOUS", "SELECT_UNAMBIGUOUS_SOURCE")
	}
	if functions == 0 {
		return declineGoErrorGuard(report, "UNKNOWN", "GO_GUARD_FUNCTION_MISSING", "DIRECT_MISSING", "SUPPLY_DECLARED_FUNCTION")
	}
	if len(matches) != 1 {
		return declineGoErrorGuard(report, "UNKNOWN", "GO_GUARD_SHAPE_UNSUPPORTED", "UNBOUNDED", "SUPPLY_SUPPORTED_GUARD_SHAPE")
	}
	match := matches[0]
	body := source[set.Position(match.handler.Pos()).Offset:set.Position(match.handler.End()).Offset]
	report.HandlerDigest = DigestBytes(body)
	if report.HandlerDigest != profile.handler {
		return declineGoErrorGuard(report, "REFUTED", "GO_HANDLER_PIN_MISMATCH", "", "")
	}
	report.EditStart, report.EditEnd = set.Position(match.call.Pos()).Offset, set.Position(match.call.End()).Offset
	call := source[report.EditStart:report.EditEnd]
	replacement := fmt.Sprintf("if _, %s := %s; %s != nil %s", match.errorName, call, match.errorName, body)
	report.CandidateSource = string(source[:report.EditStart]) + replacement + string(source[report.EditEnd:])
	report.CandidateDigest = DigestBytes([]byte(report.CandidateSource))
	report.State, report.Reason = "PROPOSED", "GOOO_BOUND_OUTPUT_ERROR_GUARD"
	return report, nil
}

func compileGoErrorGuard(filename string, source []byte) (goErrorGuardProgram, string, string, error) {
	ir, file, err := lowerPolicy(filename, source, "goerrorguard", "goerrorguard")
	if err != nil {
		return goErrorGuardProgram{}, "", "", err
	}
	if len(file.Decls) != 3 || len(file.Bindings) != 0 || len(ir.RuntimeBindings) != 0 ||
		len(ir.Policies) != 0 || len(ir.Graph.Nodes()) != 3 {
		return goErrorGuardProgram{}, "", "", fmt.Errorf("guard requires two entities and one activity")
	}
	input, inputOK := ir.Graph.NodeByName(ir.Namespace, "Source")
	output, outputOK := ir.Graph.NodeByName(ir.Namespace, "Candidate")
	if !inputOK || !outputOK || input.Kind != semantic.Entity || output.Kind != semantic.Entity ||
		len(input.Fields) != 0 || len(output.Fields) != 0 {
		return goErrorGuardProgram{}, "", "", fmt.Errorf("guard Source/Candidate entities differ")
	}
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, found := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !found || node.Kind != semantic.Activity || len(activity.Inputs) != 1 ||
			activity.Inputs[0].Name != "Source" || activity.Output != "Candidate" ||
			!activity.ValueProgramPresent || activity.ValueProgram != node.ValueProgram ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: input.ID}) ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID}) {
			return goErrorGuardProgram{}, "", "", fmt.Errorf("guard activity provenance differs")
		}
		profile, err := parseGoErrorGuardProgram(activity.ValueProgram)
		return profile, node.ID.String(), SemanticDigest(ir.StableHash()), err
	}
	return goErrorGuardProgram{}, "", "", fmt.Errorf("guard activity is missing")
}

func parseGoErrorGuardProgram(raw string) (goErrorGuardProgram, error) {
	parts := strings.Split(raw, ";")
	if parts[0] == "go-error-guard:v3" {
		return parseGoHumanGuardProgram(parts)
	}
	if parts[0] == "go-error-guard:v2" {
		return parseGoReturnGuardProgram(parts)
	}
	if len(parts) != 6 || parts[0] != "go-error-guard:v1" {
		return goErrorGuardProgram{}, fmt.Errorf("unsupported guard computes profile")
	}
	values := map[string]string{}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || value == "" || values[key] != "" {
			return goErrorGuardProgram{}, fmt.Errorf("missing or duplicate guard field")
		}
		switch key {
		case "function", "writer", "diagnostic":
			if !token.IsIdentifier(value) || value == "_" {
				return goErrorGuardProgram{}, fmt.Errorf("guard selector is not an identifier")
			}
		case "source", "handler":
			if !ValidDigest(value) {
				return goErrorGuardProgram{}, fmt.Errorf("guard pin is not a digest")
			}
		default:
			return goErrorGuardProgram{}, fmt.Errorf("unknown guard field %q", key)
		}
		values[key] = value
	}
	if len(values) != 5 || values["writer"] == values["diagnostic"] {
		return goErrorGuardProgram{}, fmt.Errorf("guard requires five distinct keys and two writers")
	}
	return goErrorGuardProgram{
		function:   values["function"],
		writer:     values["writer"],
		diagnostic: values["diagnostic"],
		source:     values["source"],
		handler:    values["handler"],
	}, nil
}

func declineGoErrorGuard(report GoErrorGuardProposal, state, reason, class, next string) (GoErrorGuardProposal, error) {
	report.State, report.Reason = state, reason
	if state == "UNKNOWN" {
		pending := revisionPending("GO_SOURCE_SELECTION", "SELECT_DISCARDED_OUTPUT_ERROR", reason, next)
		pending.UnknownClass, pending.BlockedBy = class, []string{}
		report.Pending = &pending
	}
	return report, fmt.Errorf("%s: %s", state, reason)
}

// A proposal is AST/profile-bound only. Native type checking and the frozen
// behavior oracle remain outside this compiler's admission authority.
type goErrorGuardMatch struct {
	call      *ast.CallExpr
	handler   *ast.BlockStmt
	errorName string
}
