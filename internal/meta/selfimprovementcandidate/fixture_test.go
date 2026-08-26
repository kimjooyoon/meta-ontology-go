package selfimprovementcandidate

import (
	"encoding/json"
	"strings"
	"testing/fstest"
)

const candidateContractPath = "examples/self-improvement/candidate.gooo"

const candidateContract = `package selfimprovement
namespace selfimprovement

entity ReadOnlyImprovementInput id "gooo://self-improvement/entity/read-only-improvement-input"
entity MissingCapability id "gooo://self-improvement/entity/missing-capability"
entity NonExecutingImprovementCandidate id "gooo://self-improvement/entity/non-executing-improvement-candidate"

activity SelectMissingCapability(ReadOnlyImprovementInput) -> MissingCapability
activity ProposeNonExecutingCandidate(MissingCapability) -> NonExecutingImprovementCandidate
`

func validRepository() fstest.MapFS {
	return fstest.MapFS{candidateContractPath: &fstest.MapFile{Data: []byte(candidateContract), Mode: 0o444}}
}

func sourceBytes(source sourceObservation) []byte {
	raw, _ := json.Marshal(source)
	return raw
}

func fixtureDigest() string { return digestBytes([]byte("fixture")) }

func fixtureSHA(character string) string { return strings.Repeat(character, 40) }
