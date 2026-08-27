package operationprovenance

func relations(metric metricSpec) []relation {
	return []relation{{"PRODUCES", metric.Producer, "metric:" + metric.ID}, {"CONSUMES", "metric:" + metric.ID, metric.Consumer}, {"OPERATES", metric.MetaOperation, "metric:" + metric.ID}, {"EVIDENCED_BY", "metric:" + metric.ID, metric.EvidencePath}}
}

func relationFor(metric metricSpec, kind string) relation {
	for _, link := range relations(metric) {
		if link.Kind == kind {
			return link
		}
	}
	return relation{}
}

func observedRelation(metric metricSpec, observation RelationObservation) relation {
	link := relationFor(metric, observation.Relation)
	if observation.ObservedEndpoint == "" {
		return relation{}
	}
	if observation.Relation == "PRODUCES" || observation.Relation == "OPERATES" {
		link.From = observation.ObservedEndpoint
	}
	if observation.Relation == "CONSUMES" || observation.Relation == "EVIDENCED_BY" {
		link.To = observation.ObservedEndpoint
	}
	return link
}

func relationForID(metrics []metricSpec, kind, metricID string) relation {
	for _, metric := range metrics {
		if metric.ID == metricID {
			return relationFor(metric, kind)
		}
	}
	return relation{}
}
