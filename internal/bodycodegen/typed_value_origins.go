package bodycodegen

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"
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
	Node        uint16                 `json:"node"`
	Port        string                 `json:"port"`
	FieldID     string                 `json:"field_id,omitempty"`
	Declaration *TypedInputDeclaration `json:"declaration,omitempty"`
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
	if c.compactRecords {
		r.Schema, flow.Schema = "gooo/typed-value-origins/v2", recordVersionFlowSchema
		r.Scope = "source may-dependence; no input values, runtime path, candidate selection or model call; required scalar or nominal record fields with source events and immutable field-root versions; collections/optional fields unresolved; 512 nodes, 64 live bindings, branch/helper depth 16"
		flow.Scope += "; record layouts, symbolic values and source versions retain separate identities"
	}
	if err != nil {
		r.Reason, flow.Reason = err.Error(), err.Error()
		return r, nil
	}
	return c.finishTypedOrigins(source, r, returns.value)
}

func (c *recordFlowBuilder) typedOriginOutputs(value flowValue) []TypedValueOutput {
	if value.record < 0 {
		return []TypedValueOutput{{Root: value.scalar}}
	}
	var outputs []TypedValueOutput
	for i, field := range c.body.records[value.record].Fields {
		root := c.nominalRoot(value.fields[i], RecordFlowSpan{ActivityID: c.body.activityID, View: "derived-record-output"})
		if c.compactRecords {
			root = c.add("return", [2]uint16{root}, RecordFlowSpan{ActivityID: c.body.activityID, View: "record-field-projection"}, field.ID, 0, "record_field_projection")
			c.linkRecordVersion(root, value.scalar)
		}
		outputs = append(outputs, TypedValueOutput{FieldID: field.ID, Root: root})
	}
	return outputs
}

func (c *recordFlowBuilder) finishTypedOrigins(source []byte, r TypedValueOrigins, value flowValue) (TypedValueOrigins, error) {
	flow := r.Flow
	r.Outputs = c.typedOriginOutputs(value)
	c.exportRecordVersions(flow)
	if c.err != nil {
		r.Outputs = nil
		r.Reason = c.err.Error()
		flow.Reason = r.Reason
		flow.RecordLayouts = nil
		flow.RecordValues = nil
		flow.RecordVersions = nil
		return r, nil
	}
	r.Status, flow.Status = "RESOLVED", "RESOLVED"
	flow.Nodes = append([]RecordFlowNode(nil), c.nodes[:c.count]...)
	for i := range flow.Nodes {
		if flow.Nodes[i].Span.View == "input-port" && flow.Nodes[i].Span.ActivityID == "" {
			flow.Nodes[i].Span.ActivityID = c.body.activityID
		}
	}
	for _, name := range c.helperOrder {
		flow.Helpers = append(flow.Helpers, c.helpers[name].function.identity)
	}
	for _, node := range flow.Nodes {
		if node.Kind == "input" {
			index := int(node.Local) - 1
			r.Inputs = append(r.Inputs, TypedValueInput{Node: node.ID, Port: inputbinding.PositionName(len(c.body.parameters), index), FieldID: node.FieldID, Declaration: inputDeclarationSource(c.body, source, index)})
		}
	}
	var err error
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
