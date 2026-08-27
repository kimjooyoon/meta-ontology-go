package proofchoicealgebra

func summarize(bundle Bundle) Summary {
	summary := Summary{Items: len(bundle.Items), Transitions: len(bundle.Transitions), FixedDenominator: FixedDenominator}
	for _, item := range bundle.Items {
		switch item.Kind {
		case Claim:
			summary.Claims++
		case Metric:
			summary.Metrics++
		}
		if item.Choice.Valid() {
			summary.ChoicesExplicit++
		}
	}
	for _, transition := range bundle.Transitions {
		if transition.Persistent {
			summary.PersistentTransitions++
		}
	}
	if summary.Items > 0 {
		summary.ChoiceCoverageBPS = summary.ChoicesExplicit * 10000 / summary.Items
	}
	return summary
}

func countUnknowns(bundle Bundle) int {
	count := 0
	for _, item := range bundle.Items {
		if !item.Choice.Valid() || metadataUnknown(item.Producer, item.Consumer, item.MetaOperation, item.Stage, item.Step, item.Reason) {
			count++
		}
	}
	for _, transition := range bundle.Transitions {
		if !transition.Choice.Valid() || metadataUnknown(transition.Producer, transition.Consumer, transition.MetaOperation, transition.Stage, transition.Step, transition.Reason) {
			count++
		}
	}
	return count
}
