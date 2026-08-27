package selfimprovementtermination

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type SourceCase struct {
	ID               string
	Activity         string
	Program          string
	ProgramDigest    string
	UpstreamDecision string
	MaxSteps         int
	Trace            []Observation
}

type SourceModel struct {
	Path           string
	SourceDigest   string
	SemanticDigest string
	Cases          []SourceCase
	raw            []byte
}

func BuildInput(root, repository, sourcePath, caseID string) (Input, error) {
	model, err := LoadSource(root, sourcePath)
	if err != nil {
		return Input{}, err
	}
	for _, item := range model.Cases {
		if item.ID != caseID {
			continue
		}
		input := Input{
			Schema: InputSchema, Repository: repository, Subject: Consumer,
			Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
			ProofChoice: ProofChoice, Stage: TraceStage,
			Source: SourceCausality{
				Path: model.Path, SourceDigest: model.SourceDigest,
				SemanticDigest: model.SemanticDigest, CaseID: item.ID,
				CaseProgramDigest: item.ProgramDigest,
			},
			UpstreamDecision: item.UpstreamDecision, MaxSteps: item.MaxSteps,
			Trace: append([]Observation(nil), item.Trace...),
		}
		interventions, err := buildInterventions(model.Path, model.raw, item.Program, model.SemanticDigest, item)
		if err != nil {
			return Input{}, err
		}
		input.Interventions = interventions
		return input, nil
	}
	return Input{}, fmt.Errorf("termination source case %q is not declared", caseID)
}

func LoadSource(root, sourcePath string) (SourceModel, error) {
	canonicalPath := filepath.ToSlash(filepath.Clean(sourcePath))
	if canonicalPath == "." || filepath.IsAbs(sourcePath) || strings.HasPrefix(canonicalPath, "../") {
		return SourceModel{}, fmt.Errorf("termination source path is outside the repository contract")
	}
	filename := filepath.Join(root, filepath.FromSlash(canonicalPath))
	raw, err := os.ReadFile(filename)
	if err != nil {
		return SourceModel{}, err
	}
	semanticDigest, ir, err := parseAndLower(canonicalPath, raw)
	if err != nil {
		return SourceModel{}, err
	}
	model := SourceModel{Path: canonicalPath, SourceDigest: digestBytes(raw), SemanticDigest: semanticDigest, raw: append([]byte(nil), raw...)}
	cases, err := sourceCases(ir)
	if err != nil {
		return SourceModel{}, err
	}
	model.Cases = cases
	if len(model.Cases) == 0 {
		return SourceModel{}, fmt.Errorf("termination source declares no executable cases")
	}
	return model, nil
}

func sourceCases(ir semantic.IR) ([]SourceCase, error) {
	cases := make([]SourceCase, 0)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || !strings.HasPrefix(node.ValueProgram, SourceProgramSchema+"|") {
			continue
		}
		item, err := parseSourceCase(node.ValueProgram)
		if err != nil {
			return nil, fmt.Errorf("source activity %q: %w", node.Name, err)
		}
		item.Activity = node.Name
		item.Program = node.ValueProgram
		item.ProgramDigest = digestBytes([]byte(node.ValueProgram))
		cases = append(cases, item)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	for index := 1; index < len(cases); index++ {
		if cases[index-1].ID == cases[index].ID {
			return nil, fmt.Errorf("termination source case %q is duplicated", cases[index].ID)
		}
	}
	return cases, nil
}

func findSourceCase(ir semantic.IR, caseID string) (SourceCase, error) {
	cases, err := sourceCases(ir)
	if err != nil {
		return SourceCase{}, err
	}
	for _, item := range cases {
		if item.ID == caseID {
			return item, nil
		}
	}
	return SourceCase{}, fmt.Errorf("termination source case %q is not declared after intervention", caseID)
}

