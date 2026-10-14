package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// TypedValueOrigins is a source-only field-sensitive graph for one activity.
// It does not select an assembly candidate or observe an execution path.
type TypedValueOrigins struct {
	Schema       string                 `json:"schema"`
	Status       string                 `json:"status"`
	Reason       string                 `json:"reason,omitempty"`
	Filename     string                 `json:"filename"`
	SourceSHA256 string                 `json:"source_sha256"`
	ActivityID   string                 `json:"activity_id"`
	Flow         *RecordValueFlow       `json:"flow"`
	Outputs      []TypedValueOutput     `json:"outputs,omitempty"`
	Inputs       []TypedValueInput      `json:"inputs,omitempty"`
	Sources      []BodyExpressionSource `json:"node_sources,omitempty"`
	Scope        string                 `json:"scope"`
}

type TypedValueOutput struct {
	FieldID string `json:"field_id,omitempty"`
	Root    uint16 `json:"root"`
}

type TypedValueInput struct {
	Node    uint16 `json:"node"`
	Port    string `json:"port"`
	FieldID string `json:"field_id,omitempty"`
}

// InspectTypedValueOrigins reads ordinary computes bodies without generating
// Go, calling a model, evaluating cases or changing legacy value-origin graphs.
func InspectTypedValueOrigins(filename string, source []byte, activity string) (TypedValueOrigins, error) {
	p, err := prepareActivityBodySource(filename, source, activity)
	if err != nil {
		return TypedValueOrigins{}, err
	}
	flow := &RecordValueFlow{Schema: recordFlowSchema, Status: "UNRESOLVED", ActivityID: p.activityID,
		BodySHA256: digest([]byte(p.body)), BodyView: "computes",
		Scope: "typed source definitions, field copies, field writes, fixed helper bindings and conditional joins"}
	r := TypedValueOrigins{Schema: "gooo/typed-value-origins/v1", Status: "UNRESOLVED", Filename: filename,
		SourceSHA256: digest(source), ActivityID: p.activityID, Flow: flow,
		Scope: "source may-dependence; no input values, runtime path, candidate selection or model call; flat scalar record fields; 512 nodes, 64 live bindings, branch/helper depth 16"}
	returns := recordFlowReturns{}
	c := recordFlowBuilder{body: p, fset: token.NewFileSet(), view: "computes", bodyView: "computes",
		activityID: p.activityID, returns: &returns, info: &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}}
	c.sourceValidation = c.validateOriginFile
	fn, err := c.parse()
	if err == nil {
		state := c.parameters()
		live := c.block(fn.Body, &state)
		err = c.err
		if err == nil && (live || !returns.seen) {
			err = fmt.Errorf("FLOW_RETURN_UNRESOLVED")
		}
	}
	if err != nil {
		r.Reason, flow.Reason = err.Error(), err.Error()
		return r, nil
	}
	r.Status, flow.Status = "RESOLVED", "RESOLVED"
	flow.Nodes = append([]RecordFlowNode(nil), c.nodes[:c.count]...)
	for i := range flow.Nodes {
		if flow.Nodes[i].Span.View == "input-port" && flow.Nodes[i].Span.ActivityID == "" {
			flow.Nodes[i].Span.ActivityID = p.activityID
		}
	}
	for _, name := range c.helperOrder {
		flow.Helpers = append(flow.Helpers, c.helpers[name].function.identity)
	}
	if returns.value.record < 0 {
		r.Outputs = []TypedValueOutput{{Root: returns.value.scalar}}
	} else {
		for i, field := range c.body.records[returns.value.record].Fields {
			r.Outputs = append(r.Outputs, TypedValueOutput{FieldID: field.ID, Root: returns.value.fields[i]})
		}
	}
	for _, node := range flow.Nodes {
		if node.Kind == "input" {
			r.Inputs = append(r.Inputs, TypedValueInput{Node: node.ID, Port: p.parameters[int(node.Local)-1].Name, FieldID: node.FieldID})
		}
	}
	r.Sources, err = typedOriginSources(source, flow.Nodes)
	return r, err
}

func (c *recordFlowBuilder) validateOriginFile(file *ast.File) error {
	for _, record := range c.body.records {
		for _, field := range record.Fields {
			if field.Presence != "" || field.Cardinality != "" {
				return fmt.Errorf("FLOW_REQUIRES_REQUIRED_SCALAR_FIELDS")
			}
		}
	}
	validate := func(block *ast.BlockStmt, parameters []InputParameter) error {
		readonly := make(map[string]bool, len(parameters))
		for _, p := range parameters {
			readonly[p.Name] = true
		}
		if _, err := validateBlockInputs(block, readonly, map[string]bool{}, true, c.body.records...); err != nil {
			return err
		}
		var err error
		ast.Inspect(block, func(n ast.Node) bool {
			if d, ok := n.(*ast.ValueSpec); ok && len(d.Values) != 1 {
				err = fmt.Errorf("FLOW_INITIAL_VALUE_REQUIRED")
			}
			return err == nil
		})
		return err
	}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		parameters := c.body.parameters
		if helper, ok := c.helpers[fn.Name.Name]; ok {
			parameters = helper.function.parameters
		}
		if err := validate(fn.Body, parameters); err != nil {
			return err
		}
	}
	return nil
}
