package proofchoicealgebra

type Item struct {
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	Statement        string   `json:"statement,omitempty"`
	PriorState       string   `json:"prior_state,omitempty"`
	Choice           string   `json:"choice,omitempty"`
	Resolution       string   `json:"resolution"`
	ObservationState string   `json:"observation_state"`
	Observations     []string `json:"observations"`
	Numerator        int      `json:"numerator,omitempty"`
	Denominator      int      `json:"denominator,omitempty"`
	EvidenceDigest   string   `json:"evidence_digest"`
	Provenance       []string `json:"provenance"`
}

type ObservationSlot struct {
	ID         string   `json:"id"`
	Observed   bool     `json:"observed"`
	Provenance []string `json:"provenance"`
}

type Evidence struct {
	ClaimID              string            `json:"claim_id"`
	Subject              string            `json:"subject"`
	Route                Route             `json:"route"`
	Producer             string            `json:"producer"`
	Consumer             string            `json:"consumer"`
	State                string            `json:"state"`
	Reason               string            `json:"reason,omitempty"`
	ObservationIDs       []string          `json:"observation_ids"`
	ObservationSlots     []ObservationSlot `json:"observation_slots"`
	StableIdentity       string            `json:"stable_identity,omitempty"`
	OriginDigest         string            `json:"origin_digest,omitempty"`
	SubjectBinding       string            `json:"subject_binding,omitempty"`
	ProjectionA          string            `json:"projection_a,omitempty"`
	ProjectionB          string            `json:"projection_b,omitempty"`
	Agreement            bool              `json:"agreement,omitempty"`
	FirstArtifactDigest  string            `json:"first_artifact_digest,omitempty"`
	SecondArtifactDigest string            `json:"second_artifact_digest,omitempty"`
	ByteEqual            bool              `json:"byte_equal,omitempty"`
	SemanticEqual        bool              `json:"semantic_equal,omitempty"`
	EvidenceDigest       string            `json:"evidence_digest"`
	Provenance           []string          `json:"provenance"`
}

type Composition struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	Members        []string `json:"members"`
	Routes         []Route  `json:"routes"`
	Operator       string   `json:"operator"`
	Result         string   `json:"result"`
	EvidenceDigest string   `json:"evidence_digest"`
	Provenance     []string `json:"provenance"`
}

type Transition struct {
	Sequence       int      `json:"sequence"`
	ClaimID        string   `json:"claim_id"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	Choice         string   `json:"choice,omitempty"`
	Resolution     string   `json:"resolution"`
	Stage          string   `json:"stage"`
	Step           string   `json:"step"`
	Reason         string   `json:"reason"`
	EvidenceDigest string   `json:"evidence_digest"`
	Provenance     []string `json:"provenance"`
	Persistent     bool     `json:"persistent"`
}
