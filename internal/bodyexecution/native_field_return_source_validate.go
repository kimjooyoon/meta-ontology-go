package bodyexecution

import "fmt"

func validateFieldReturnSources(r CompositionRuntime) error {
	for _, binding := range r.FieldBindings {
		if binding.ExpressionKind != "return" {
			continue
		}
		source := binding.Declaration
		if source == nil || source.View != "selected_gooo_source" || source.SourceSHA256 != r.SelectedSourceSHA256 {
			return fmt.Errorf("return expression declaration source differs")
		}
		switch source.Status {
		case "MAPPED":
			if source.Reason != "" || source.BodySHA256 == "" || source.Start < 0 || source.End <= source.Start ||
				source.DecodedStart < 0 || source.DecodedEnd <= source.DecodedStart || source.Expression == "" {
				return fmt.Errorf("mapped return expression declaration is incomplete")
			}
		case "UNMAPPED":
			if source.Reason == "" || source.Start != 0 || source.End != 0 {
				return fmt.Errorf("unmapped return expression declaration claims a range")
			}
		default:
			return fmt.Errorf("return expression declaration mapping differs")
		}
	}
	return nil
}
