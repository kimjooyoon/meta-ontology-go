package bodycodegen

import "slices"

func consistentCriteriaRound(r CriteriaQuestionRound, space int) bool {
	if r.Candidates < 1 || r.Candidates > space || r.Unobserved != space-r.Candidates || len(r.Evaluations) != r.Candidates {
		return false
	}
	var surviving, unavailable, rejected []uint16
	for i, c := range r.Evaluations {
		if c.Mask != uint16(i) {
			return false
		}
		switch c.Status {
		case "SATISFIED":
			surviving = append(surviving, c.Mask)
		case "UNKNOWN":
			unavailable = append(unavailable, c.Mask)
		case "UNSATISFIED":
			rejected = append(rejected, c.Mask)
		default:
			return false
		}
	}
	return slices.Equal(surviving, r.SurvivingMasks) && slices.Equal(unavailable, r.UnavailableMasks) && slices.Equal(rejected, r.RejectedMasks)
}

func criteriaExplanationOutput(q *CriteriaQuestionReceipt, index int, mask uint16) (CriteriaQuestionOutput, bool) {
	if index < 0 || index >= len(q.Outputs) {
		return CriteriaQuestionOutput{}, false
	}
	o := q.Outputs[index]
	return o, o.Mask == mask && o.Error == "" && o.Value != ""
}

func criteriaExplanationBinding(q *CriteriaQuestionReceipt, id string) *RecordCriterionBinding {
	var found *RecordCriterionBinding
	for _, binding := range q.Bindings {
		if binding.ID != id {
			continue
		}
		if found != nil || binding.ActivityID == "" || binding.GeneratedSHA256 == "" {
			return nil
		}
		copy := binding
		found = &copy
	}
	return found
}
