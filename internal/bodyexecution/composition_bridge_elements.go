package bodyexecution

import "fmt"

func bindBridgeElementReads(f nativeBridgeFile, traces []CompositionTrace) error {
	if f.Schema == "gooo/native-field-local-bridge/v1" {
		if f.ElementReads != nil {
			return fmt.Errorf("v1 bridge cannot contain element references")
		}
		return nil
	}
	if f.ElementReads == nil || len(f.ElementReads) != len(f.Fields) {
		return fmt.Errorf("bridge element reference population differs")
	}
	for i, row := range f.Fields {
		bridge := traces[row[0]].Bridge
		bridge.ElementReads = append(bridge.ElementReads, f.ElementReads[i])
	}
	return nil
}

func iterationFieldSources(r CompositionRuntime) []*CompositionFieldSourceBinding {
	derived := make([]*CompositionFieldSourceBinding, len(r.FieldSites))
	if r.FieldSources != nil {
		for i := range r.FieldSources.Bindings {
			b := &r.FieldSources.Bindings[i]
			if b.Kind == "write" && b.SiteIndex >= 0 && b.SiteIndex < len(derived) && b.Derivation != nil {
				derived[b.SiteIndex] = b
			}
		}
	}
	return derived
}

func validateBridgeElementReads(t CompositionTrace, r CompositionRuntime) error {
	reads := t.Bridge.ElementReads
	if reads == nil {
		return nil // Historical journals made no element-to-field claim.
	}
	if len(reads) != len(t.FieldWrites) {
		return fmt.Errorf("bridge element references differ from field writes")
	}
	if len(reads) > 0 && !t.ReadsObserved {
		return fmt.Errorf("bridge element references have no observed reads")
	}
	derived := iterationFieldSources(r)
	last, previous := -1, -1
	for i, ref := range reads {
		write := t.FieldWrites[i]
		if write.SiteIndex < 0 || write.SiteIndex >= len(derived) {
			return fmt.Errorf("iteration field write has no retained site")
		}
		d := derived[write.SiteIndex]
		if d == nil {
			if ref != -1 {
				return fmt.Errorf("non-iteration field claims an element read")
			}
			continue
		}
		if ref == -2 && t.ReadsDropped > 0 {
			continue
		}
		if ref < 0 || ref >= len(t.Reads) {
			return fmt.Errorf("iteration element read is absent or outside its case")
		}
		if ref < last || ref == last && (previous < 0 || !sameIterationBinding(t, r, previous, i)) {
			return fmt.Errorf("iteration element reference precedes or reuses another binding")
		}
		if previous >= 0 && sameIterationBinding(t, r, previous, i) && ref != last {
			return fmt.Errorf("one record binding names different element reads")
		}
		if err := validateBridgeElementSource(t.Reads[ref], r, write.SiteIndex, d); err != nil {
			return err
		}
		last, previous = ref, i
	}
	return nil
}

func sameIterationBinding(t CompositionTrace, r CompositionRuntime, a, b int) bool {
	x, y := t.FieldWrites[a], t.FieldWrites[b]
	return r.FieldSites[x.SiteIndex].Binding == r.FieldSites[y.SiteIndex].Binding &&
		x.Frame == y.Frame && t.Bridge.Fields[a].StartClock == t.Bridge.Fields[b].StartClock
}

func validateBridgeElementSource(read CompositionReadObservation, r CompositionRuntime, field int, source *CompositionFieldSourceBinding) error {
	if read.SiteIndex < 0 || read.SiteIndex >= len(r.ReadSites) {
		return fmt.Errorf("iteration element read has no retained site")
	}
	if field < 0 || field >= len(r.FieldSites) || r.FieldSites[field].Binding < 0 ||
		r.FieldSites[field].Binding >= len(r.FieldBindings) {
		return fmt.Errorf("iteration field has no retained binding")
	}
	site, binding := r.ReadSites[read.SiteIndex], r.FieldBindings[r.FieldSites[field].Binding]
	if site.Kind != "range_value" || site.RootActivityID != binding.RootActivityID ||
		site.ActivityID != binding.ActivityID || site.ElementType != binding.Record.GoName {
		return fmt.Errorf("iteration element read type or activity differs")
	}
	// Read and local decoded coordinates belong to distinct projection views.
	// Their shared anchor is the exact range in the selected Gooo source.
	if s, value := site.Source, source.Source.Value; s != nil && value != nil &&
		s.Status == "MAPPED" && value.Status == "MAPPED" &&
		(s.SourceSHA256 != value.SourceSHA256 || s.BodySHA256 != value.BodySHA256 ||
			s.Start != value.Start || s.End != value.End || s.Expression != value.Expression) {
		return fmt.Errorf("iteration element read names a different authored range")
	}
	return nil
}

func hasBridgeElementReads(t CompositionTrace) bool {
	if t.Bridge != nil {
		for _, ref := range t.Bridge.ElementReads {
			if ref != -1 {
				return true
			}
		}
	}
	return false
}

func summarizeBridgeElementReads(s *CompositionBridgeSummary, t CompositionTrace) {
	if t.Bridge.ElementReads == nil {
		return
	}
	s.Schema = "gooo/native-field-local-bridge-summary/v2"
	for _, read := range t.Bridge.ElementReads {
		if read >= 0 {
			s.ElementLinks++
			if t.Reads[read].Origin == nil && t.Reads[read].Construction == nil {
				s.UnresolvedElements++
			}
		} else if read == -2 {
			s.MissingElements++
		}
	}
}

func bindBridgeElementSummary(s *CompositionBridgeSummary, r CompositionRuntime) error {
	if s.Schema != "gooo/native-field-local-bridge-summary/v2" {
		return nil
	}
	if r.FieldSources == nil || r.ReadObservation == nil {
		return fmt.Errorf("element bridge requires its original source and read observations")
	}
	s.MetadataSHA256 = compositionDigest([]any{s.MetadataSHA256, r.FieldSources.SHA256,
		r.ReadObservation.SHA256, r.ReadSites, r.ReadInputs})
	s.Scope += "; record-iteration fields reference original case-local element reads; missing reads and unresolved input origins remain partial"
	return nil
}
