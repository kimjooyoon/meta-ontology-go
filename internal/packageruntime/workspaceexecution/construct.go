package workspaceexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

const workspaceConstructionSchema = "gooo/workspace-caller-construction/v1"

// ConstructionResult binds bounded joint search to the original package image,
// source call identities and untranslated caller examples. Evaluation is fresh
// evidence for the selected program, never feedback for this construction.
type ConstructionResult struct {
	Schema             string                          `json:"schema"`
	Program            Program                         `json:"program"`
	ConstructionCases  bodyexecution.CompositionCases  `json:"construction_cases"`
	Construction       bodyexecution.JointConstruction `json:"construction"`
	Evaluation         bodyexecution.JointEvaluation   `json:"evaluation"`
	ReplayedFromSHA256 string                          `json:"replayed_from_sha256,omitempty"`
	Scope              string                          `json:"scope"`
}

type ConstructOptions struct {
	ProgramBudget int
	ModelPath     string
	FillModelPath string
	GoBinary      string
}

// ConstructWorkspace leaves assembly declarations intact during lowering. The
// joint constructor can therefore reconsider imported helpers after each native
// caller execution, instead of freezing locally satisfactory choices first.
func ConstructWorkspace(ctx context.Context, manifest packageruntime.Manifest,
	construction, evaluation bodyexecution.CompositionCases, options ConstructOptions) (ConstructionResult, error) {
	r := ConstructionResult{Schema: workspaceConstructionSchema, ConstructionCases: construction,
		Scope: "source-bound package closure with bounded caller-guided construction; original package identities and examples retained; evaluation follows selection; finite observations only"}
	if err := constructionContext(ctx); err != nil {
		return r, err
	}
	var err error
	r.Program, err = Prepare(manifest)
	if err != nil {
		return r, err
	}
	feedback, err := translateCases(r.Program, construction)
	if err != nil {
		return r, err
	}
	cases, err := translateCases(r.Program, evaluation)
	if err != nil {
		return r, err
	}
	r.Construction, r.Evaluation, err = bodyexecution.ConstructAndEvaluateJoint(ctx, "workspace.gooo",
		[]byte(r.Program.Source), feedback, cases, bodyexecution.JointOptions{
			EntryActivity: constructionEntry(r.Program), ProgramBudget: options.ProgramBudget,
			ModelPath: options.ModelPath, FillModelPath: options.FillModelPath, GoBinary: options.GoBinary})
	return r, err
}

// ReplayWorkspaceConstruction rebinds the current package sources before
// reconstructing and executing every saved attempt. It loads no model. Prior
// evaluation observations remain historical; supplied cases are executed anew.
func ReplayWorkspaceConstruction(ctx context.Context, manifest packageruntime.Manifest,
	prior ConstructionResult, evaluation bodyexecution.CompositionCases, goBinary string) (ConstructionResult, error) {
	if err := constructionContext(ctx); err != nil {
		return ConstructionResult{}, err
	}
	program, err := Prepare(manifest)
	if err != nil {
		return ConstructionResult{}, err
	}
	if err := verifyWorkspaceConstruction(program, prior); err != nil {
		return ConstructionResult{}, err
	}
	cases, err := translateCases(program, evaluation)
	if err != nil {
		return ConstructionResult{}, err
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		return ConstructionResult{}, err
	}
	r := ConstructionResult{Schema: workspaceConstructionSchema, Program: program,
		ConstructionCases: prior.ConstructionCases, Construction: prior.Construction,
		ReplayedFromSHA256: sourceSHA256(encoded),
		Scope:              "current package image and caller cases rebound; all saved construction attempts re-executed without inference; prior evaluation remains historical; current evaluation is fresh"}
	r.Evaluation, err = bodyexecution.ReplayJointComposition(ctx, "workspace.gooo", []byte(program.Source),
		prior.Construction, cases, goBinary)
	return r, err
}

func verifyWorkspaceConstruction(program Program, prior ConstructionResult) error {
	if prior.Schema != workspaceConstructionSchema || !sameConstructionJSON(program, prior.Program) {
		return fmt.Errorf("saved construction differs from current package sources, identities, dependencies, entry or lowered graph")
	}
	feedback, err := translateCases(program, prior.ConstructionCases)
	if err != nil {
		return err
	}
	if !sameConstructionJSON(feedback, prior.Construction.ConstructionCases) ||
		prior.Construction.Initial.Plan.EntryActivity != constructionEntry(program) {
		return fmt.Errorf("saved construction caller cases or entry differ from the original workspace binding")
	}
	return nil
}

func constructionEntry(program Program) string {
	if program.PureCalls != nil {
		return program.Entry.LoweredName
	}
	return ""
}

func constructionContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("workspace construction requires a context")
	}
	return ctx.Err()
}

func sameConstructionJSON(left, right any) bool {
	a, leftErr := json.Marshal(left)
	b, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(a, b)
}
