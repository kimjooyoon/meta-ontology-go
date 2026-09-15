package workfrontier

type r4Binding struct {
	payload  string
	digest   string
	reason   string
	expected []byte
}
type r4SnapshotProjection struct {
	SchemaVersion string            `json:"schema_version"`
	Roots         []string          `json:"root_obligation_ids"`
	States        []ObligationState `json:"states"`
	Paths         []RepairPath      `json:"paths"`
	IterationUses []r4IterationUse  `json:"iteration_uses"`
}
type r4IterationUse struct {
	SCCDigest      string `json:"scc_digest"`
	IterationsUsed uint64 `json:"iterations_used"`
}
type r4PolicyProjection struct {
	SchemaVersion            string                 `json:"schema_version"`
	MinimumSelectedPressures uint32                 `json:"minimum_selected_pressures"`
	Capacity                 Capacity               `json:"capacity"`
	MaxIterations            []r4MaxIterationPolicy `json:"max_iterations"`
}
type r4MaxIterationPolicy struct {
	SCCDigest     string `json:"scc_digest"`
	MaxIterations uint64 `json:"max_iterations"`
}
type r4RegistryProjection struct {
	SchemaVersion string     `json:"schema_version"`
	Pressures     []Pressure `json:"pressures"`
}

// BindR4Payloads derives the three exact canonical projection payloads and
// their SHA-256 bindings from the normalized R4 input.
func BindR4Payloads(input R4Input) (R4Input, error) {
	input = normalizeR4Input(input)
	projections, err := r4ProjectionBytes(input)
	if err != nil {
		return R4Input{}, err
	}
	input.SnapshotPayload, input.SnapshotDigest = r4PayloadBinding(projections.snapshot)
	input.PolicyPayload, input.PolicyDigest = r4PayloadBinding(projections.policy)
	input.RegistryPayload, input.RegistryDigest = r4PayloadBinding(projections.registry)
	return input, nil
}

type r4ProjectionSet struct {
	snapshot []byte
	policy   []byte
	registry []byte
}
