package bodyexecution

import "fmt"

func (p *compositionFieldProjection) bindRangeFieldSources(prior Composition, bindings []CompositionFieldSourceBinding) (bool, error) {
	locals, err := compositionSourceLocals(prior)
	if err != nil {
		return false, err
	}
	derived := false
	for _, n := range p.nodes {
		if n.kind != "range_record_write" {
			continue
		}
		binding := p.inventory.bindings[n.binding]
		var selected compositionLocalOrigin
		matches := 0
		for key, origin := range locals {
			if key != (conditionOriginKey{binding.RootActivityID, binding.ActivityID, origin.definition.NativeOrdinal}) ||
				origin.definition.VariableOrdinal != binding.VariableOrdinal ||
				origin.definition.Kind != "iteration" {
				continue
			}
			selected = origin
			matches++
		}
		if matches != 1 {
			continue
		}
		for field, member := range binding.Record.Fields {
			index := n.first + field
			bindings[index].Derivation = &CompositionFieldDerivedOrigin{
				Kind: "record_iteration_field", RecordID: binding.Record.ID,
				FieldID: member.ID, Local: selected.definition,
			}
			bindings[index].Source.Value = selected.source.Value
			bindings[index].Source.Target = selected.source.Declaration
			derived = true
		}
	}
	return derived, nil
}

func validateDerivedRangeFieldSource(b CompositionFieldSourceBinding, site CompositionFieldSite, identity CompositionFieldBinding, schema string) error {
	d := b.Derivation
	if d == nil {
		return nil
	}
	o := d.Local
	if b.Source.Value == nil || b.Source.Target == nil {
		return fmt.Errorf("range-derived source is absent")
	}
	if schema != "gooo/native-field-sources/v2" ||
		b.Origin != nil ||
		b.Kind != "write" ||
		site.Kind != "record_assignment" ||
		site.DependencyMode != "shared_rhs" ||
		d.Kind != "record_iteration_field" ||
		d.RecordID != identity.Record.ID ||
		d.FieldID != identity.Record.Fields[site.Field].ID ||
		o.Kind != "iteration" ||
		o.ActivityID != identity.ActivityID ||
		o.VariableOrdinal != identity.VariableOrdinal ||
		o.Name != identity.Name ||
		o.Declaration.Expression != identity.Name ||
		o.Operation != ":=" ||
		o.NativeOrdinal < 0 ||
		o.BodySHA256 == "" {
		return fmt.Errorf("range-derived field identity differs")
	}
	if o.Start < 0 ||
		o.End <= o.Start ||
		len(o.Expression) != o.End-o.Start ||
		o.Declaration.Start < 0 ||
		o.Declaration.End <= o.Declaration.Start ||
		len(o.Declaration.Expression) != o.Declaration.End-o.Declaration.Start {
		return fmt.Errorf("range-derived authored bounds differ")
	}
	for _, source := range []struct {
		actual   string
		expected string
	}{{b.Source.Value.Expression, o.Expression}, {b.Source.Target.Expression, o.Declaration.Expression}} {
		if source.actual != source.expected {
			return fmt.Errorf("range-derived source expression differs")
		}
	}
	return nil
}
