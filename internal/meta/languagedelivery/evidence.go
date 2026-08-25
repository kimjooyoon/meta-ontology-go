package languagedelivery

type EvidenceSet struct {
	UserJourney []byte
	Conformance []byte
	LSP         []byte
	Release     []byte
	Execution   []byte
	Profile     []byte
	Readiness   []byte
}

func (set EvidenceSet) Bytes(source SourceName) []byte {
	switch source {
	case SourceUserJourney:
		return set.UserJourney
	case SourceConformance:
		return set.Conformance
	case SourceLSP:
		return set.LSP
	case SourceRelease:
		return set.Release
	case SourceExecution:
		return set.Execution
	case SourceProfile:
		return set.Profile
	case SourceReadiness:
		return set.Readiness
	default:
		return nil
	}
}

type JourneyReceipt struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Source     struct {
		ExpectedHeadSHA string `json:"expected_head_sha"`
	} `json:"source"`
	Journeys          []JourneyEvidence   `json:"journeys"`
	Indicators        []UpstreamIndicator `json:"indicators"`
	RepositoryWrites  int                 `json:"repository_writes"`
	MutationAuthority bool                `json:"mutation_authority"`
}

type JourneyEvidence struct {
	ID             string `json:"ID"`
	Samples        int    `json:"Samples"`
	Successful     int    `json:"Successful"`
	OutputReplay   bool   `json:"OutputReplay"`
	EnvelopePassed bool   `json:"EnvelopePassed"`
}

type UpstreamIndicator struct {
	ID        string `json:"ID"`
	Satisfied bool   `json:"Satisfied"`
}
