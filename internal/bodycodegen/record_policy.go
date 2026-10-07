package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

type preparedRecordPolicy struct {
	body        preparedBody
	function    *ast.FuncDecl
	information types.Info
	input       types.Object
	receipt     *RecordAssemblyControl
}

// ValidateRecordAssemblyPolicy checks the source and adapter contract without
// loading a model, evaluating a candidate or executing an external process.
func ValidateRecordAssemblyPolicy(ctx context.Context, policy RecordAssemblyPolicy) error {
	_, err := prepareRecordPolicy(ctx, policy)
	return err
}

func prepareRecordPolicy(ctx context.Context, policy RecordAssemblyPolicy) (*preparedRecordPolicy, error) {
	if ctx == nil || len(policy.Source) == 0 || len(policy.Source) > 128<<10 || policy.Activity == "" {
		return nil, fmt.Errorf("assembly policy requires a context, bounded Gooo source and activity")
	}
	spec, err := SourceAssembly(ctx, "assembly-policy.gooo", []byte(policy.Source), policy.Activity)
	if err != nil || spec != nil {
		return nil, fmt.Errorf("assembly policy requires a fixed computes body: %v", err)
	}
	body, err := prepareActivityBody("assembly-policy.gooo", []byte(policy.Source), policy.Activity)
	if err != nil {
		return nil, fmt.Errorf("assembly policy: %w", err)
	}
	if err = validateRecordPolicySignature(body); err != nil {
		return nil, err
	}
	p := &preparedRecordPolicy{body: body, receipt: &RecordAssemblyControl{
		Schema: "gooo/record-assembly-control/v1", Policy: policy, SourceSHA256: digest([]byte(policy.Source)),
		GeneratedSHA256: digest(body.base.source), ActivityID: body.activityID}}
	err = p.prepareEvaluator()
	return p, err
}

func validateRecordPolicySignature(body preparedBody) error {
	if len(body.parameters) != 1 {
		return fmt.Errorf("assembly policy requires one observation record")
	}
	for _, contract := range []struct {
		name, id, kind string
		fields         []string
	}{
		{body.parameters[0].Type, "gooo://tools/assembly-observation", "integer", []string{"matched", "total", "best", "scored", "budget"}},
		{body.outputType, "gooo://tools/assembly-explanation", "string", []string{"state", "next_operation", "message"}},
	} {
		record := recordTypeByName(body.records, contract.name)
		if record == nil || record.ID != contract.id || len(record.Fields) != len(contract.fields) {
			return fmt.Errorf("assembly policy requires the observation/explanation record contracts")
		}
		for _, name := range contract.fields {
			found := false
			for _, field := range record.Fields {
				found = found || (field.Name == name && field.ID == contract.id+"/"+policyFieldID(name) &&
					field.TypeID == "urn:gooo:type:"+contract.kind && field.Presence != "optional")
			}
			if !found {
				return fmt.Errorf("assembly policy field %q differs from its typed contract", name)
			}
		}
	}
	return nil
}

func policyFieldID(name string) string {
	if name == "next_operation" {
		return "next-operation"
	}
	return name
}

func (p *preparedRecordPolicy) prepareEvaluator() error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "assembly-policy.go", p.body.base.source, parser.AllErrors)
	if err != nil {
		return err
	}
	p.information = types.Info{Types: make(map[ast.Expr]types.TypeAndValue),
		Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}
	if _, err = new(types.Config).Check(file.Name.Name, fset, []*ast.File{file}, &p.information); err != nil {
		return err
	}
	var ok bool
	p.function, ok = findFunction(file, p.body.activity.Name)
	if !ok || len(p.function.Type.Params.List) != 1 || len(p.function.Type.Params.List[0].Names) != 1 {
		return fmt.Errorf("assembly policy projection has no single-input function")
	}
	p.input = p.information.Defs[p.function.Type.Params.List[0].Names[0]]
	return nil
}

func (p *preparedRecordPolicy) decide(ctx context.Context, counts RecordPolicyCounts) (RecordPolicyDecision, error) {
	raw, err := json.Marshal(counts)
	if err != nil {
		return RecordPolicyDecision{}, err
	}
	input, err := decodeRecordCaseValue(raw, p.input.Type(), p.body.records)
	if err != nil {
		return RecordPolicyDecision{}, err
	}
	e := integerBodyEvaluator{context: ctx, information: p.information,
		environment: map[types.Object]any{p.input: input}}
	value, returned, err := e.evaluateBlock(p.function.Body)
	if err != nil || !returned {
		return RecordPolicyDecision{}, fmt.Errorf("assembly policy did not return: %v", err)
	}
	return p.decodeDecision(counts, value)
}

func (p *preparedRecordPolicy) decodeDecision(counts RecordPolicyCounts, value any) (RecordPolicyDecision, error) {
	record, ok := value.(recordBodyValue)
	if !ok {
		return RecordPolicyDecision{}, fmt.Errorf("assembly policy did not return a record")
	}
	decision := RecordPolicyDecision{Input: counts}
	for i, field := range recordTypeByName(p.body.records, p.body.outputType).Fields {
		switch field.Name {
		case "state":
			decision.State = record.Values[i].Text
		case "next_operation":
			decision.Operation = record.Values[i].Text
		case "message":
			decision.Message = record.Values[i].Text
		}
	}
	switch decision.Operation {
	case "CONTINUE_CANDIDATES", "EVALUATE_CANDIDATES":
		decision.Continue = true
	case "USE_OBSERVED_CANDIDATE", "OBSERVE_NEW_INPUTS", "EXPAND_DECLARED_CHOICES", "DECLARE_CASES", "RECONCILE_COUNTS":
	default:
		return decision, fmt.Errorf("unsupported assembly policy operation %q", decision.Operation)
	}
	return decision, nil
}

func observeRecordPolicy(ctx context.Context, p recordAssemblyPlan, r *RecordAssemblyReceipt, policy *preparedRecordPolicy) (bool, error) {
	decision, err := recordPolicyDecision(ctx, p, r, policy)
	if err != nil {
		return false, err
	}
	r.Control.Decisions = append(r.Control.Decisions, decision)
	return decision.Continue, nil
}

func recordPolicyDecision(ctx context.Context, p recordAssemblyPlan, r *RecordAssemblyReceipt, policy *preparedRecordPolicy) (RecordPolicyDecision, error) {
	counts := RecordPolicyCounts{Budget: min(p.spec.MaxAttempts, len(r.Ranking))}
	for _, attempt := range r.Attempts {
		if attempt.Total == 0 {
			counts.Budget--
			continue
		}
		counts.Scored++
		counts.Best = max(counts.Best, attempt.Passed)
	}
	latestIndex := len(r.Attempts) - 1
	for latestIndex > 0 && r.Attempts[latestIndex].Total == 0 {
		latestIndex--
	}
	latest := r.Attempts[latestIndex]
	counts.Matched, counts.Total = latest.Passed, latest.Total
	decision, err := policy.decide(ctx, counts)
	if err != nil {
		return decision, err
	}
	decision.AttemptIndex, decision.Mask = latestIndex, latest.Mask
	return decision, nil
}
