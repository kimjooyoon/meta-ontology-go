package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumchannel"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

type options struct{ source, policy, artifact, manifest, head, dependencies, out string }
type generatedManifest struct {
	Status          string `json:"status"`
	SemanticDigest  string `json:"semantic_digest"`
	GeneratedDigest string `json:"generated_digest"`
}

func main() {
	var value options
	flags := flag.NewFlagSet("evidence-quorum-artifact-observer", flag.ExitOnError)
	flags.StringVar(&value.source, "source", "", "raw Gooo source")
	flags.StringVar(&value.policy, "policy", "", "raw Gooo quorum policy")
	flags.StringVar(&value.artifact, "artifact", "", "generated artifact")
	flags.StringVar(&value.manifest, "manifest", "", "generated artifact manifest")
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.dependencies, "dependencies", "", "observer dependency manifest")
	flags.StringVar(&value.out, "out", "", "channel receipt output")
	flags.Parse(os.Args[1:])
	if err := run(value); err != nil {
		panic(err)
	}
}

func run(value options) error {
	source, err := os.ReadFile(value.source)
	if err != nil {
		return err
	}
	policySource, err := os.ReadFile(value.policy)
	if err != nil {
		return err
	}
	policy, err := evidencequorumpolicy.Parse(value.policy, policySource)
	if err != nil {
		return err
	}
	artifact, err := os.ReadFile(value.artifact)
	if err != nil {
		return err
	}
	manifestRaw, err := os.ReadFile(value.manifest)
	if err != nil {
		return err
	}
	var manifest generatedManifest
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
		return err
	}
	if manifest.Status != "pass" || manifest.SemanticDigest == "" {
		return fmt.Errorf("generated artifact manifest is not a passing observation")
	}
	semantic := "sha256:" + manifest.SemanticDigest
	executable, err := evidencequorumchannel.ExecutableDigest(os.Args[0])
	if err != nil {
		return err
	}
	dependencies, err := evidencequorumchannel.ReadDependencies(value.dependencies)
	if err != nil {
		return err
	}
	receipt := evidencequorumchannel.NewReceipt(policy, evidencequorumwire.CurrentEvidence, "generated-artifact-observer", value.head,
		value.source, evidencequorumwire.DigestBytes(source), semantic, executable, dependencies, "GENERATED_ARTIFACT_OBSERVED")
	receipt.GeneratedArtifactDigest = evidencequorumwire.DigestJSON(struct {
		Artifact string `json:"artifact"`
		Manifest string `json:"manifest"`
	}{evidencequorumwire.DigestBytes(artifact), evidencequorumwire.DigestBytes(manifestRaw)})
	receipt.ObservationDigest = evidencequorumwire.ObservationDigest(receipt)
	return evidencequorumchannel.Write(value.out, receipt)
}
