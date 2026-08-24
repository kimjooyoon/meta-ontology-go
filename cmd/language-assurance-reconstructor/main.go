package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type config struct { root, subjectSHA, input, output string }

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.subjectSHA, "subject-sha", "", "exact evaluated commit sha")
	flag.StringVar(&cfg.input, "input", "", "raw transaction input")
	flag.StringVar(&cfg.output, "output", "", "reconstruction receipt outside the repository")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.subjectSHA == "" || cfg.input == "" || cfg.output == "" { return fmt.Errorf("root, subject-sha, input, and output are required") }
	if err := requireExternalOutput(cfg.root, cfg.output); err != nil { return err }
	input, err := readTransaction(cfg.input)
	if err != nil { return err }
	receipt, err := reconstruct(cfg.subjectSHA, input)
	if err != nil { return err }
	if err := writeReceipt(cfg.output, receipt); err != nil { return err }
	_, err = fmt.Fprintf(stdout, "language-assurance-reconstructor: verifier=%s subject=%s candidate=%s raw=%s\n", receipt.VerifierID, receipt.SubjectSHA, receipt.Observation.CandidateDecision, receipt.RawTransactionDigest)
	return err
}