func parseAndLower(filename string, raw []byte) (string, semantic.IR, error) {
	file, diagnostics := syntax.ParseFile(filename, string(raw))
	if err := diagnostics.Error(); err != nil || file == nil {
		if err == nil {
			err = fmt.Errorf("parser returned no source tree")
		}
		return "", semantic.IR{}, fmt.Errorf("termination Gooo source: %w", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", semantic.IR{}, fmt.Errorf("termination Gooo semantic lowering: %w", err)
	}
	return "sha256:" + ir.StableHash(), ir, nil
}

func parseSourceCase(program string) (SourceCase, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 5 || parts[0] != SourceProgramSchema {
		return SourceCase{}, fmt.Errorf("value program schema is not %q", SourceProgramSchema)
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return SourceCase{}, fmt.Errorf("malformed value program field %q", part)
		}
		if _, exists := values[key]; exists {
			return SourceCase{}, fmt.Errorf("duplicate value program field %q", key)
		}
		values[key] = value
	}
	for _, key := range []string{"id", "max_steps", "upstream", "trace"} {
		if values[key] == "" {
			return SourceCase{}, fmt.Errorf("value program field %q is missing", key)
		}
	}
	maxSteps, err := strconv.Atoi(values["max_steps"])
	if err != nil || maxSteps < 1 || maxSteps > MaxTraceSteps {
		return SourceCase{}, fmt.Errorf("value program max_steps is invalid")
	}
	trace, err := parseTrace(values["trace"])
	if err != nil {
		return SourceCase{}, err
	}
	if len(trace) == 0 || len(trace) > maxSteps {
		return SourceCase{}, fmt.Errorf("value program trace length is outside max_steps")
	}
	return SourceCase{ID: values["id"], UpstreamDecision: values["upstream"], MaxSteps: maxSteps, Trace: trace}, nil
}

func parseTrace(raw string) ([]Observation, error) {
	items := strings.Split(raw, ";")
	trace := make([]Observation, 0, len(items))
	for _, item := range items {
		fields := strings.Split(item, ",")
		if len(fields) != 7 {
			return nil, fmt.Errorf("value program trace observation has %d fields, want 7", len(fields))
		}
		step, stepErr := strconv.Atoi(fields[0])
		beforeRank, beforeErr := strconv.Atoi(fields[3])
		afterRank, afterErr := strconv.Atoi(fields[4])
		if stepErr != nil || beforeErr != nil || afterErr != nil {
			return nil, fmt.Errorf("value program trace has non-integer coordinate")
		}
		trace = append(trace, Observation{Stage: TraceStage, Step: step,
			BeforeState: stateDigest(fields[1]), AfterState: stateDigest(fields[2]),
			BeforeRank: beforeRank, AfterRank: afterRank, Decision: fields[5], Reason: fields[6]})
	}
	return trace, nil
}

