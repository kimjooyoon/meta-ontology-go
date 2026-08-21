package closure_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
)

const fixtureRepository = "kimjooyoon/meta-ontology-go"
const fixtureSHA = "0123456789abcdef0123456789abcdef01234567"

func fixtureDigest(character string) string {
	return "sha256:" + strings.Repeat(character, 64)
}

func bytesDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fixtureInput() closure.Input {
	source := []byte("package metricmetaprogram\n")
	program := fixtureProgram(source)
	programJSON, _ := json.Marshal(program)
	verificationJSON, _ := json.Marshal(fixtureVerification(program))
	return closure.Input{
		Repository: fixtureRepository, SubjectSHA: fixtureSHA, RunID: 17, RunAttempt: 1,
		Artifact: closure.ArtifactIdentity{
			ID: 19, Name: "metric-meta-program-" + fixtureSHA,
			Digest: strings.Repeat("e", 64),
			URL: "https://github.com/" + fixtureRepository + "/actions/runs/17/artifacts/19",
		},
		ProgramJSON: programJSON, Source: source, VerificationJSON: verificationJSON,
	}
}
