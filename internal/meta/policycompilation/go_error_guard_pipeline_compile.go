package policycompilation

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type goGuardPipelinePlan struct {
	guard          goErrorGuardProgram
	guardActivity  string
	formatActivity string
	toolchain      string
	semanticDigest string
}

func compileGoGuardPipeline(filename string, source []byte) (goGuardPipelinePlan, error) {
	ir, file, err := lowerPolicy(filename, source, "goerrorguard", "goerrorguard")
	if err != nil {
		return goGuardPipelinePlan{}, err
	}
	if err := checkGoGuardPipelineNodes(ir, file); err != nil {
		return goGuardPipelinePlan{}, err
	}
	guard, render, err := selectGoGuardPipelineActivities(file)
	if err != nil {
		return goGuardPipelinePlan{}, err
	}
	guardID, err := goGuardPipelineActivityID(ir, guard, "Source", "RawCandidate")
	if err != nil {
		return goGuardPipelinePlan{}, err
	}
	renderID, err := goGuardPipelineActivityID(ir, render, "RawCandidate", "Candidate")
	if err != nil {
		return goGuardPipelinePlan{}, err
	}
	profile, err := parseGoErrorGuardProgram(guard.ValueProgram)
	if err != nil {
		return goGuardPipelinePlan{}, err
	}
	version, err := parseGoGuardRenderer(render.ValueProgram)
	if err != nil {
		return goGuardPipelinePlan{}, err
	}
	return goGuardPipelinePlan{
		guard: profile, guardActivity: guardID, formatActivity: renderID,
		toolchain: version, semanticDigest: SemanticDigest(ir.StableHash()),
	}, nil
}

func checkGoGuardPipelineNodes(ir semantic.IR, file *syntax.File) error {
	if len(file.Decls) != 5 || len(file.Bindings) != 0 || len(ir.RuntimeBindings) != 0 ||
		len(ir.Policies) != 0 || len(ir.Graph.Nodes()) != 5 {
		return fmt.Errorf("pipeline requires exactly three entities and two activities")
	}
	for _, name := range []string{"Source", "RawCandidate", "Candidate"} {
		node, ok := ir.Graph.NodeByName(ir.Namespace, name)
		if !ok || node.Kind != semantic.Entity || len(node.Fields) != 0 {
			return fmt.Errorf("pipeline entity %s differs", name)
		}
	}
	return nil
}

func selectGoGuardPipelineActivities(file *syntax.File) (*syntax.ActivityDecl, *syntax.ActivityDecl, error) {
	var guard, render *syntax.ActivityDecl
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if len(activity.Inputs) != 1 {
			return nil, nil, fmt.Errorf("pipeline activity needs exactly one input")
		}
		switch {
		case activity.Inputs[0].Name == "Source" && activity.Output == "RawCandidate" && guard == nil:
			guard = activity
		case activity.Inputs[0].Name == "RawCandidate" && activity.Output == "Candidate" && render == nil:
			render = activity
		default:
			return nil, nil, fmt.Errorf("pipeline has a duplicate or disconnected activity")
		}
	}
	if guard == nil || render == nil {
		return nil, nil, fmt.Errorf("pipeline activity is missing")
	}
	return guard, render, nil
}

func goGuardPipelineActivityID(ir semantic.IR, activity *syntax.ActivityDecl, from, to string) (string, error) {
	node, found := ir.Graph.NodeByName(ir.Namespace, activity.Name)
	input, inputOK := ir.Graph.NodeByName(ir.Namespace, from)
	output, outputOK := ir.Graph.NodeByName(ir.Namespace, to)
	if !found || !inputOK || !outputOK || node.Kind != semantic.Activity ||
		!activity.ValueProgramPresent || activity.ValueProgram != node.ValueProgram ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: input.ID}) ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID}) {
		return "", fmt.Errorf("pipeline activity provenance differs")
	}
	return node.ID.String(), nil
}

func parseGoGuardRenderer(raw string) (string, error) {
	parts := strings.Split(raw, ";")
	if len(parts) != 2 || parts[0] != "go-source-format:v1" {
		return "", fmt.Errorf("unsupported canonical Go rendering profile")
	}
	key, version, ok := strings.Cut(parts[1], "=")
	if !ok || key != "toolchain" || !strings.HasPrefix(version, "go1.") ||
		strings.ContainsAny(version, " \t\r\n") {
		return "", fmt.Errorf("renderer needs an explicit released Go toolchain")
	}
	return version, nil
}
