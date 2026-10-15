package bodyexecution

import (
	"fmt"
	"strings"
)

func validateFieldInventory(p *compositionFieldProjection) error {
	if p == nil || len(p.inventory.bindings) < 1 || len(p.inventory.bindings) > 4096 || len(p.sites) > 4096 || len(p.uses) > 65536 ||
		len(p.sourceSHA256) != 71 || !strings.HasPrefix(p.sourceSHA256, "sha256:") {
		return fmt.Errorf("field inventory or projection identity differs")
	}
	seen := map[conditionOriginKey]bool{}
	for _, binding := range p.inventory.bindings {
		key := conditionOriginKey{binding.RootActivityID, binding.ActivityID, binding.VariableOrdinal}
		if binding.RootActivityID == "" || binding.ActivityID == "" || binding.Name == "" || seen[key] || binding.Record.ID == "" || binding.Record.GoName == "" || len(binding.Record.Fields) < 1 || len(binding.Record.Fields) > 32 {
			return fmt.Errorf("field lexical binding or record identity differs")
		}
		seen[key] = true
		if err := validateFieldBinding(binding); err != nil {
			return err
		}

	}
	for group, sites := range [2][]CompositionFieldSite{p.sites, p.uses} {
		for _, site := range sites {
			if site.Binding < 0 || site.Binding >= len(p.inventory.bindings) || site.Field < 0 || site.Field >= len(p.inventory.bindings[site.Binding].Record.Fields) || site.Start < 0 || site.End <= site.Start {
				return fmt.Errorf("field site binding, field or projection span differs")
			}
			if group == 1 {
				if site.Kind != "field_read" && site.Kind != "record_read" {
					return fmt.Errorf("field use kind differs")
				}
				continue
			}
			switch site.Kind {
			case "parameter", "record_assignment", "record_constructor", "field_assignment":
			default:
				return fmt.Errorf("field write kind differs")
			}
			switch site.DependencyMode {
			case "call_parameter_or_input", "shared_rhs", "field_copy", "record_return", "scalar_rhs", "compound_update":
			default:
				return fmt.Errorf("field dependency mode differs")
			}
			if err := validateFieldOperation(site); err != nil {
				return err
			}
		}
	}
	return nil
}

func fieldFileFromTraces(traces []CompositionTrace) (nativeFieldFile, error) {
	f := nativeFieldFile{Schema: "gooo/native-field-observation/v1", Writes: [][]int{}, Uses: [][]int{}}
	for c, trace := range traces {
		if trace.CaseIndex != c {
			return f, fmt.Errorf("field observations require the complete original case population")
		}
		f.Frames = append(f.Frames, trace.FieldFrames)
		f.FrameDropped = append(f.FrameDropped, trace.FieldFrameDropped)
		f.WriteDropped = append(f.WriteDropped, trace.FieldWriteDropped)
		f.UseDropped = append(f.UseDropped, trace.FieldUseDropped)
		for _, write := range trace.FieldWrites {
			done := 0
			if write.Completed {
				done = 1
			}
			f.Writes = append(f.Writes, []int{c, write.SiteIndex, write.Frame, write.UseStart, write.UseEnd, write.UsesDropped, done, write.ParentAssignment, write.Completion})
		}
		for _, use := range trace.FieldUses {
			if use.Reason != fieldUseReason(use.Assignment) {
				return f, fmt.Errorf("field use missing-assignment reason differs")
			}
			f.Uses = append(f.Uses, []int{c, use.SiteIndex, use.Frame, use.Assignment, use.CompletionClock})
		}
	}
	return f, nil
}
