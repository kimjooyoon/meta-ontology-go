package predecessorselection

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorbinding"
)

func bindCandidate(candidate Candidate) (Selection, []byte, []byte, error) {
	readinessRaw, err := base64.StdEncoding.DecodeString(candidate.ReadinessPayloadBase64)
	if err != nil || len(readinessRaw) == 0 {
		return Selection{}, nil, nil, fmt.Errorf("readiness payload unavailable")
	}
	var readiness readinessartifact.Receipt
	if json.Unmarshal(readinessRaw, &readiness) != nil || readinessartifact.Validate(readiness) != nil {
		return Selection{}, nil, nil, fmt.Errorf("readiness payload invalid")
	}
	bindingRaw, err := base64.StdEncoding.DecodeString(candidate.BindingPayloadBase64)
	if err != nil || len(bindingRaw) == 0 {
		return Selection{}, nil, nil, fmt.Errorf("binding payload unavailable")
	}
	var binding predecessorbinding.Report
	if json.Unmarshal(bindingRaw, &binding) != nil ||
		predecessorbinding.Validate(binding, candidate.HeadSHA) != nil {
		return Selection{}, nil, nil, fmt.Errorf("binding payload invalid")
	}
	if readiness.HeadSHA != candidate.HeadSHA || binding.HeadSHA != candidate.HeadSHA {
		return Selection{}, nil, nil, fmt.Errorf("payload head mismatch")
	}
	summary := readiness.Snapshot.Summary
	reference := readinessartifact.FoundationBaseline(readinessartifact.BaselineReference{
		RunID:          candidate.RunID,
		ArtifactName:   candidate.ReadinessArtifactName,
		HeadSHA:        candidate.HeadSHA,
		FileSHA256:     digestBytes(readinessRaw),
		ArtifactDigest: readiness.ArtifactDigest,
		SnapshotDigest: readiness.Snapshot.Digest,
		RegistryDigest: readiness.Snapshot.RegistryDigest,
		Completed:      summary.Completed,
		Total:          summary.Total,
		BasisPoints:    summary.ReadinessBPS,
	})
	selection := Selection{
		RunID:                 candidate.RunID,
		RunAttempt:            candidate.RunAttempt,
		WorkflowConclusion:    candidate.Conclusion,
		ProducerJobID:         candidate.ProducerJobID,
		ProducerJobRunAttempt: candidate.ProducerJobRunAttempt,
		ProducerJobName:       candidate.ProducerJobName,
		ProducerJobConclusion: candidate.ProducerJobConclusion,
		ReadinessArtifactID:   candidate.ReadinessArtifactID,
		BindingArtifactID:     candidate.BindingArtifactID,
		Baseline:              reference,
	}
	return selection, readinessRaw, bindingRaw, nil
}
