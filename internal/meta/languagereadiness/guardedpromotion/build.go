package guardedpromotion

func Build(source Source) Report {
	coordinates := Coordinates(source)
	summary := Summarize(source, coordinates)
	decision, reason, resolution := Evaluate(source, coordinates, summary)
	report := Report{
		Schema: Schema, Decision: decision, Reason: reason, Resolution: resolution,
		Source: source, Summary: summary, Coordinates: coordinates,
		Indicators: Indicators(source, summary, coordinates),
		Proofs:     Proofs(coordinates),
	}
	seal(&report)
	return report
}
