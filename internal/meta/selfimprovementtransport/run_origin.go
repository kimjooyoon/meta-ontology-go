package selfimprovementtransport

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport/originpolicy"
)

// RunOrigin records the consumer policy and the API event used to select a role.
// Subject/checkout binding and payload integrity remain separate obligations.
type RunOrigin struct {
	State        string `json:"state"`
	Reason       string `json:"reason"`
	Event        string `json:"event"`
	PolicyDigest string `json:"policy_digest"`
}

func resolveRunOrigin(input evaluationInput) RunOrigin {
	reason := originpolicy.BindProducerRun(input.metadata.ProducerEvent, input.producer.SubjectSHA,
		input.producer.WorkflowSHA, input.metadata.OrchestrationHeadSHA)
	state := ResolutionRefuted
	if strings.HasSuffix(reason, "_BOUND") {
		state = ResolutionClosed
	} else if strings.HasSuffix(reason, "_UNKNOWN") || reason == "PRODUCER_EVENT_UNSUPPORTED" {
		state = ResolutionUnknown
	}
	return RunOrigin{State: state, Reason: reason, Event: input.metadata.ProducerEvent,
		PolicyDigest: originpolicy.SourceDigest()}
}

func applyRunOrigin(resolution ProvenanceResolution, origin RunOrigin) ProvenanceResolution {
	resolution.Stage, resolution.Step = "CONSUME", "bind-producer-run-origin"
	if origin.State == ResolutionRefuted {
		return refutedResolution(resolution, origin.Reason)
	}
	return unknownResolution(resolution, resolution.Stage, resolution.Step, origin.Reason,
		"ORIGIN_UNRESOLVED", []string{"producer-run-event", "producer-run-origin"})
}