func buildInterventions(filename string, raw []byte, program, semanticBefore string, baseline SourceCase) ([]Intervention, error) {
	baseClass := classify(Input{UpstreamDecision: baseline.UpstreamDecision, MaxSteps: baseline.MaxSteps, Trace: baseline.Trace})
	targetMaxSteps, targetUpstream, targetTrace := semanticInterventionTarget(baseClass)
	semanticProgram, err := rewriteSourceProgram(program, targetMaxSteps, targetUpstream, targetTrace)
	if err != nil {
		return nil, err
	}
	semanticSource := strings.Replace(string(raw), program, semanticProgram, 1)
	if semanticSource == string(raw) {
		return nil, fmt.Errorf("semantic intervention did not find the selected computes value")
	}
	semanticAfter, semanticIR, err := parseAndLower(filename, []byte(semanticSource))
	if err != nil {
		return nil, err
	}
	semanticCase, err := findSourceCase(semanticIR, baseline.ID)
	if err != nil {
		return nil, err
	}
	semanticClass := classify(Input{UpstreamDecision: semanticCase.UpstreamDecision, MaxSteps: semanticCase.MaxSteps, Trace: semanticCase.Trace})
	if semanticBefore == semanticAfter || (baseClass.decision == semanticClass.decision && baseClass.resolution == semanticClass.resolution) ||
		!canonicalOpenTransition(transitions(semanticClass, len(semanticCase.Trace))) {
		return nil, fmt.Errorf("semantic intervention did not causally change the selected trace outcome")
	}
	commentSource := append([]byte("// nonsemantic comment intervention\n"), raw...)
	commentAfter, commentIR, err := parseAndLower(filename, commentSource)
	if err != nil {
		return nil, err
	}
	commentCase, err := findSourceCase(commentIR, baseline.ID)
	if err != nil {
		return nil, err
	}
	commentClass := classify(Input{UpstreamDecision: commentCase.UpstreamDecision, MaxSteps: commentCase.MaxSteps, Trace: commentCase.Trace})
	sourceBefore := digestBytes(raw)
	semanticIntervention := InterventionOutcome{
		SourceDigest: digestBytes([]byte(semanticSource)), SemanticDigest: semanticAfter,
		Decision: semanticClass.decision, Resolution: semanticClass.resolution,
		ClaimTransitions: transitions(semanticClass, len(semanticCase.Trace)),
	}
	commentIntervention := InterventionOutcome{
		SourceDigest: digestBytes(commentSource), SemanticDigest: commentAfter,
		Decision: commentClass.decision, Resolution: commentClass.resolution,
		ClaimTransitions: transitions(commentClass, len(commentCase.Trace)),
	}
	baselineOutcome := InterventionOutcome{
		SourceDigest: sourceBefore, SemanticDigest: semanticBefore,
		Decision: baseClass.decision, Resolution: baseClass.resolution,
		ClaimTransitions: transitions(baseClass, len(baseline.Trace)),
	}
	return []Intervention{
		{ID: "semantic-trace", Schema: InterventionSchema, Stage: InterventionStage, Step: 1,
			Reason:             "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST",
			SourceBeforeDigest: sourceBefore, SourceAfterDigest: semanticIntervention.SourceDigest,
			SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
			SourceChanged: true, SemanticChanged: semanticBefore != semanticAfter,
			Baseline: baselineOutcome, Intervened: semanticIntervention},
		{ID: "nonsemantic-comment", Schema: InterventionSchema, Stage: InterventionStage, Step: 2,
			Reason:             "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST",
			SourceBeforeDigest: sourceBefore, SourceAfterDigest: commentIntervention.SourceDigest,
			SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: commentAfter,
			SourceChanged: true, SemanticChanged: semanticBefore != commentAfter,
			Baseline: baselineOutcome, Intervened: commentIntervention},
	}, nil
}

func semanticInterventionTarget(base classification) (int, string, string) {
	if base.decision == DecisionInProgress {
		return 1, UpstreamChanged, "1,semantic-a,semantic-b,0,1,CHANGED,METAPROGRAM_STATE_CHANGED"
	}
	return 4, UpstreamChanged, "1,semantic-a,semantic-b,0,1,CHANGED,METAPROGRAM_STATE_CHANGED"
}

func rewriteSourceProgram(program string, maxSteps int, upstream, trace string) (string, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != SourceProgramSchema {
		return "", fmt.Errorf("cannot rewrite non-case computes value")
	}
	found := map[string]bool{}
	for index := 1; index < len(parts); index++ {
		key, value, ok := strings.Cut(parts[index], "=")
		if !ok {
			return "", fmt.Errorf("cannot rewrite malformed computes field %q", parts[index])
		}
		switch key {
		case "max_steps":
			parts[index] = fmt.Sprintf("max_steps=%d", maxSteps)
			found[key] = true
		case "upstream":
			parts[index] = "upstream=" + upstream
			found[key] = true
		case "trace":
			parts[index] = "trace=" + trace
			found[key] = true
		default:
			if value == "" {
				return "", fmt.Errorf("cannot rewrite empty computes field %q", key)
			}
		}
	}
	for _, key := range []string{"max_steps", "upstream", "trace"} {
		if !found[key] {
			return "", fmt.Errorf("cannot rewrite computes value without %s", key)
		}
	}
	return strings.Join(parts, "|"), nil
}
