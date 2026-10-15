package bodyexecution

import "fmt"

func validateFieldChoiceSource(r CompositionRuntime, binding CompositionFieldFileBinding) error {
	choice := binding.Choice
	if choice == nil {
		return nil
	}
	if binding.Kind != "write" || choice.Schema != "gooo/field-choice-input-source/v1" || choice.ChoiceID == "" ||
		(choice.OptionID != "value_first" && choice.OptionID != "value_second") || choice.SelectedBodySHA == "" {
		return fmt.Errorf("field choice source identity differs")
	}
	site := r.FieldSites[binding.SiteIndex]
	identity := r.FieldBindings[site.Binding]
	if choice.RecordID != identity.Record.ID || choice.FieldID != identity.Record.Fields[site.Field].ID ||
		!((choice.Kind == "field_value" && site.Kind == "record_constructor") || (choice.Kind == "field_update" && site.Kind == "field_assignment")) {
		return fmt.Errorf("field choice typed site differs")
	}
	matched := false
	for _, source := range r.FieldSources.Bindings {
		if source.Kind == binding.Kind && source.SiteIndex == binding.SiteIndex && source.Source.Value != nil {
			value := source.Source.Value
			matched = value.Status == "MAPPED" && value.BodySHA256 == choice.SelectedBodySHA &&
				value.DecodedStart == choice.SelectedDecoded.Start && value.DecodedEnd == choice.SelectedDecoded.End
		}
	}
	if !matched {
		return fmt.Errorf("field choice selected expression differs")
	}
	for _, location := range []CompositionLocalFileLocation{choice.Declaration, choice.Value} {
		if err := validateInputFileLocation(location); err != nil {
			return err
		}
		if location.Status == "MAPPED" && location.Expression == "" {
			return fmt.Errorf("field choice source expression absent")
		}
	}
	if choice.Declaration.SourceSHA256 != choice.Value.SourceSHA256 || choice.Declaration.PackagePath != choice.Value.PackagePath ||
		choice.Declaration.Filename != choice.Value.Filename || choice.Declaration.ActivityID != choice.Value.ActivityID {
		return fmt.Errorf("field choice declaration and value origins differ")
	}
	return nil
}
