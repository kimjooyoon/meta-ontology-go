package languageconcept

const (
	ArtifactSchema    = "gooo/language-concept-artifact/v1"
	CatalogSourcePath = "internal/meta/languageconcept"
)

type BindingObservation struct {
	Paths       int    `json:"paths"`
	Files       int    `json:"files"`
	Bytes       int64  `json:"bytes"`
	Missing     int    `json:"missing"`
	Unsupported int    `json:"unsupported"`
	Digest      string `json:"digest"`
}

type Artifact struct {
	Schema             string             `json:"schema"`
	Decision           string             `json:"decision"`
	Reason             string             `json:"reason"`
	Producer           string             `json:"producer"`
	Consumer           string             `json:"consumer"`
	MetaOperation      string             `json:"meta_operation"`
	CatalogSource      string             `json:"catalog_source"`
	CatalogDigest      string             `json:"catalog_digest"`
	Report             Report             `json:"report"`
	ReplayReportDigest string             `json:"replay_report_digest"`
	ReplayEqual        bool               `json:"replay_equal"`
	Bindings           BindingObservation `json:"bindings"`
	RepositoryWrites   int                `json:"repository_writes"`
	ArtifactDigest     string             `json:"artifact_digest"`
}

func (artifact Artifact) Ready() bool {
	return artifact.Decision == "PASS" &&
		artifact.Reason == "LANGUAGE_CONCEPT_ARTIFACT_READY"
}
