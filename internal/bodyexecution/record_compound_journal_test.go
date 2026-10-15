package bodyexecution

import "testing"

func assertRecordCompoundJournal(t *testing.T, r CompositionRuntime) {
	t.Helper()
	inputs := 0
	for _, binding := range r.FieldBindings {
		if binding.InputKind == "root_input" {
			inputs++
			if binding.InputPort != "input0" || binding.Name != "기록" {
				t.Fatal("canonical input port and authored name", binding)
			}
		}
	}
	if inputs != 1 {
		t.Fatal("root record parameter population", inputs)
	}
	for _, trace := range r.Traces {
		updates := 0
		for index, write := range trace.FieldWrites {
			site := r.FieldSites[write.SiteIndex]
			if site.DependencyMode != "compound_update" {
				continue
			}
			updates++
			if !write.Completed || write.UseEnd <= write.UseStart || write.UsesDropped != 0 {
				t.Fatal(write)
			}
			read := trace.FieldUses[write.UseStart]
			useSite := r.FieldUseSites[read.SiteIndex]
			if read.Assignment < 0 || read.Assignment >= index || useSite.Binding != site.Binding ||
				useSite.Field != site.Field || read.Frame != write.Frame {
				t.Fatal("prior field identity", write, read)
			}
			previous := trace.FieldWrites[read.Assignment]
			if previous.Completion >= write.Completion {
				t.Fatal("update temporal order", previous, write)
			}
			if site.Expression == "delta" && trace.Bridge.Fields[index].OtherEnd-trace.Bridge.Fields[index].OtherStart != 1 {
				t.Fatal("compound update lost its scalar-local operand", trace.Bridge.Fields[index])
			}
		}
		if updates != 4 {
			t.Fatal("compound update population", updates)
		}
	}
	for _, binding := range r.FieldSources.Bindings {
		if binding.Kind != "write" {
			continue
		}
		site := r.FieldSites[binding.SiteIndex]
		if site.DependencyMode != "compound_update" {
			continue
		}
		if binding.Origin == nil || binding.Origin.Operation != site.Operation ||
			binding.Source.Target.Status != "MAPPED" || binding.Source.Value.Status != "MAPPED" {
			t.Fatal("compound authored source", binding)
		}
	}
}

func TestCompoundFieldWindowRequiresPreviousValue(t *testing.T) {
	p := &compositionFieldProjection{sites: []CompositionFieldSite{{Binding: 1, Field: 2, DependencyMode: "compound_update"}},
		uses: []CompositionFieldSite{{Binding: 1, Field: 2}, {Binding: 1, Field: 3}}}
	write := []int{0, 0, 4, 0, 1, 0, 1, -1, 7}
	uses := [][]int{{0, 0, 4, 1, 6}}
	if err := validateCompoundFieldReadWindow(write, uses, p); err != nil {
		t.Fatal(err)
	}
	write[4] = 0
	if err := validateCompoundFieldReadWindow(write, uses, p); err == nil {
		t.Fatal("empty previous-value window")
	}
	write[4], uses[0][1] = 1, 1
	if err := validateCompoundFieldReadWindow(write, uses, p); err == nil {
		t.Fatal("wrong prior field")
	}
}
