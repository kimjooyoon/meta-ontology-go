package semanticdelta

// factIdentity is deliberately structured instead of being a delimiter-joined
// string. Semantic values are opaque at this boundary; a delimiter must never
// be able to make two different triples look like the same fact.
type factIdentity struct {
	subject   string
	predicate string
	object    string
}

func factIdentityOf(fact Fact) factIdentity {
	return factIdentity{subject: fact.Subject, predicate: fact.Predicate, object: fact.Object}
}
