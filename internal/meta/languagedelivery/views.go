package languagedelivery

func buildViews(results []ObligationResult, receiptDecision, receiptResolution, factsDigest string) []AudienceView {
	views := make([]AudienceView, 0, len(audienceOrder))
	for index, audience := range audienceOrder {
		visible := visibleResults(results, index)
		value := coordinates(visible)
		decision, resolution, _ := reportDecision(value, EffectSummary{})
		views = append(views, AudienceView{
			Audience: audience, ProjectionDecision: decision, ProjectionResolution: resolution,
			ReceiptDecision: receiptDecision, ReceiptResolution: receiptResolution,
			Coordinates: value, VisibleUnknowns: value.Unknown,
			HiddenUnknowns: coordinates(results).Unknown - value.Unknown,
			EvidenceDigest: factsDigest,
		})
	}
	return views
}

func visibleResults(results []ObligationResult, maxAudience int) []ObligationResult {
	allowed := map[Audience]bool{}
	for index, audience := range audienceOrder {
		if index <= maxAudience {
			allowed[audience] = true
		}
	}
	visible := make([]ObligationResult, 0, (maxAudience+1)*12)
	for _, result := range results {
		if allowed[result.Audience] {
			visible = append(visible, result)
		}
	}
	return visible
}
