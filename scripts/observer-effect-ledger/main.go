package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/observereffect"
)

type options struct {
	mode, root, source                        string
	topologyRoot                              string
	ledger, observationReceipt, effectReceipt string
	allowIntentionalWrite                     bool
}

func main() {
	var opts options
	flag.StringVar(&opts.mode, "mode", "observe", "observe, unknown, or violate")
	flag.StringVar(&opts.root, "root", ".", "root whose effects are measured")
	flag.StringVar(&opts.topologyRoot, "topology-root", "", "root whose workflow trigger topology is audited")
	flag.StringVar(&opts.source, "source", "", "the actual .gooo source")
	flag.StringVar(&opts.ledger, "ledger", "", "observer-effect ledger output")
	flag.StringVar(&opts.observationReceipt, "observation-receipt", "", "observation receipt output")
	flag.StringVar(&opts.effectReceipt, "effect-receipt", "", "observer-effect receipt output")
	flag.BoolVar(&opts.allowIntentionalWrite, "allow-intentional-write", false, "explicitly permit the counterexample write")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "observer-effect-ledger:", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if opts.source == "" || opts.ledger == "" || opts.observationReceipt == "" || opts.effectReceipt == "" {
		return fmt.Errorf("source and all three output paths are required")
	}
	if opts.mode != "observe" && opts.mode != "unknown" && opts.mode != "violate" {
		return fmt.Errorf("unsupported mode %q", opts.mode)
	}
	if opts.mode == "violate" && !opts.allowIntentionalWrite {
		return fmt.Errorf("violate mode requires -allow-intentional-write")
	}
	report, observationReceipt, effectReceipt, err := observereffect.Build(observereffect.BuildOptions{
		Mode: opts.mode, Root: opts.root, SourcePath: opts.source,
		TopologyRoot:          opts.topologyRoot,
		AllowIntentionalWrite: opts.allowIntentionalWrite,
	})
	if err != nil {
		return err
	}
	if err := writeJSON(opts.ledger, report); err != nil {
		return err
	}
	if err := writeJSON(opts.observationReceipt, observationReceipt); err != nil {
		return err
	}
	if err := writeJSON(opts.effectReceipt, effectReceipt); err != nil {
		return err
	}
	fmt.Printf("observer-effect-ledger: decision=%s resolution=%s metrics=%d/%d repository_writes=%d\n",
		report.Decision, report.Resolution, report.Metrics.Satisfied,
		report.Metrics.FixedDenominator, report.RepositoryWrites)
	return nil
}

func writeJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	payload = append(payload, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
