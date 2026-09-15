package toolchainrelease

import (
	"fmt"
	"reflect"
)

type Corpus struct {
	Schema    string       `json:"schema"`
	Version   int          `json:"version"`
	Toolchain string       `json:"toolchain"`
	Targets   []Target     `json:"targets"`
	Cases     []CorpusCase `json:"cases"`
}

func DecodeCorpus(raw []byte) (Corpus, string, error) {
	corpus, err := decodeStrict[Corpus](raw)
	if err != nil {
		return Corpus{}, "", err
	}
	if err := validateCorpus(corpus); err != nil {
		return Corpus{}, "", err
	}
	digest, err := digestValue(corpus)
	return corpus, digest, err
}

func validateCorpus(corpus Corpus) error {
	if corpus.Schema != CorpusSchema || corpus.Version != 1 {
		return fmt.Errorf("TOOLCHAIN_RELEASE_CORPUS_SCHEMA_DRIFT")
	}
	if corpus.Toolchain != ExpectedToolchain {
		return fmt.Errorf("TOOLCHAIN_RELEASE_TOOLCHAIN_DRIFT")
	}
	if !reflect.DeepEqual(corpus.Targets, targetRegistry) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_TARGET_REGISTRY_DRIFT")
	}
	if !reflect.DeepEqual(corpus.Cases, expectedCases()) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_CASE_REGISTRY_DRIFT")
	}
	return nil
}
