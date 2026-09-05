package main

import (
	"flag"
	"log"
)

type options struct {
	mode, contractPath, v24RequestPath, v24ResolutionPath, v24VerificationPath string
	v25ContractPath, sourcePath, sourceProvenancePath, outputPath, outputDir    string
	grantRequestPath, resolutionPath                      string
	decision, decisionSource, actor, repository, event    string
	workflowRunID                                         int64
	workflowRunAttempt                                    int
	check                                                 bool
}

func main() {
	if err := run(parseOptions()); err != nil {
		log.Fatal(err)
	}
}

func parseOptions() options {
	mode := flag.String("mode", "live", "live, cases, or verify")
	contract := flag.String("contract", "examples/self-improvement-execution-grant/grant.gooo", "caller-selected grant policy")
	v24Request := flag.String("v24-request", "", "optional v24 authorization request JSON")
	v24Resolution := flag.String("v24-resolution", "", "optional v24 authorization resolution JSON")
	v24Verification := flag.String("v24-verification", "", "optional v24 authorization verification JSON")
	v25Contract := flag.String("v25-contract", "", "optional v25 pre-execution contract JSON")
	sourceProvenance := flag.String("source-provenance", "", "exact canonical source artifact provenance JSON for canonical-fixture mode")
	outputDir := flag.String("output-dir", "", "caller-owned output directory for canonical-fixture mode")
	grantRequest := flag.String("grant-request", "", "grant request JSON for verify mode")
	resolution := flag.String("resolution", "", "grant resolution JSON for verify mode")
	source := flag.String("source", "", "optional source artifact metadata JSON")
	output := flag.String("output", "", "caller-owned output artifact")
	decision := flag.String("decision", "", "optional explicit ALLOW or DENY")
	decisionSource := flag.String("decision-source", "", "workflow_dispatch or canonical-fixture")
	actor := flag.String("actor", "", "GitHub event actor evidence")
	repository := flag.String("repository", "", "GitHub repository evidence")
	event := flag.String("event", "", "GitHub event evidence")
	runID := flag.Int64("workflow-run-id", 0, "GitHub workflow run evidence")
	runAttempt := flag.Int("workflow-run-attempt", 0, "GitHub workflow run attempt evidence")
	check := flag.Bool("check", false, "validate the emitted artifact")
	flag.Parse()
	return options{mode: *mode, contractPath: *contract, v24RequestPath: *v24Request, v24ResolutionPath: *v24Resolution, v24VerificationPath: *v24Verification, v25ContractPath: *v25Contract, sourceProvenancePath: *sourceProvenance, outputDir: *outputDir, grantRequestPath: *grantRequest, resolutionPath: *resolution, sourcePath: *source, outputPath: *output, decision: *decision, decisionSource: *decisionSource, actor: *actor, repository: *repository, event: *event, workflowRunID: *runID, workflowRunAttempt: *runAttempt, check: *check}
}
