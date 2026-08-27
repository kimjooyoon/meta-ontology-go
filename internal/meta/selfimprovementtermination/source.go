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
		interventions, err := buildInterventions(model.Path, model.raw, item.Program, model.SemanticDigest)
		if err != nil {
			return Input{}, err
		}
		return Input{
			Schema: InputSchema, Repository: repository, Subject: Consumer,
			Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
			ProofChoice: ProofChoice, Stage: TraceStage,
			Source: SourceCausality{
				Path: model.Path, SourceDigest: model.SourceDigest,
				SemanticDigest: model.SemanticDigest, CaseID: item.ID,
				CaseProgramDigest: item.ProgramDigest,
			},
			UpstreamDecision: item.UpstreamDecision, MaxSteps: item.MaxSteps,
			Trace: append([]Observation(nil), item.Trace...), Interventions: interventions,
		}, nil
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
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || !strings.HasPrefix(node.ValueProgram, SourceProgramSchema+"|") {
			continue
		}
		item, err := parseSourceCase(node.ValueProgram)
		if err != nil {
			return SourceModel{}, fmt.Errorf("source activity %q: %w", node.Name, err)
		}
		item.Activity = node.Name
		item.Program = node.ValueProgram
		item.ProgramDigest = digestBytes([]byte(node.ValueProgram))
		model.Cases = append(model.Cases, item)
	}
	if len(model.Cases) == 0 {
		return SourceModel{}, fmt.Errorf("termination source declares no executable cases")
	}
	sort.Slice(model.Cases, func(i, j int) bool { return model.Cases[i].ID < model.Cases[j].ID })
	for index := 1; index < len(model.Cases); index++ {
		if model.Cases[index-1].ID == model.Cases[index].ID {
			return SourceModel{}, fmt.Errorf("termination source case %q is duplicated", model.Cases[index].ID)
		}
	}
	return model, nil
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
	if len(trace) > maxSteps {
		return SourceCase{}, fmt.Errorf("value program trace exceeds max_steps")
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

func buildInterventions(filename string, raw []byte, program, semanticBefore string) ([]Intervention, error) {
	semanticSource := strings.Replace(string(raw), program, program+"|intervention=semantic-trace", 1)
	if semanticSource == string(raw) {
		return nil, fmt.Errorf("semantic intervention did not find the selected computes value")
	}
	semanticAfter, _, err := parseAndLower(filename, []byte(semanticSource))
	if err != nil {
		return nil, err
	}
	commentSource := append([]byte("// nonsemantic comment intervention\n"), raw...)
	commentAfter, _, err := parseAndLower(filename, commentSource)
	if err != nil {
		return nil, err
	}
	return []Intervention{
		{ID: "semantic-trace", Schema: InterventionSchema, Stage: InterventionStage, Step: 1,
			Reason:             "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST",
			SourceBeforeDigest: digestBytes(raw), SourceAfterDigest: digestBytes([]byte(semanticSource)),
			SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
			SourceChanged: true, SemanticChanged: semanticBefore != semanticAfter},
		{ID: "nonsemantic-comment", Schema: InterventionSchema, Stage: InterventionStage, Step: 2,
			Reason:             "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST",
			SourceBeforeDigest: digestBytes(raw), SourceAfterDigest: digestBytes(commentSource),
			SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: commentAfter,
			SourceChanged: true, SemanticChanged: semanticBefore != commentAfter},
	}, nil
}
