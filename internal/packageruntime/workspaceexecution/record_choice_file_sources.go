package workspaceexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func workspaceRecordChoiceOrigin(composition bodyexecution.Composition, identity bodyexecution.CompositionFieldBinding,
	field int, source bodycodegen.BodyFieldSource) *bodycodegen.RecordChoiceOrigin {
	if source.Value == nil || source.Value.Status != "MAPPED" {
		return nil
	}
	var result *bodycodegen.RecordChoiceOrigin
	for _, step := range composition.Steps {
		receipt := step.Generation.Report.RecordAssembly
		if receipt == nil || step.Generation.Report.ActivityID != identity.RootActivityID {
			continue
		}
		for _, origin := range receipt.ChoiceOrigins {
			if origin.ActivityID != identity.ActivityID || origin.RecordID != identity.Record.ID || origin.FieldID != identity.Record.Fields[field].ID ||
				origin.SelectedBodySHA != source.Value.BodySHA256 || origin.SelectedDecoded.Start != source.Value.DecodedStart || origin.SelectedDecoded.End != source.Value.DecodedEnd {
				continue
			}
			if result != nil {
				return nil
			}
			copy := origin
			result = &copy
		}
	}
	return result
}

func workspaceRecordChoiceSource(input workspaceFillInput, origin bodycodegen.RecordChoiceOrigin) *bodyexecution.CompositionFieldChoiceSource {
	result := &bodyexecution.CompositionFieldChoiceSource{Schema: "gooo/field-choice-input-source/v1",
		ChoiceID: origin.ChoiceID, OptionID: origin.OptionID, Kind: origin.Kind, RecordID: origin.RecordID, FieldID: origin.FieldID,
		SelectedBodySHA: origin.SelectedBodySHA, SelectedDecoded: origin.SelectedDecoded,
		Declaration: unmappedRecordChoiceLocation(input, "CHOICE_DECLARATION_UNAVAILABLE"),
		Value:       unmappedRecordChoiceLocation(input, "CHOICE_EXPRESSION_UNAVAILABLE")}
	declarations, err := bodycodegen.InspectRecordChoiceDeclarationOrigins(input.filename, []byte(input.source), input.activity)
	if err != nil {
		return result
	}
	for _, declaration := range declarations {
		if declaration.ChoiceID != origin.ChoiceID || declaration.Kind != origin.Kind {
			continue
		}
		result.Declaration = mappedRecordChoiceLocation(input, declaration.DeclarationSource, input.source[declaration.DeclarationSource.Start:declaration.DeclarationSource.End])
		if origin.OptionID == "value_second" {
			result.Value = mappedRecordChoiceLocation(input, declaration.AlternativeSource, declaration.Alternative)
		} else if origin.OptionID == "value_first" {
			result.Value = recordChoiceTemplateLocation(input, origin.TemplateBodySHA, origin.TemplateDecoded)
		}
		break
	}
	return result
}
