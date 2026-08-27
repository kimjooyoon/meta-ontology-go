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
	artifacts    map[string][]relationObservation
}

type artifactEnvelope struct {
	Kind     string `json:"kind"`
	MetricID string `json:"metric_id"`
	Endpoint string `json:"endpoint"`
	Output   string `json:"output"`
	Reads    string `json:"reads"`
	Source   string `json:"source"`
	Status   string `json:"status"`
	Input    string `json:"input"`
	Executed bool   `json:"executed"`
	Path     string `json:"path"`
	Payload  string `json:"payload"`
}
