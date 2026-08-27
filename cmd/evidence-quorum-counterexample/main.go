package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

type options struct{ mode, input, out string }

func main() {
	var value options
	flags := flag.NewFlagSet("evidence-quorum-counterexample", flag.ExitOnError)
	flags.StringVar(&value.mode, "mode", "", "duplicate, valid-conflict, invalid-conflict, or unknown")
	flags.StringVar(&value.input, "input", "", "current raw channel receipt")
	flags.StringVar(&value.out, "out", "", "synthetic receipt output")
	flags.Parse(os.Args[1:])
	if err := run(value); err != nil {
		panic(err)
	}
}

func run(value options) error {
	raw, err := os.ReadFile(value.input)
	if err != nil {
		return err
	}
	var receipt evidencequorumwire.Receipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return err
	}
	receipt.EvidenceClass = evidencequorumwire.SyntheticCounterexample
	switch value.mode {
	case "duplicate":
		receipt.Channel, receipt.Predicate = "synthetic-duplicate", "DUPLICATE_REPLICA"
	case "valid-conflict":
		receipt.Channel, receipt.Predicate = "synthetic-valid-conflict", "VALID_CONTRADICTION"
	case "invalid-conflict":
		receipt.Channel, receipt.Predicate = "synthetic-invalid-conflict", "UNSUPPORTED_CONTRADICTION"
	case "unknown":
		receipt.Channel, receipt.Predicate = "synthetic-unknown", "UNKNOWN_OBSERVATION"
	default:
		return fmt.Errorf("unknown counterexample mode %q", value.mode)
	}
	receipt.ObservationDigest = evidencequorumwire.ObservationDigest(receipt)
	receipt.Digest = ""
	receipt = evidencequorumwire.Seal(receipt)
	result, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	return os.WriteFile(value.out, append(result, '\n'), 0o644)
}
