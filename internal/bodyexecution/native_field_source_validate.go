package bodyexecution

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateFieldSources(r CompositionRuntime) error {
	set := r.FieldSources
	if set == nil {
		if r.FieldFileSources != nil {
			return fmt.Errorf("field file sources have no selected source table")
		}
		return nil
	}
	if r.FieldObservation == nil ||
		(set.Schema != "gooo/native-field-sources/v1" &&
			set.Schema != "gooo/native-field-sources/v2") ||
		set.SelectedSourceSHA256 != r.SelectedSourceSHA256 ||
		set.SHA256 != fieldSourcesDigest(*set) ||
		len(set.Bindings) != len(r.FieldSites)+len(r.FieldUseSites) {
		return fmt.Errorf("field source table identity or population differs")
	}
	if err := validateFieldReturnSources(r); err != nil {
		return err
	}
	seen := map[[2]int]bool{}
	for _, binding := range set.Bindings {
		sites, kind := r.FieldSites, 0
		if binding.Kind == "use" {
			sites, kind = r.FieldUseSites, 1
		} else if binding.Kind != "write" {
			return fmt.Errorf("field source subject differs")
		}
		key := [2]int{kind, binding.SiteIndex}
		if binding.SiteIndex < 0 || binding.SiteIndex >= len(sites) || seen[key] {
			return fmt.Errorf("field source site identity differs")
		}
		seen[key] = true
		site := sites[binding.SiteIndex]
		identity := r.FieldBindings[site.Binding]
		if err := validateFieldSourceOrigin(binding, site, identity, set.Schema); err != nil {
			return err
		}
		bodySHA := fieldSourceBodySHA(binding)
		if err := validateSelectedFieldSource(binding, r.SelectedSourceSHA256, bodySHA); err != nil {
			return err
		}
	}
	return validateFieldFileSources(r)
}

func fieldSourceBodySHA(b CompositionFieldSourceBinding) string {
	if b.Origin != nil {
		return b.Origin.BodySHA256
	}
	if b.Derivation != nil {
		return b.Derivation.Local.BodySHA256
	}
	return ""
}

func validateFieldSourceOrigin(binding CompositionFieldSourceBinding, site CompositionFieldSite,
	identity CompositionFieldBinding, schema string) error {
	if err := validateDerivedRangeFieldSource(binding, site, identity, schema); err != nil {
		return err
	}
	if o := binding.Origin; o != nil {
		switch o.Kind {
		case "field_assignment", "record_constructor", "field_read":
		default:
			return fmt.Errorf("field source syntax kind differs")
		}
		if o.ActivityID != identity.ActivityID ||
			o.Kind != site.Kind ||
			o.Operation != site.Operation ||
			o.RecordID != identity.Record.ID ||
			o.FieldID != identity.Record.Fields[site.Field].ID ||
			o.NativeOrdinal < 0 ||
			o.NativeOrdinal >= o.NativePopulation ||
			o.BodySHA256 == "" {
			return fmt.Errorf("field source typed identity differs")
		}
		for _, span := range []bodycodegen.BodyElementOrigin{o.Value, o.Target} {
			if span.Start < 0 || span.End <= span.Start || len(span.Expression) != span.End-span.Start {
				return fmt.Errorf("field source authored range differs")
			}
		}
	}
	return nil
}

func validateSelectedFieldSource(binding CompositionFieldSourceBinding, selectedSHA, bodySHA string) error {
	for _, source := range []*bodycodegen.BodyExpressionSource{binding.Source.Value, binding.Source.Target} {
		if source == nil || source.View != "selected_gooo_source" || source.SourceSHA256 != selectedSHA {
			return fmt.Errorf("field selected source identity differs")
		}
		switch source.Status {
		case "UNMAPPED":
			if source.Reason == "" || source.Start != 0 || source.End != 0 {
				return fmt.Errorf("unmapped field source claims a range")
			}
		case "MAPPED":
			if bodySHA == "" ||
				source.Reason != "" ||
				source.BodySHA256 != bodySHA ||
				source.Start < 0 ||
				source.End <= source.Start ||
				source.DecodedStart < 0 ||
				source.DecodedEnd <= source.DecodedStart {
				return fmt.Errorf("mapped field source is incomplete")
			}
		default:
			return fmt.Errorf("field source mapping status differs")
		}
	}
	return nil
}
