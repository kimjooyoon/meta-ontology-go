package languagedelivery

type ProfileReceipt struct {
	Schema            string         `json:"schema"`
	SubjectSHA        string         `json:"subject_sha"`
	Decision          string         `json:"decision"`
	Resolution        string         `json:"resolution"`
	Summary           ProfileSummary `json:"summary"`
	RepositoryWrites  int            `json:"repository_writes"`
	MutationAuthority bool           `json:"mutation_authority"`
}

type ProfileSummary struct {
	Profiles int `json:"profiles"`
	Samples  int `json:"samples"`
	Unknowns int `json:"unknowns"`
	Effects  struct {
		RepositoryWrites  int  `json:"repository_writes"`
		MutationAuthority bool `json:"mutation_authority"`
	} `json:"effects"`
}

func inspectProfile(data []byte, head string, receipt *ProfileReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceProfile, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceProfile, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.RepositoryWrites + receipt.Summary.Effects.RepositoryWrites
	observation.MutationAuthority = receipt.MutationAuthority || receipt.Summary.Effects.MutationAuthority
	if receipt.SubjectSHA != head || receipt.Summary.Unknowns != 0 {
		return headUnknown(observation)
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/language-profile-experiment-report/v1")
}
