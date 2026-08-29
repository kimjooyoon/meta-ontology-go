package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffectverification"
)

func main() {
	var opts transformationeffectverification.Options
	flag.StringVar(&opts.PlanPath, "plan", "", "generation plan")
	flag.StringVar(&opts.ExecutionPath, "execution", "", "execution manifest")
	flag.StringVar(&opts.LedgerPath, "ledger", "", "transformation ledger")
	flag.StringVar(&opts.ReceiptsPath, "receipts", "", "receipt report")
	flag.StringVar(&opts.ProvenancePath, "provenance", "", "executed provenance")
	flag.StringVar(&opts.PatchPath, "patch", "", "content patch")
	flag.StringVar(&opts.RuntimePath, "runtime", "", "runtime observation")
	flag.StringVar(&opts.ExpectedHead, "expected-head", "", "expected checked-out head")
	flag.StringVar(&opts.OutputPath, "output", "", "verification report")
	flag.StringVar(&opts.Counterexample, "counterexample", "", "binding counterexample suite")
	flag.Parse()
	var report transformationeffectverification.Report
	var err error
	if opts.Counterexample == "" {
		report, err = transformationeffectverification.Verify(opts)
	} else {
		report, err = transformationeffectverification.VerifyCounterexamples(opts)
	}
	if writeErr := transformationeffectverification.WriteReport(opts.OutputPath, report); writeErr != nil {
		fmt.Fprintln(os.Stderr, "transformation-effect-verify:", writeErr)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "transformation-effect-verify:", err)
		os.Exit(1)
	}
	fmt.Printf("transformation-effect verification: decision=%s resolution=%s selected=%d bound=%d unbound=%d\n",
		report.Decision, report.Resolution, report.SelectedPlanOperations,
		report.BoundExecutorOperations, report.UnboundExecutorOperations)
}
