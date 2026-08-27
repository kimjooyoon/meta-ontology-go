package verify

func cloneArtifacts(input map[string][]relationObservation) map[string][]relationObservation {
	output := map[string][]relationObservation{}
	for id, observations := range input {
		output[id] = append([]relationObservation(nil), observations...)
	}
	return output
}

func observedLink(metric cMetric, observation relationObservation) cRelation {
	link := relationFor(metric, observation.Relation)
	if observation.Relation == "PRODUCES" || observation.Relation == "OPERATES" {
		link.from = observation.ObservedEndpoint
	}
	if observation.Relation == "CONSUMES" || observation.Relation == "EVIDENCED_BY" {
		link.to = observation.ObservedEndpoint
	}
	return link
}

func relationFor(metric cMetric, kind string) cRelation {
	return map[string]cRelation{"PRODUCES": {"PRODUCES", metric.producer, "metric:" + metric.id}, "CONSUMES": {"CONSUMES", "metric:" + metric.id, metric.consumer}, "OPERATES": {"OPERATES", metric.operation, "metric:" + metric.id}, "EVIDENCED_BY": {"EVIDENCED_BY", "metric:" + metric.id, metric.evidence}}[kind]
}

func relationForID(metrics []cMetric, kind, id string) cRelation {
	for _, metric := range metrics {
		if metric.id == id {
			return relationFor(metric, kind)
		}
	}
	return cRelation{}
}
