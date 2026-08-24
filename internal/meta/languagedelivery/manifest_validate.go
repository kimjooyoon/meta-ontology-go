package languagedelivery

import "fmt"

func manifestIssues(manifest SourceManifest, evidence EvidenceSet, head string) map[SourceName]string {
	issues := map[SourceName]string{}
	global := manifestGlobalIssue(manifest, head)
	seen := map[SourceName]bool{}
	for _, source := range sourceOrder {
		entry, ok := manifest.entry(source)
		if global != "" {
			issues[source] = global
			continue
		}
		if !ok || seen[source] {
			issues[source] = "SOURCE_MANIFEST_ENTRY_MISSING_OR_DUPLICATE"
			continue
		}
		seen[source] = true
		if entry.ArtifactName != artifactPrefix(source)+head {
			issues[source] = "SOURCE_ARTIFACT_HEAD_UNBOUND"
			continue
		}
		if entry.ArtifactID <= 0 || entry.ArchiveDigest == "" {
			issues[source] = "SOURCE_ARTIFACT_IDENTITY_UNKNOWN"
			continue
		}
		if entry.ReportDigest != digestBytes(evidence.Bytes(source)) {
			issues[source] = "SOURCE_REPORT_DIGEST_MISMATCH"
		}
	}
	return issues
}

func manifestGlobalIssue(manifest SourceManifest, head string) string {
	if manifest.Schema != ManifestSchema || manifest.SubjectSHA != head {
		return "SOURCE_MANIFEST_SUBJECT_UNKNOWN"
	}
	if manifest.WorkflowRunID <= 0 || manifest.WorkflowName != "Transformation effect ledger" {
		return "SOURCE_WORKFLOW_IDENTITY_UNKNOWN"
	}
	if manifest.WorkflowDecision != "success" || manifest.RepositoryWrites != 0 {
		return "SOURCE_WORKFLOW_NOT_EXACT"
	}
	if len(manifest.Artifacts) != len(sourceOrder) {
		return fmt.Sprintf("SOURCE_MANIFEST_CARDINALITY_%d", len(manifest.Artifacts))
	}
	return ""
}
