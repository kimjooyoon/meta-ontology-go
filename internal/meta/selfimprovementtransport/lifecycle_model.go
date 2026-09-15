package selfimprovementtransport

const (
	LifecycleReceiptSchema     = "gooo/github-actions-artifact-lifecycle-receipt/v1"
	LifecycleMetricID          = "gooo.metric.github-actions-artifact-lifecycle.gal5.v1"
	LifecycleDenominatorID     = "gooo/github-actions-artifact-lifecycle-denominator/v1"
	LifecycleEffectNoEffect    = "NO_EFFECT"
	LifecycleResolutionExact   = "EXACT"
	LifecycleResolutionUnknown = "UNKNOWN"
	LifecycleClaimOpen         = "OPEN"
	LifecycleClaimDischarged   = "DISCHARGED"
	lifecycleFixedStepTotal    = 5
)

type lifecycleDefinition struct {
	Stage, Step, Class, ProofChoice, MetaOperation, SuccessReason string
	ClaimID, Statement                                            string
}

var lifecycleDefinitions = []lifecycleDefinition{
	{"LOCATE", "read-artifact-metadata", "DRIVER", "FOUNDATION",
		"meta.artifact.lifecycle.read-metadata:v1", "ARTIFACT_METADATA_READ",
		"gooo.claim.artifact-lifecycle.metadata-read.v1", "artifact metadata is readable"},
	{"LOCATE", "resolve-artifact", "OUTCOME", "COHERENCE",
		"meta.artifact.lifecycle.resolve-artifact:v1", "ARTIFACT_RESOLVED",
		"gooo.claim.artifact-lifecycle.resolved.v1", "the exact run artifact is resolved"},
	{"LOCATE", "validate-artifact-metadata", "GUARDRAIL", "REGRESSION",
		"meta.artifact.lifecycle.validate-metadata:v1", "ARTIFACT_METADATA_VALID",
		"gooo.claim.artifact-lifecycle.metadata-valid.v1", "artifact metadata is valid and live"},
	{"TRANSPORT", "download-archive", "OUTCOME", "COHERENCE",
		"meta.artifact.lifecycle.download-archive:v1", "ARCHIVE_DOWNLOADED",
		"gooo.claim.artifact-lifecycle.archive-downloaded.v1", "artifact archive bytes are available"},
	{"TRANSPORT", "verify-archive-digest", "GUARDRAIL", "REGRESSION",
		"meta.artifact.lifecycle.verify-archive-digest:v1", "ARCHIVE_DIGEST_VERIFIED",
		"gooo.claim.artifact-lifecycle.archive-digest.v1", "archive bytes match immutable metadata"},
}

type ArtifactLifecycleInput struct {
	Selection           ArtifactSelectionInput
	RunLookupExit       int
	ArtifactsLookupExit int
}
