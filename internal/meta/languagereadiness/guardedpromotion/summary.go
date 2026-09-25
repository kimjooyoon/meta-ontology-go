package guardedpromotion

func Summarize(source Source, coordinates []Coordinate) Summary {
	summary := Summary{
		Total:                        len(coordinates),
		ValidPredecessors:            source.ValidCandidates,
		AmbiguousCandidates:          source.AmbiguousCandidates,
		RepositoryWrites:             source.RepositoryWrites,
		RepositoryMutationAuthorized: source.RepositoryMutationAuthorized,
	}
	for _, coordinate := range coordinates {
		switch coordinate.Status {
		case statusSatisfied:
			summary.Satisfied++
		case statusUnresolved:
			summary.Unresolved++
		default:
			summary.NotSatisfied++
		}
	}
	if summary.Total > 0 {
		summary.ReadinessBPS = summary.Satisfied * 10000 / summary.Total
	}
	summary.ReadinessPromotionAuthorized = summary.Satisfied == summary.Total &&
		summary.Unresolved == 0 && source.AmbiguousCandidates == 0
	return summary
}
