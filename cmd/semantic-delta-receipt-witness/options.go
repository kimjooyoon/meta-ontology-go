package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	caseID, subjectSHA, observedCheckoutSHA, before, after, effectsBefore, effectsAfter, output                                      string
	oldExpectation, newExpectation, persistenceManifest                                                                              string
	semanticClaimManifest, identityFault, persistenceBefore, persistenceAfter, persistenceAlternateBefore, persistenceAlternateAfter string
	tamperMatrix, evolution, persistenceProbe                                                                                        bool
	identityFaultCardinality                                                                                                         bool
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	set := flag.NewFlagSet("semantic-delta-receipt-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	var result options
	set.StringVar(&result.caseID, "case", "", "fixed case ID or suite")
	set.StringVar(&result.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&result.observedCheckoutSHA, "observed-checkout-sha", "", "observed checkout SHA")
	set.StringVar(&result.before, "before", "", "before .gooo source")
	set.StringVar(&result.after, "after", "", "after .gooo source")
	set.StringVar(&result.effectsBefore, "effects-before", "", "pre-execution workspace content snapshot")
	set.StringVar(&result.effectsAfter, "effects-after", "", "post-execution workspace content snapshot")
	set.StringVar(&result.output, "output", "", "JSON output path")
	set.BoolVar(&result.tamperMatrix, "tamper-matrix", false, "write the fixed consumer tamper matrix evidence")
	set.BoolVar(&result.evolution, "evolution", false, "reconstruct claim identity expectation evolution")
	set.StringVar(&result.semanticClaimManifest, "semantic-claim-manifest", "", "reconstruct semantic claim delta fixtures from raw source")
	set.BoolVar(&result.persistenceProbe, "persistence-probe", false, "compare two raw-source persistence observations")
	set.BoolVar(&result.identityFaultCardinality, "identity-fault-cardinality", false, "exercise the fixed identity-fault semantic-slot cardinality cases")
	set.StringVar(&result.identityFault, "identity-fault", "", "apply a separately declared identity fault to the alternate observation")
	set.StringVar(&result.persistenceBefore, "persistence-before", "", "persistence baseline before source")
	set.StringVar(&result.persistenceAfter, "persistence-after", "", "persistence baseline after source")
	set.StringVar(&result.persistenceAlternateBefore, "persistence-alternate-before", "", "persistence alternate before source")
	set.StringVar(&result.persistenceAlternateAfter, "persistence-alternate-after", "", "persistence alternate after source")
	set.StringVar(&result.oldExpectation, "old-expectation", "examples/semantic-delta-receipt/claim-transition-expectations-v2.json", "old claim identity artifact")
	set.StringVar(&result.newExpectation, "new-expectation", "examples/semantic-delta-receipt/claim-transition-expectations.json", "new claim identity artifact")
	set.StringVar(&result.persistenceManifest, "persistence-manifest", "examples/semantic-delta-receipt/claim-identity-persistence-manifest.json", "baseline/alternate source-pair manifest")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if result.subjectSHA == "" || result.output == "" {
		return options{}, fmt.Errorf("--subject-sha and --output are required")
	}
	if result.tamperMatrix && (result.caseID != "" || result.before != "" || result.after != "" || result.evolution || result.semanticClaimManifest != "" || result.persistenceProbe || result.identityFault != "") {
		return options{}, fmt.Errorf("--tamper-matrix cannot be combined with another witness mode")
	}
	if result.evolution && (result.caseID != "" || result.before != "" || result.after != "" || result.tamperMatrix || result.semanticClaimManifest != "" || result.persistenceProbe || result.identityFault != "") {
		return options{}, fmt.Errorf("--evolution cannot be combined with another witness mode")
	}
	if result.semanticClaimManifest != "" && (result.caseID != "" || result.before != "" || result.after != "" || result.persistenceProbe || result.identityFault != "") {
		return options{}, fmt.Errorf("--semantic-claim-manifest cannot be combined with another witness mode")
	}
	if result.persistenceProbe && (result.caseID != "" || result.before != "" || result.after != "" || result.semanticClaimManifest != "") {
		return options{}, fmt.Errorf("--persistence-probe cannot be combined with another witness mode")
	}
	if result.identityFault != "" && !result.persistenceProbe {
		return options{}, fmt.Errorf("--identity-fault requires --persistence-probe")
	}
	if result.identityFaultCardinality && (!result.persistenceProbe || result.identityFault == "") {
		return options{}, fmt.Errorf("--identity-fault-cardinality requires --persistence-probe and --identity-fault")
	}
	if result.evolution {
		return result, nil
	}
	if result.semanticClaimManifest != "" {
		return result, nil
	}
	if result.persistenceProbe {
		if result.persistenceBefore == "" || result.persistenceAfter == "" || result.persistenceAlternateBefore == "" || result.persistenceAlternateAfter == "" {
			return options{}, fmt.Errorf("--persistence-probe requires all four persistence source paths")
		}
		return result, nil
	}
	if result.caseID == "" && !result.tamperMatrix && (result.before == "" || result.after == "") {
		return options{}, fmt.Errorf("--case or both --before and --after are required")
	}
	if result.caseID != "" && result.caseID != "suite" && result.before != "" {
		return options{}, fmt.Errorf("--case cannot be combined with --before or --after")
	}
	return result, nil
}
