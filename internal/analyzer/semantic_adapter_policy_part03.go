package analyzer

func (m RelationMapping) allowsOrigin(origin ObservationOrigin) bool {
	if len(m.AllowedOrigins) == 0 {
		return true
	}
	for _, allowed := range m.AllowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}
func (p MappingPolicy) lookup(relation Relation) (RelationMapping, bool) {
	mapping, ok := p.mappings[relation]
	return mapping, ok
}
func knownAnalyzerRelation(relation Relation) bool {
	switch relation {
	case RelationInvokes, RelationUses, RelationGenerates, RelationReferences:
		return true
	default:
		return false
	}
}
