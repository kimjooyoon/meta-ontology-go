package bodyexecution

import (
	"fmt"
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type fieldSourceKey struct {
	root, activity, kind string
	ordinal              int
}
type retainedFieldSource struct {
	origin bodycodegen.BodyFieldOrigin
	source bodycodegen.BodyFieldSource
}

func compositionSourceFields(prior Composition) (map[fieldSourceKey]retainedFieldSource, error) {
	var origins []bodycodegen.BodyFieldOrigin
	var keys []fieldSourceKey
	for _, step := range prior.Steps {
		for _, origin := range step.Generation.Report.BodyFields {
			origins = append(origins, origin)
			keys = append(keys, fieldSourceKey{step.Generation.Report.ActivityID, origin.ActivityID, origin.Kind, origin.NativeOrdinal})
		}
	}
	result := map[fieldSourceKey]retainedFieldSource{}
	if len(origins) == 0 {
		return result, nil
	}
	sources, err := bodycodegen.BindBodyFieldSources([]byte(prior.GoooSource), origins)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		if _, duplicate := result[key]; duplicate {
			return nil, fmt.Errorf("field origin has duplicate source identity")
		}
		result[key] = retainedFieldSource{origins[i], sources[i]}
	}
	return result, nil
}

func nativeFieldSyntax(file *ast.File, prior Composition) (map[ast.Node]fieldSourceKey, map[fieldSourceKey]int) {
	nodes := map[ast.Node]fieldSourceKey{}
	populations := map[fieldSourceKey]int{}
	ids := arithmeticFunctionIDs(prior)
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || ids[fn.Name.Name][0] == "" {
			continue
		}
		names := ids[fn.Name.Name]
		counts := map[string]int{}
		bodycodegen.WalkFieldSyntax(fn.Body, func(slot bodycodegen.FieldSyntax) {
			nodes[slot.Node] = fieldSourceKey{names[0], names[1], slot.Kind, counts[slot.Kind]}
			counts[slot.Kind]++
		})
		for kind, count := range counts {
			populations[fieldSourceKey{names[0], names[1], kind, 0}] = count
		}
	}
	return nodes, populations
}

func (p *compositionFieldProjection) bindSources(file *ast.File, prior Composition) error {
	retained, err := compositionSourceFields(prior)
	if err != nil {
		return err
	}
	if len(retained) == 0 {
		return nil
	}
	nodes, populations := nativeFieldSyntax(file, prior)
	sourceSHA := digest([]byte(prior.GoooSource))
	bindings := make([]CompositionFieldSourceBinding, 0, len(p.sites)+len(p.uses))
	for group, sites := range [2][]CompositionFieldSite{p.sites, p.uses} {
		kind := "write"
		if group == 1 {
			kind = "use"
		}
		for index := range sites {
			missing := bodycodegen.UnmappedReadSource(sourceSHA, "SITE_HAS_NO_RETAINED_FIELD_SYNTAX")
			bindings = append(bindings, CompositionFieldSourceBinding{Kind: kind, SiteIndex: index, Source: bodycodegen.BodyFieldSource{Value: missing, Target: missing}})
		}
	}
	for _, node := range p.nodes {
		index := node.first
		switch node.kind {
		case "write", "part", "compound_write":
		case "use":
			index += len(p.sites)
		default:
			continue
		}
		key, found := nodes[node.node]
		if !found {
			continue
		}
		source, found := retained[key]
		if !found {
			continue
		}
		sites := p.sites
		if node.kind == "use" {
			sites = p.uses
		}
		site := sites[node.first]
		binding := p.inventory.bindings[site.Binding]
		population := populations[fieldSourceKey{key.root, key.activity, key.kind, 0}]
		if source.origin.NativePopulation != population || source.origin.RecordID != binding.Record.ID || source.origin.FieldID != binding.Record.Fields[site.Field].ID || source.origin.Kind != site.Kind || source.origin.Operation != site.Operation {
			missing := bodycodegen.UnmappedReadSource(sourceSHA, "TYPED_FIELD_PROJECTION_DIFFERS")
			bindings[index].Source = bodycodegen.BodyFieldSource{Value: missing, Target: missing}
			continue
		}
		origin := source.origin
		bindings[index].Origin = &origin
		bindings[index].Source = source.source
	}
	p.recordSources(prior, bindings)
	return nil
}
