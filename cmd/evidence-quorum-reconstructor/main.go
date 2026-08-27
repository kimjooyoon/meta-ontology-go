package main

import (
	"flag"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumchannel"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

type options struct{ source, policy, head, dependencies, out string }

func main() {
	var value options
	flags := flag.NewFlagSet("evidence-quorum-reconstructor", flag.ExitOnError)
	flags.StringVar(&value.source, "source", "", "raw Gooo source")
	flags.StringVar(&value.policy, "policy", "", "raw Gooo quorum policy")
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.dependencies, "dependencies", "", "reconstructor dependency manifest")
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
	semantic, err := evidencequorumpolicy.SemanticDigest(value.source, source)
	if err != nil {
		return err
	}
	executable, err := evidencequorumchannel.ExecutableDigest(os.Args[0])
	if err != nil {
		return err
	}
	dependencies, err := evidencequorumchannel.ReadDependencies(value.dependencies)
	if err != nil {
		return err
	}
	receipt := evidencequorumchannel.NewReceipt(policy, evidencequorumwire.CurrentEvidence, "raw-source-semantic-reconstructor", value.head,
		value.source, evidencequorumwire.DigestBytes(source), semantic, executable, dependencies, "RAW_SOURCE_SEMANTIC_RECONSTRUCTED")
	receipt.ObservationDigest = evidencequorumwire.ObservationDigest(receipt)
	return evidencequorumchannel.Write(value.out, receipt)
}
