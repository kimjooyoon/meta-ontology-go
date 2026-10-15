package bodycodegen

import "fmt"

type RecordChoiceOrigin struct {
	Schema          string           `json:"schema"`
	ActivityID      string           `json:"activity_id"`
	ChoiceID        string           `json:"choice_id"`
	OptionID        string           `json:"option_id"`
	Kind            string           `json:"kind"`
	RecordID        string           `json:"record_id"`
	FieldID         string           `json:"field_id"`
	TemplateBodySHA string           `json:"template_body_sha256"`
	SelectedBodySHA string           `json:"selected_body_sha256"`
	TemplateDecoded FillOriginRange  `json:"template_decoded"`
	SelectedDecoded FillOriginRange  `json:"selected_decoded"`
	TargetDecoded   *FillOriginRange `json:"target_decoded,omitempty"`
	Expression      string           `json:"expression"`
}

func (p recordAssemblyPlan) choiceOrigins(mask uint16, body string) ([]RecordChoiceOrigin, error) {
	if p.scalar != nil {
		return nil, nil
	}
	selectedSHA, err := NormalizedConditionBodySHA256(body)
	if err != nil {
		return nil, err
	}
	var result []RecordChoiceOrigin
	for i, site := range p.sites {
		expression, option := site.choice.First, "value_first"
		if mask&(1<<i) != 0 {
			expression, option = "("+site.choice.Second+")", "value_second"
		}
		start := site.start
		for j, previous := range p.sites {
			if previous.start < site.start && mask&(1<<j) != 0 {
				start += len(previous.choice.Second) + 2 - (previous.end - previous.start)
			}
		}
		end := start + len(expression)
		if start < 0 || end > len(body) || body[start:end] != expression {
			return nil, fmt.Errorf("selected record choice range differs")
		}
		origin := RecordChoiceOrigin{Schema: "gooo/record-choice-source/v1", ActivityID: p.body.activityID,
			ChoiceID: site.choice.ID, OptionID: option, Kind: site.kind, RecordID: site.choice.RecordID, FieldID: site.choice.FieldID,
			TemplateBodySHA: digest([]byte(p.body.body)), SelectedBodySHA: selectedSHA,
			TemplateDecoded: FillOriginRange{site.start, site.end}, SelectedDecoded: FillOriginRange{start, end}, Expression: expression}
		for _, field := range p.body.base.report.BodyFields {
			if field.ActivityID == p.body.activityID && field.RecordID == origin.RecordID && field.FieldID == origin.FieldID && field.Value.Start == site.start && field.Value.End == site.end {
				span := FillOriginRange{field.Target.Start, field.Target.End}
				origin.TargetDecoded = &span
			}
		}
		result = append(result, origin)
	}
	return result, nil
}
