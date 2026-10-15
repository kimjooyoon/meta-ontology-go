// Package candidatedelivery binds a workflow delivery to its logical subject.
package candidatedelivery

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

const Schema = "gooo/candidate-delivery/v1"
const ManifestSchema = "gooo/candidate-delivery-manifest/v1"
const MaxArchiveBytes = 16 << 20

type Run struct {
	ID         int64  `json:"id"`
	Attempt    int    `json:"run_attempt"`
	Head       string `json:"head_sha"`
	Event      string `json:"event"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Repository struct {
		Name string `json:"full_name"`
	} `json:"repository"`
}

type Artifact struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Digest  string `json:"digest"`
	Size    int64  `json:"size_in_bytes"`
	Expired bool   `json:"expired"`
	Run     struct {
		ID   int64  `json:"id"`
		Head string `json:"head_sha"`
	} `json:"workflow_run"`
}

func (a *Artifact) UnmarshalJSON(raw []byte) error {
	type wire Artifact
	var value wire
	var presence struct {
		Expired *bool `json:"expired"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &presence); err != nil {
		return err
	}
	if presence.Expired == nil {
		return fmt.Errorf("artifact expiration observation missing")
	}
	*a = Artifact(value)
	return nil
}

type Manifest struct {
	Schema          string `json:"schema"`
	Role            string `json:"role"`
	RunID           int64  `json:"run_id"`
	Attempt         int    `json:"run_attempt"`
	Event           string `json:"event"`
	WorkflowSHA     string `json:"workflow_sha"`
	RunHeadSHA      string `json:"run_head_sha"`
	SubjectSHA      string `json:"subject_sha"`
	CandidateDigest string `json:"candidate_bytes_digest"`
	TransportDigest string `json:"transport_bytes_digest"`
	Digest          string `json:"digest"`
}

type Receipt struct {
	Schema          string                                  `json:"schema"`
	State           string                                  `json:"state"`
	Reason          string                                  `json:"reason"`
	Unknown         *selfimprovementtransport.CausalUnknown `json:"unknown,omitempty"`
	Repository      string                                  `json:"repository"`
	Expected        Selection                               `json:"expected"`
	Run             Run                                     `json:"run"`
	Artifact        Artifact                                `json:"artifact"`
	SubjectSHA      string                                  `json:"subject_sha,omitempty"`
	PolicyDigest    string                                  `json:"policy_digest"`
	CandidateDigest string                                  `json:"candidate_bytes_digest,omitempty"`
	TransportDigest string                                  `json:"transport_bytes_digest,omitempty"`
	Digest          string                                  `json:"digest"`
}

type Bundle struct{ Candidate, Transport, Manifest []byte }
type Selection struct {
	Repository string `json:"repository"`
	RunID      int64  `json:"run_id"`
	Attempt    int    `json:"run_attempt"`
}

func Digest(raw []byte) string    { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }
func digestJSON(value any) string { raw, _ := json.Marshal(value); return Digest(raw) }

func Metadata(receipt Receipt) selfimprovementcandidate.ArtifactMetadata {
	return selfimprovementcandidate.ArtifactMetadata{Repository: receipt.Repository,
		SubjectSHA: receipt.SubjectSHA, RunID: receipt.Run.ID, RunAttempt: receipt.Run.Attempt,
		ArtifactID: receipt.Artifact.ID, ArtifactName: receipt.Artifact.Name,
		ArchiveDigest: receipt.Artifact.Digest, SizeBytes: receipt.Artifact.Size, Expired: receipt.Artifact.Expired}
}
