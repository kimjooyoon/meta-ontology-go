package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumchannel"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

type options struct{ receipt, source, policy, head, executable, dependencies, out string }

func main() {
	var value options
	flags := flag.NewFlagSet("evidence-quorum-source-channel", flag.ExitOnError)
	flags.StringVar(&value.receipt, "receipt", "", "actual cmd/gooo source-execution receipt")
	flags.StringVar(&value.source, "source", "", "raw Gooo source")
	flags.StringVar(&value.policy, "policy", "", "raw Gooo quorum policy")
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.executable, "source-executable", "", "cmd/gooo executable")
	flags.StringVar(&value.dependencies, "dependencies", "", "source executable dependency manifest")
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
	raw, err := os.ReadFile(value.receipt)
	if err != nil {
		return err
	}
	var sourceReceipt sourceExecutionReceipt
	if err := json.Unmarshal(raw, &sourceReceipt); err != nil {
		return err
	}
	semantic, err := evidencequorumpolicy.SemanticDigest(value.source, source)
	if err != nil {
		return err
	}
	executable, err := evidencequorumchannel.ExecutableDigest(value.executable)
	if err != nil {
		return err
	}
	dependencies, err := evidencequorumchannel.ReadDependencies(value.dependencies)
	if err != nil {
		return err
	}
	receipt := evidencequorumchannel.NewReceipt(policy, evidencequorumwire.CurrentEvidence, "gooo-source-execution", value.head,
		value.source, evidencequorumwire.DigestBytes(source), semantic, executable, dependencies, "SOURCE_ACTIVITY_EXECUTED")
	receipt.SourceExecutionReceiptDigest = sourceReceipt.Digest
	receipt.SourceExecutionReceipt = json.RawMessage(raw)
	receipt.ObservationDigest = evidencequorumwire.ObservationDigest(receipt)
	return evidencequorumchannel.Write(value.out, receipt)
}

type sourceExecutionReceipt struct {
	Digest string `json:"digest"`
}
