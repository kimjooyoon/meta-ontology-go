package compiler

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type Result struct {
	Policy         model.FreshnessPolicy
	PolicyDigest   string
	SemanticDigest string
	IR             semantic.IR
}

// Compile is the canonical freshness compiler. It parses the .gooo source,
// lowers ordinary declarations through the repository's bidirectional
// semantic lowerer, and compiles only the formal freshness values carried by
// the syntax AST.
func Compile(filename string, source []byte) (Result, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Result{}, fmt.Errorf("freshness source parse failed: %d diagnostics", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Result{}, fmt.Errorf("freshness source lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return Result{}, fmt.Errorf("freshness semantic invariant failed: %w", err)
	}
	policy, err := compilePolicy(file.Freshness)
	if err != nil {
		return Result{}, err
	}
	if err := ValidatePolicy(policy); err != nil {
		return Result{}, err
	}
	// The repository IR exposes a bare hex stable hash. Evidence artifacts use
	// the shared sha256:<hex> digest grammar, so preserve the same canonical IR
	// bytes while normalizing the wire representation at this boundary.
	return Result{Policy: policy, PolicyDigest: model.DigestJSON(policy), SemanticDigest: model.DigestBytes([]byte(ir.SemanticCanonical())), IR: ir}, nil
}

func ValidatePolicy(policy model.FreshnessPolicy) error {
	if policy.Schema != model.PolicySchema || policy.ComparisonPolicy != "earliest_changed" || policy.PriorClaimState != model.ClaimOpen ||
		policy.BoundaryPolicy != "logical_epoch_environment" || policy.RawMaterialPolicy != "raw_material_digest" ||
		policy.SemanticPolicy != "comments_ignored" || policy.ClaimLedgerPolicy != "open_discharge_or_preserve" ||
		policy.EffectPolicy != "read_only_ci_before_after" || len(policy.Axes) != model.AxisTotal {
		return fmt.Errorf("freshness policy invariant failed")
	}
	names := []string{"subject", "material", "recipe", "environment", "runner", "verifier"}
	stages := []string{model.StageSubject, model.StageMaterial, model.StageRecipe, model.StageEnvironment, model.StageRunner, model.StageVerifier}
	for index, axis := range policy.Axes {
		if axis.Name != names[index] || axis.Stage != stages[index] || axis.Step != "compare-"+axis.Name {
			return fmt.Errorf("freshness axis invariant failed at %d", index)
		}
	}
	return nil
}

func compilePolicy(declarations []*syntax.FreshnessDecl) (model.FreshnessPolicy, error) {
	values := make(map[string][]string, len(declarations))
	for _, declaration := range declarations {
		if declaration == nil || declaration.Kind == "" {
			return model.FreshnessPolicy{}, fmt.Errorf("freshness policy declaration is empty")
		}
		if _, exists := values[declaration.Kind]; exists {
			return model.FreshnessPolicy{}, fmt.Errorf("duplicate freshness policy %q", declaration.Kind)
		}
		items := make([]string, len(declaration.Values))
		for index, value := range declaration.Values {
			if value.Name == "" {
				return model.FreshnessPolicy{}, fmt.Errorf("freshness policy %q has empty value %d", declaration.Kind, index)
			}
			items[index] = value.Name
		}
		values[declaration.Kind] = items
	}
	required := map[string][]string{
		"axes":                {"subject", "material", "recipe", "environment", "runner", "verifier"},
		"comparison_policy":   {"earliest_changed"},
		"prior_claim_state":   {model.ClaimOpen},
		"boundary_policy":     {"logical_epoch_environment"},
		"raw_material_policy": {"raw_material_digest"},
		"semantic_policy":     {"comments_ignored"},
		"claim_ledger_policy": {"open_discharge_or_preserve"},
		"effect_policy":       {"read_only_ci_before_after"},
	}
	for kind, expected := range required {
		actual, ok := values[kind]
		if !ok || len(actual) != len(expected) {
			return model.FreshnessPolicy{}, fmt.Errorf("freshness policy %q is missing or has wrong arity", kind)
		}
		for index := range expected {
			if actual[index] != expected[index] {
				return model.FreshnessPolicy{}, fmt.Errorf("freshness policy %q value %q is not allowed", kind, actual[index])
			}
		}
	}
	if len(values) != len(required) {
		return model.FreshnessPolicy{}, fmt.Errorf("freshness policy has unknown declaration")
	}
	axisNames := values["axes"]
	stages := []string{model.StageSubject, model.StageMaterial, model.StageRecipe, model.StageEnvironment, model.StageRunner, model.StageVerifier}
	axes := make([]model.AxisPolicy, len(axisNames))
	for index, name := range axisNames {
		axes[index] = model.AxisPolicy{Name: name, Stage: stages[index], Step: "compare-" + name, Reason: ""}
	}
	return model.FreshnessPolicy{
		Schema: model.PolicySchema, Axes: axes,
		ComparisonPolicy: values["comparison_policy"][0], PriorClaimState: values["prior_claim_state"][0],
		BoundaryPolicy: values["boundary_policy"][0], RawMaterialPolicy: values["raw_material_policy"][0],
		SemanticPolicy: values["semantic_policy"][0], ClaimLedgerPolicy: values["claim_ledger_policy"][0],
		EffectPolicy: values["effect_policy"][0],
	}, nil
}
