package verify

type cMetric struct {
	id, family, claim                       string
	producer, consumer, operation, evidence string
	dependsOn                               []string
}
type cScenario struct {
	id, removeRelation, dependency, reason string
}
type cRelation struct{ kind, from, to string }
type cEdge struct{ from, to, kind string }
type cFixture struct {
	id, mutation string
	nodes        map[string]string
	edges        []cEdge
	metrics      []cMetric
}
