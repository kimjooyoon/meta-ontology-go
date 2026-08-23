package main

import (
	"fmt"
	"io"
	"reflect"

	metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.corpus == "" || cfg.concept == "" || cfg.expectedHead == "" { return fmt.Errorf("root, corpus, concept, and expected-head are required") }
	if (cfg.output == "") == (cfg.check == "") { return fmt.Errorf("exactly one of output or check is required") }
	target := cfg.output; if cfg.check != "" { target = cfg.check }
	if err := requireExternal(cfg.root, target); err != nil { return err }
	corpus, err := readStrict[metalsp.Corpus](cfg.corpus)
	if err != nil { return err }
	concept, err := readConcept(cfg.concept)
	if err != nil { return err }
	report := metalsp.Evaluate(cfg.expectedHead, corpus, concept)
	if cfg.check != "" {
		existing, readErr := readStrict[metalsp.Report](cfg.check)
		if readErr != nil { return readErr }
		if err := metalsp.Validate(existing, cfg.expectedHead); err != nil { return err }
		if !reflect.DeepEqual(existing, report) { return fmt.Errorf("toolchain lsp replay mismatch") }
	} else if err := writeExclusive(cfg.output, report); err != nil { return err }
	fmt.Fprintf(stdout, "toolchain-lsp: decision=%s resolution=%s cases=%d/%d capabilities=%d/8 writes=%d\n",
		report.Decision, report.Resolution, report.Summary.CasesSatisfied, report.Summary.CasesTotal,
		report.Summary.AdvertisedCapabilities, report.RepositoryWrites)
	if report.Decision != metalsp.DecisionPass { return fmt.Errorf("%s", report.Reason) }
	return nil
}
