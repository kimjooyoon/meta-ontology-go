package bodycodegen

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const pathContextSchema = "gooo/compiler-typed-path-context/v2"
const semanticPathContextSchema = "gooo/compiler-typed-path-context/v3"

type PathContextInput struct {
	DecisionID           string `json:"decision_id"`
	OriginalIntentSHA    string `json:"original_intent_sha256"`
	NaturalIntentSHA     string `json:"natural_intent_sha256"`
	InputSHA             string `json:"input_sha256,omitempty"`
	SourceFeatureSHA     string `json:"source_features_sha256,omitempty"`
	Bytes                int    `json:"bytes"`
	CallerPrefixReplaced bool   `json:"caller_prefix_replaced,omitempty"`
}

type PathModelContextReceipt struct {
	Schema            string               `json:"schema"`
	Status            string               `json:"status"`
	ActivityID        string               `json:"activity_id"`
	SourceSemanticSHA string               `json:"source_semantic_sha256"`
	OriginalPlanSHA   string               `json:"original_plan_sha256"`
	RankedPlanSHA     string               `json:"ranked_plan_sha256,omitempty"`
	MetadataSHA       string               `json:"model_metadata_sha256,omitempty"`
	ArtifactSHA       string               `json:"model_artifact_sha256,omitempty"`
	ModelFingerprint  string               `json:"model_fingerprint,omitempty"`
	FeatureVersion    string               `json:"feature_version"`
	ArithmeticVersion string               `json:"arithmetic_version,omitempty"`
	Inputs            []PathContextInput   `json:"inputs"`
	DeclinedDecision  string               `json:"declined_decision,omitempty"`
	Reason            string               `json:"reason,omitempty"`
	SeedSkipped       bool                 `json:"seed_skipped,omitempty"`
	FeedbackSkipped   bool                 `json:"feedback_skipped,omitempty"`
	Scope             string               `json:"scope"`
	DeclaredInputs    *PathDeclaredInputs  `json:"complete_declared_inputs,omitempty"`
	ContractCases     *ContractCaseContext `json:"declared_contract_cases,omitempty"`
}

// DeclaredInputs preserves the original caller text, separate from the source-
// projected canonical model input. It grants no source or prediction authority.
type PathDeclaredInputs struct {
	Text      string `json:"text"`
	SHA256    string `json:"sha256"`
	Bytes     int    `json:"bytes"`
	Decisions int    `json:"declared_decisions"`
}

// Context construction follows source binding; it never reads test outcomes.
// A representation decline skips optional ranking, retaining deterministic TDD.
func preparePathModelContext(ctx context.Context, document pathplan.Document, original *pathplan.PreparedPlan,
	model *decision.Model, activityID, sourceSemanticSHA string) (*pathplan.PreparedPlan,
	*PathModelContextReceipt, bool, error) {
	if model == nil || model.FeatureVersion() != decision.SplitContextIntentFeatureVersion &&
		model.FeatureVersion() != decision.SemanticContextIntentFeatureVersion {
		return original, nil, false, nil
	}
	return prepareCompilerPathContextWithFeature(ctx, document, original, activityID, sourceSemanticSHA,
		model.MetadataSHA256(), model.FeatureVersion())
}

func prepareCompilerPathContext(ctx context.Context, document pathplan.Document, original *pathplan.PreparedPlan,
	activityID, sourceSemanticSHA, metadataSHA string) (*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	return prepareCompilerPathContextWithFeature(ctx, document, original, activityID, sourceSemanticSHA,
		metadataSHA, decision.SplitContextIntentFeatureVersion)
}

