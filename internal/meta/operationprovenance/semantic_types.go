package operationprovenance

const relationDenominator = 4

type metricSpec struct {
	ID, Family, PriorClaim string
	Producer, Consumer     string
	MetaOperation          string
	EvidencePath           string
	DependsOn              []string
}

type scenarioSpec struct {
	ID, RemoveRelation, Dependency, Reason string
}

type edge struct{ From, To, Kind string }
type fixture struct {
	ID, Mutation string
	Nodes        map[string]string
	Edges        []edge
	Metrics      []metricSpec
	Artifacts    map[string][]RelationObservation
}
type relation struct{ Kind, From, To string }

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
