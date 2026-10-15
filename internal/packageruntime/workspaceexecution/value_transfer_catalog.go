package workspaceexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func prepareValueTransferCatalog(ctx context.Context, result Result) (valueTransferCatalog, error) {
	c := valueTransferCatalog{byID: map[string]*valueTransferActivity{}, byName: map[string]*valueTransferActivity{}, binds: map[string]string{}}
	source := result.Composition.GoooSource
	if result.Schema != "gooo/workspace-body-execution/v1" || len(source) == 0 || len(source) > 128<<10 || sourceSHA256([]byte(source)) != result.Composition.SelectedSourceSHA256 {
		return c, fmt.Errorf("selected value origins require a retained source with its matching hash")
	}
	plan := result.Composition.Plan
	if len(plan.Activities)+len(plan.Preparations) > 128 || len(plan.Edges) > 256 {
		return c, fmt.Errorf("selected plan exceeds value-origin query bounds")
	}
	names, err := selectedValueActivityNames(result)
	if err != nil {
		return c, err
	}
	file, diagnostics := bodycodegen.ParseBodyFile("workspace.gooo", []byte(source))
	if diagnostics.HasErrors() {
		return c, diagnostics.Error()
	}
	if err := addSelectedValueActivities(ctx, &c, file, names); err != nil {
		return c, err
	}
	if err := bindSelectedValueActivities(&c, file, result); err != nil {
		return c, err
	}
	if err := retainSelectedValueFragments(&c, result); err != nil {
		return c, err
	}
	return c, nil
}

func selectedValueActivityNames(result Result) (map[string]string, error) {
	plan := result.Composition.Plan
	names := map[string]string{}
	register := func(name, id string) error {
		if name == "" || id == "" || names[name] != "" {
			return fmt.Errorf("selected activity name or identity is empty or duplicated")
		}
		names[name] = id
		return nil
	}
	for _, a := range plan.Activities {
		if err := register(a.Name, a.ID); err != nil {
			return nil, err
		}
	}
	for _, a := range plan.Preparations {
		if err := register(a.Name, a.ID); err != nil {
			return nil, err
		}
	}
	if result.Program.PureCalls != nil {
		for _, a := range result.Program.PureCalls.Activities {
			if previous := names[a.Activity.LoweredName]; previous != "" {
				if previous != a.LoweredID {
					return nil, fmt.Errorf("pure-call identity differs from selected plan")
				}
				continue
			}
			if err := register(a.Activity.LoweredName, a.LoweredID); err != nil {
				return nil, err
			}
		}
	}
	if len(names) > 128 {
		return nil, fmt.Errorf("selected call closure exceeds 128 activities")
	}
	return names, nil
}

func addSelectedValueActivities(ctx context.Context, c *valueTransferCatalog, file *syntax.File, names map[string]string) error {
	nodeCount := 0
	for _, decl := range file.Declarations {
		if err := ctx.Err(); err != nil {
			return err
		}
		d, ok := decl.(*syntax.ActivityDecl)
		if !ok || names[d.Name] == "" {
			continue
		}
		id := names[d.Name]
		actual, err := workspaceActivityIdentity(file.Namespace.Name, d.Name, d.ID)
		if err != nil || actual != id {
			return fmt.Errorf("selected plan activity identity differs from its source")
		}
		if c.byID[id] != nil {
			return fmt.Errorf("selected activity identity is duplicated")
		}
		transfer := bodycodegen.InspectBodyValueTransfers(d.ValueProgram, len(d.Inputs), syntax.ActivityDeclaredInputNames(d)...)
		a := &valueTransferActivity{id: id, name: d.Name, body: d.ValueProgram, arity: len(d.Inputs), inputNames: syntax.ActivityDeclaredInputNames(d), graph: transfer.Graph, nodes: map[string]bodycodegen.BodyValueOrigin{}, calls: map[string]bodycodegen.BodyValueCall{}}
		for _, node := range a.graph.Nodes {
			a.nodes[node.ID] = node
		}
		for _, call := range transfer.Calls {
			a.calls[call.NodeID] = call
		}
		c.byID[id], c.byName[d.Name] = a, a
		nodeCount += len(a.nodes)
		if nodeCount > 8192 {
			return fmt.Errorf("selected value graph exceeds 8192 nodes")
		}
	}
	if len(c.byName) != len(names) {
		return fmt.Errorf("selected source omits an activity from its plan")
	}
	return nil
}

func bindSelectedValueActivities(c *valueTransferCatalog, file *syntax.File, result Result) error {
	plan := result.Composition.Plan
	sourceEdges := map[string]bool{}
	for _, edge := range file.Bindings {
		p := c.byName[edge.Producer.Activity.Name]
		d := c.byName[edge.Consumer.Activity.Name]
		if p != nil && d != nil {
			sourceEdges[p.id+"\x00"+edge.Producer.Port.Name+"\x00"+d.id+"\x00"+edge.Consumer.Port.Name] = true
		}
	}
	for _, edge := range plan.Edges {
		key := edge.Consumer + "\x00" + edge.ConsumerPort
		if !sourceEdges[edge.Producer+"\x00"+edge.ProducerPort+"\x00"+edge.Consumer+"\x00"+edge.ConsumerPort] {
			return fmt.Errorf("selected bind differs from its source")
		}
		if c.byID[edge.Producer] == nil || c.byID[edge.Consumer] == nil || c.binds[key] != "" || edge.ProducerPort != "result" {
			return fmt.Errorf("selected bind cannot be resolved uniquely")
		}
		c.binds[key] = edge.Producer
	}
	if len(sourceEdges) != len(plan.Edges) {
		return fmt.Errorf("selected plan omits a source bind")
	}
	return nil
}

func retainSelectedValueFragments(c *valueTransferCatalog, result Result) error {
	generations, err := originQueryGenerations(result)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, g := range generations {
		if g.Report.BodyFill == nil || g.Report.BodyFill.SourceOrigin == nil {
			continue
		}
		o := g.Report.BodyFill.SourceOrigin
		a := c.byID[o.ActivityID]
		if a == nil || a.graph.BodySHA256 != o.SelectedBodySHA256 {
			continue
		}
		if len(o.Fragments) > 256 {
			return fmt.Errorf("candidate origin exceeds fragment bounds")
		}
		key := o.ActivityID + "\x00" + o.SelectedBodySHA256 + "\x00" + o.CandidateID
		if !seen[key] {
			a.origins = append(a.origins, o)
			seen[key] = true
		}
	}
	return nil
}
