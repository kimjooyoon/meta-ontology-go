package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func validateEvidence(bundle evidence) error {
	if bundle.Schema != evidenceSchema || bundle.Repository == "" || bundle.Event == "" || !validEventRef(bundle.Event, bundle.EventRef) || bundle.CheckoutRef != bundle.HeadSHA || bundle.BaseRef == "" || bundle.RunID <= 0 || bundle.RunAttempt <= 0 || bundle.Toolchain == "" || !bundle.SlotPreservation || !bundle.NoWriteOutsideGenerated {
		return fmt.Errorf("CI evidence metadata is incomplete")
	}
	if !validSHA(bundle.BaseSHA) || !validSHA(bundle.HeadSHA) || !validSHA(bundle.WorkflowSHA) || bundle.BaseSHA == bundle.HeadSHA {
		return fmt.Errorf("CI evidence metadata contains invalid revisions")
	}
	if len(bundle.Jobs) != len(canonicalJobs) {
		return fmt.Errorf("CI evidence must contain all six canonical jobs")
	}
	if !artifactProvenanceBound(bundle.ArtifactProvenance, bundle.BaseSHA, bundle.HeadSHA) {
		return fmt.Errorf("CI evidence artifact provenance is missing or unbound")
	}
	seenIDs := make(map[int64]bool, len(bundle.Jobs))
	for index, job := range bundle.Jobs {
		if job.Name != canonicalJobs[index] || job.ID <= 0 || seenIDs[job.ID] || job.Status != "completed" || job.Conclusion != "success" || job.HeadSHA != bundle.HeadSHA || job.RunID != bundle.RunID || job.RunAttempt != bundle.RunAttempt {
			return fmt.Errorf("CI evidence job %q is incomplete or mismatched", job.Name)
		}
		seenIDs[job.ID] = true
	}
	for _, digest := range []string{bundle.Digests.SourceSHA256, bundle.Digests.IRSHA256, bundle.Digests.GeneratorFixtureSHA256, bundle.Digests.GeneratedOutputSHA256, bundle.Digests.SourceMapSHA256, bundle.Digests.PolicySHA256, bundle.Digests.ToolchainSHA256, bundle.Digests.BundleSHA256} {
		if len(digest) != 64 {
			return fmt.Errorf("CI evidence digest is missing or malformed")
		}
	}
	recorded := bundle.Digests.BundleSHA256
	bundle.Digests.BundleSHA256 = ""
	payload, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	if digestBytes(payload) != recorded {
		return fmt.Errorf("CI evidence bundle digest mismatch")
	}
	return nil
}

func verifyFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var bundle evidence
	if err := json.Unmarshal(data, &bundle); err != nil {
		return err
	}
	return validateEvidence(bundle)
}
