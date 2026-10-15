package languagesyntax

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

// SourceInventory reports source membership without replaying program cases.
type SourceInventory struct {
	Schema            string                   `json:"schema"`
	RegistryDigest    string                   `json:"registry_digest"`
	ObservationDigest string                   `json:"observation_digest"`
	Sources           []replay.FileObservation `json:"sources"`
	Unregistered      []string                 `json:"unregistered"`
	Missing           []string                 `json:"missing"`
	SemanticAdmission string                   `json:"semantic_admission"`
	NativeExecutions  int                      `json:"native_executions"`
	RepositoryWrites  int                      `json:"repository_writes"`
}

func InspectSourceInventory(repository fs.FS, registryRaw []byte) (SourceInventory, error) {
	registry, err := decodeRegistry(registryRaw)
	if err != nil {
		return SourceInventory{}, err
	}
	observed, err := replay.Observe(repository)
	if err != nil {
		return SourceInventory{}, err
	}
	extra, missing := compareCorpus(registry, observed)
	return SourceInventory{Schema: "gooo/syntax-source-inventory/v1", RegistryDigest: digestBytes(registryRaw),
		ObservationDigest: digestJSON(observed), Sources: observed, Unregistered: extra, Missing: missing,
		SemanticAdmission: "UNASSESSED"}, nil
}