func prepareCompilerPathContextWithFeature(ctx context.Context, document pathplan.Document,
	original *pathplan.PreparedPlan, activityID, sourceSemanticSHA, metadataSHA, featureVersion string) (
	*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	if featureVersion != decision.SplitContextIntentFeatureVersion && featureVersion != decision.SemanticContextIntentFeatureVersion {
		return nil, nil, false, fmt.Errorf("unsupported compiler context feature version")
	}
	receipt := &PathModelContextReceipt{Schema: pathContextSchema, Status: "ENCODED", ActivityID: activityID,
		SourceSemanticSHA: sourceSemanticSHA, OriginalPlanSHA: original.PlanSHA256(), MetadataSHA: metadataSHA,
		FeatureVersion: featureVersion, Inputs: make([]PathContextInput, 0, len(document.Plan.Decisions)),
		Scope: "fallback-normalized validated typed nodes and source-relative legal alternatives, bound to original source; one-hop node facts, not a complete semantic IR or intended-answer authority"}
	if featureVersion == decision.SemanticContextIntentFeatureVersion {
		receipt.Schema, receipt.Scope = semanticPathContextSchema,
			"direct fallback-normalized typed source features and complete natural intent; lossy graph/option facts exclude literal magnitudes, name spellings, tests and outcomes; source bound before projection"
	}
	plan := document.Plan
	plan.Decisions = append([]pathplan.Choice(nil), plan.Decisions...)
	var facts bodyplan.Plan
	if featureVersion == decision.SplitContextIntentFeatureVersion {
		facts = fallbackContextFacts(document.Plan)
	}
	for i, choice := range plan.Decisions {
		if err := ctx.Err(); err != nil {
			return nil, receipt, false, err
		}
		text, input, reason := encodeCompilerPathContext(featureVersion, facts, original, choice)
		receipt.Inputs = append(receipt.Inputs, input)
		if reason != "" {
			receipt.Status, receipt.Reason, receipt.DeclinedDecision = "DECLINED_TO_DETERMINISTIC", reason, choice.ID
			return original, receipt, true, nil
		}
		plan.Decisions[i].Intent = text
	}
	prepared, err := pathplan.Prepare(plan)
	if err != nil {
		return nil, receipt, false, err
	}
	receipt.RankedPlanSHA = prepared.PlanSHA256()
	return prepared, receipt, false, nil
}

func encodeCompilerPathContext(version string, facts bodyplan.Plan, original *pathplan.PreparedPlan,
	choice pathplan.Choice) (string, PathContextInput, string) {
	if version == decision.SplitContextIntentFeatureVersion {
		return encodePathContext(facts, choice)
	}
	intent := choice.Intent
	input := PathContextInput{DecisionID: choice.ID, OriginalIntentSHA: digest([]byte(intent))}
	if index := strings.LastIndex(intent, "intent: "); index >= 0 {
		intent, input.CallerPrefixReplaced = intent[index+len("intent: "):], true
	}
	input.NaturalIntentSHA = digest([]byte(intent))
	if len(intent) == 0 {
		return "", input, "EMPTY_INTENT_AFTER_FEATURE_SEPARATOR"
	}
	fields, err := original.SourceFeatures(choice.ID)
	if err != nil {
		return "", input, err.Error()
	}
	input.SourceFeatureSHA = digest(fields[:])
	input.Bytes = len("gooo;sem64=") + len(fields)*2 + len(";intent: ") + len(intent)
	text, err := decision.EncodeSemanticContextInput(fields, intent)
	if err != nil {
		return "", input, "COMBINED_CONTEXT_INTENT_EXCEEDS_MODEL_BOUND"
	}
	input.InputSHA = digest([]byte(text))
	return text, input, ""
}

// The fact snapshot applies all declared fallbacks without changing the search
// plan. Its copied slices are bounded to 128 nodes by the preceding preparation.
func fallbackContextFacts(plan pathplan.Plan) bodyplan.Plan {
	base := plan.Base
	base.Expressions = append([]bodyplan.Expr(nil), base.Expressions...)
	base.Statements = append([]bodyplan.Stmt(nil), base.Statements...)
	base.Root = append([]int(nil), base.Root...)
	for _, choice := range plan.Decisions {
		for _, option := range choice.Options {
			if option.Label != choice.Fallback {
				continue
			}
			switch choice.Kind {
			case pathplan.LocalReference:
				base.Expressions[choice.Target].Name = option.Name
			case pathplan.AssignmentTarget:
				base.Statements[choice.Target].Name = option.Name
			case pathplan.OperandOrder:
				if option.Reverse {
					e := &base.Expressions[choice.Target]
					e.Left, e.Right = e.Right, e.Left
				}
			case pathplan.BranchLayout:
				if option.Reverse {
					s := &base.Statements[choice.Target]
					s.Then, s.Else = s.Else, s.Then
				}
			case pathplan.RootOrder:
				base.Root = append(base.Root[:0], option.Order...)
			}
		}
	}
	return base
}

type pathContextBuffer struct {
	bytes  [decision.InputMaxBytes]byte
	used   int
	wanted int
	bound  bool
}

