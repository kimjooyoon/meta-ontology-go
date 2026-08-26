package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root")
	jobsPath := flag.String("jobs", "ci-jobs.json", "workflow jobs JSON")
	generated := flag.String("generated", "ci-generated/first", "generated output directory")
	provenance := flag.String("provenance", "", "artifact provenance envelope")
	output := flag.String("output", "ci-evidence.json", "evidence output")
	verifyPath := flag.String("verify", "", "verify an existing evidence file")
	flag.Parse()
	var err error
	if *verifyPath != "" {
		err = verifyFile(*verifyPath)
	} else {
		err = build(*root, *jobsPath, *generated, *provenance, *output)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func build(root, jobsPath, generated, provenancePath, output string) error {
	jobsData, err := os.ReadFile(jobsPath)
	if err != nil {
		return fmt.Errorf("read workflow jobs: %w", err)
	}
	var apiJobs []apiJob
	if err := json.Unmarshal(jobsData, &apiJobs); err != nil {
		return fmt.Errorf("parse workflow jobs: %w", err)
	}
	metadata, err := readMetadata()
	if err != nil {
		return err
	}
	provenance, err := readArtifactProvenance(
		provenancePath, metadata.BaseSHA, metadata.HeadSHA)
	if err != nil {
		return err
	}
	jobs, err := normalizeJobs(apiJobs, metadata.HeadSHA, metadata.RunID, metadata.RunAttempt)
	if err != nil {
		return err
	}
	digestSet, err := computeDigests(root, generated)
	if err != nil {
		return err
	}
	bundle := evidence{Schema: evidenceSchema, Repository: metadata.Repository, Event: metadata.Event, EventRef: metadata.EventRef, CheckoutRef: metadata.CheckoutRef, BaseRef: metadata.BaseRef, BaseSHA: metadata.BaseSHA, HeadSHA: metadata.HeadSHA, RunID: metadata.RunID, RunAttempt: metadata.RunAttempt, WorkflowSHA: metadata.WorkflowSHA, Toolchain: metadata.Toolchain, SlotPreservation: os.Getenv("CI_SLOT_PRESERVATION") == "true", NoWriteOutsideGenerated: os.Getenv("CI_NO_WRITE_OUTSIDE_GENERATED") == "true", Jobs: jobs, ArtifactProvenance: provenance, Digests: digestSet}
	payload, err := json.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("marshal evidence payload: %w", err)
	}
	bundle.Digests.BundleSHA256 = digestBytes(payload)
	if err := validateEvidence(bundle); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evidence bundle: %w", err)
	}
	return os.WriteFile(output, append(encoded, '\n'), 0o644)
}
