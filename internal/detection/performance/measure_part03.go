package performance

// MeasureAll observes all supplied stages in canonical pipeline order and
// evaluates each stage against its own budget. The input slice is not changed.
func MeasureAll(specs []StageSpec, config Config) (Report, error) {
	if err := validateSpecs(specs); err != nil {
		return Report{}, err
	}
	ordered := orderedSpecs(specs)
	report := Report{Observations: make([]Observation, 0, len(ordered))}
	for _, spec := range ordered {
		observation, err := Measure(spec, config)
		if err != nil {
			return report, err
		}
		report.Observations = append(report.Observations, observation)
		report.Violations = append(report.Violations, DetectBudget(observation)...)
	}
	return report, nil
}
