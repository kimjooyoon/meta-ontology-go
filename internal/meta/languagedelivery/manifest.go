package languagedelivery

type SourceManifest struct {
	Schema           string          `json:"schema"`
	SubjectSHA       string          `json:"subject_sha"`
	WorkflowRunID    int64           `json:"workflow_run_id"`
	WorkflowName     string          `json:"workflow_name"`
	WorkflowDecision string          `json:"workflow_decision"`
	Artifacts        []ManifestEntry `json:"artifacts"`
	RepositoryWrites int             `json:"repository_writes"`
}

type ManifestEntry struct {
	Source        SourceName `json:"source"`
	ArtifactID    int64      `json:"artifact_id"`
	ArtifactName  string     `json:"artifact_name"`
	ArchiveDigest string     `json:"archive_digest"`
	ReportDigest  string     `json:"report_digest"`
}

func (manifest SourceManifest) entry(name SourceName) (ManifestEntry, bool) {
	for _, entry := range manifest.Artifacts {
		if entry.Source == name {
			return entry, true
		}
	}
	return ManifestEntry{}, false
}

func artifactPrefix(name SourceName) string {
	switch name {
	case SourceUserJourney:
		return "user-journey-scorecard-"
	case SourceConformance:
		return "toolchain-conformance-"
	case SourceLSP:
		return "toolchain-lsp-"
	case SourceRelease:
		return "toolchain-cross-platform-release-"
	case SourceExecution:
		return "language-source-execution-"
	case SourceProfile:
		return "language-profile-"
	case SourceReadiness:
		return "language-readiness-artifact-"
	default:
		return ""
	}
}