func (b *pathContextBuffer) add(value string) {
	b.wanted += len(value)
	if len(value) > len(b.bytes)-b.used {
		b.bound = true
		return
	}
	b.used += copy(b.bytes[b.used:], value)
}

func (b *pathContextBuffer) integer(value int64) {
	var digits [24]byte
	b.add(string(strconv.AppendInt(digits[:0], value, 10)))
}

func (b *pathContextBuffer) order(values []int) {
	b.add("[")
	for _, value := range values {
		b.integer(int64(value))
		b.add(",")
	}
	b.add("]")
}

func (b *pathContextBuffer) expression(base bodyplan.Plan, index int, children bool) {
	e := base.Expressions[index]
	b.integer(int64(index))
	b.add(":" + e.Kind)
	switch e.Kind {
	case bodyplan.ExprInput, bodyplan.ExprLocal:
		b.add(":" + e.Name)
	case bodyplan.ExprInt:
		b.add(":")
		b.integer(e.Int)
	case bodyplan.ExprBool:
		b.add(":" + strconv.FormatBool(e.Bool))
	case bodyplan.ExprBinary, bodyplan.ExprHole:
		b.add(":" + e.Operation)
		if e.Kind == bodyplan.ExprHole {
			b.add(":" + e.Fallback)
		}
		if children {
			b.add("(l=")
			b.expression(base, e.Left, false)
			b.add(";r=")
			b.expression(base, e.Right, false)
			b.add(")")
		}
	}
}

func encodePathContext(base bodyplan.Plan, choice pathplan.Choice) (string, PathContextInput, string) {
	intent := choice.Intent
	input := PathContextInput{DecisionID: choice.ID, OriginalIntentSHA: digest([]byte(intent))}
	if index := strings.LastIndex(intent, "intent: "); index >= 0 {
		intent, input.CallerPrefixReplaced = intent[index+len("intent: "):], true
	}
	input.NaturalIntentSHA = digest([]byte(intent))
	if len(intent) == 0 {
		return "", input, "EMPTY_INTENT_AFTER_FEATURE_SEPARATOR"
	}
	var b pathContextBuffer
	b.add("gooo;result=" + string(base.ResultType) + ";kind=" + choice.Kind + ";target=")
	b.integer(int64(choice.Target))
	b.add(";fallback=" + choice.Fallback + ";basis=source_fallback;node=")
	switch choice.Kind {
	case pathplan.LocalReference, pathplan.OperandOrder:
		b.expression(base, choice.Target, true)
	case pathplan.AssignmentTarget, pathplan.BranchLayout:
		s := base.Statements[choice.Target]
		b.add(s.Kind + ":" + s.Name + ";expr=")
		b.expression(base, s.Expr, true)
		if s.Kind == bodyplan.StmtIf {
			b.add(";then=")
			b.order(s.Then)
			b.add(";else=")
			b.order(s.Else)
		}
	case pathplan.RootOrder:
		b.order(base.Root)
	default:
		return "", input, fmt.Sprintf("UNSUPPORTED_VALIDATED_CHOICE_%s", choice.Kind)
	}
	b.add(";legal=")
	addSourceRelativeOptions(&b, choice)
	b.add(";intent: ")
	b.add(intent)
	input.Bytes = b.wanted
	if b.bound {
		return "", input, "COMBINED_CONTEXT_INTENT_EXCEEDS_MODEL_BOUND"
	}
	text := string(b.bytes[:b.used])
	input.InputSHA = digest([]byte(text))
	return text, input, ""
}

// Reverse flags describe the actual option relative to the bound source body.
// Search still receives the original plan: normalizing that base would swap twice.
func addSourceRelativeOptions(b *pathContextBuffer, choice pathplan.Choice) {
	fallbackReverse := false
	for _, option := range choice.Options {
		if option.Label == choice.Fallback {
			fallbackReverse = option.Reverse
		}
	}
	for _, option := range choice.Options {
		b.add(option.Label)
		switch choice.Kind {
		case pathplan.OperandOrder, pathplan.BranchLayout:
			b.add(":reverse_source=" + strconv.FormatBool(option.Reverse != fallbackReverse))
		case pathplan.LocalReference, pathplan.AssignmentTarget:
			b.add(":name=" + option.Name)
		case pathplan.RootOrder:
			b.add(":order=")
			b.order(option.Order)
		}
		b.add("|")
	}
}
