package candidatedelivery

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport/originpolicy"
)

func Produce(identity Manifest, name string, bundle Bundle) (Manifest, error) {
	identity.Schema, identity.Role = ManifestSchema, originpolicy.CandidateArtifactRole(name)
	identity.CandidateDigest, identity.TransportDigest = Digest(bundle.Candidate), Digest(bundle.Transport)
	if identity.Role == "" || identity.RunID <= 0 || identity.Attempt <= 0 || !validSHA(identity.WorkflowSHA) {
		return Manifest{}, fmt.Errorf("candidate delivery producer identity invalid")
	}
	reason := originpolicy.BindProducerRun(identity.Event, identity.SubjectSHA, identity.WorkflowSHA, identity.RunHeadSHA)
	if !validSHA(identity.RunHeadSHA) || (reason != "WORKFLOW_ORIGIN_BOUND" && reason != "SUBJECT_ORIGIN_BOUND") {
		return Manifest{}, fmt.Errorf("candidate delivery producer origin invalid: %s", reason)
	}
	if _, err := validateBundle(identity, bundle); err != nil {
		return Manifest{}, err
	}
	identity.Digest = ""
	identity.Digest = digestJSON(identity)
	return identity, nil
}

func Consume(r Receipt, archive []byte) (Receipt, Bundle) {
	sealed := closeReceipt(r, r.State, r.Reason)
	if sealed.Digest != r.Digest || r.PolicyDigest != originpolicy.SourceDigest() || r.State != "LOCATED" {
		return closeReceipt(r, "REFUTED", "DELIVERY_SELECTION_MISMATCH"), Bundle{}
	}
	if len(archive) == 0 {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_ARCHIVE_MISSING"), Bundle{}
	}
	if len(archive) > MaxArchiveBytes || Digest(archive) != r.Artifact.Digest || int64(len(archive)) != r.Artifact.Size {
		return closeReceipt(r, "REFUTED", "DELIVERY_ARCHIVE_MISMATCH"), Bundle{}
	}
	bundle, err := readBundle(archive)
	if err != nil {
		return closeReceipt(r, "REFUTED", "DELIVERY_ARCHIVE_INVALID"), Bundle{}
	}
	if len(bundle.Candidate) == 0 || len(bundle.Transport) == 0 || len(bundle.Manifest) == 0 {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_MEMBER_MISSING"), Bundle{}
	}
	var manifest Manifest
	if json.Unmarshal(bundle.Manifest, &manifest) != nil {
		return closeReceipt(r, "REFUTED", "DELIVERY_MANIFEST_INVALID"), Bundle{}
	}
	unsigned := manifest
	unsigned.Digest = ""
	if manifest.Schema != ManifestSchema || manifest.Digest != digestJSON(unsigned) ||
		manifest.Role != originpolicy.CandidateArtifactRole(r.Artifact.Name) ||
		manifest.RunID != r.Run.ID || manifest.Attempt != r.Run.Attempt || manifest.Event != r.Run.Event ||
		manifest.RunHeadSHA != r.Run.Head {
		return closeReceipt(r, "REFUTED", "DELIVERY_MANIFEST_MISMATCH"), Bundle{}
	}
	if manifest.WorkflowSHA == "" {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_WORKFLOW_ORIGIN_UNKNOWN"), Bundle{}
	}
	if !validSHA(manifest.WorkflowSHA) {
		return closeReceipt(r, "REFUTED", "DELIVERY_WORKFLOW_ORIGIN_INVALID"), Bundle{}
	}
	reason := originpolicy.BindProducerRun(r.Run.Event, manifest.SubjectSHA, manifest.WorkflowSHA, r.Run.Head)
	if reason != "WORKFLOW_ORIGIN_BOUND" && reason != "SUBJECT_ORIGIN_BOUND" {
		state := "REFUTED"
		if stringsUnknown(reason) {
			state = "UNKNOWN"
		}
		return closeReceipt(r, state, reason), Bundle{}
	}
	if _, err := validateBundle(manifest, bundle); err != nil {
		state := "REFUTED"
		if stringsUnknown(err.Error()) {
			state = "UNKNOWN"
		}
		return closeReceipt(r, state, err.Error()), Bundle{}
	}
	r.SubjectSHA, r.CandidateDigest, r.TransportDigest = manifest.SubjectSHA, manifest.CandidateDigest, manifest.TransportDigest
	return closeReceipt(r, "CLOSED", "CANDIDATE_DELIVERY_BOUND"), bundle
}

func validateBundle(manifest Manifest, bundle Bundle) (string, error) {
	var candidate selfimprovementcandidate.Report
	var transport selfimprovementtransport.Report
	if json.Unmarshal(bundle.Candidate, &candidate) != nil || json.Unmarshal(bundle.Transport, &transport) != nil {
		return "", fmt.Errorf("DELIVERY_MEMBER_INVALID")
	}
	reason := originpolicy.BindCandidateSubject(candidate.SubjectSHA, transport.SubjectSHA, manifest.SubjectSHA)
	if reason != "CANDIDATE_SUBJECT_BOUND" {
		return "", fmt.Errorf("%s", reason)
	}
	if !validSHA(manifest.SubjectSHA) {
		return "", fmt.Errorf("DELIVERY_SUBJECT_INVALID")
	}
	if manifest.CandidateDigest != Digest(bundle.Candidate) || manifest.TransportDigest != Digest(bundle.Transport) ||
		selfimprovementcandidate.Validate(candidate, manifest.SubjectSHA, candidate.SourceWorkflowRunID) != nil ||
		candidate.Decision != selfimprovementcandidate.DecisionProposed || candidate.SourceWorkflowRunID <= 0 ||
		selfimprovementtransport.ValidateReport(transport) != nil ||
		candidate.SourceFileDigest != transport.SourceObservationDigest {
		return "", fmt.Errorf("DELIVERY_PAYLOAD_BINDING_MISMATCH")
	}
	if transport.ProvenanceState == selfimprovementtransport.ResolutionUnknown {
		return "", fmt.Errorf("DELIVERY_TRANSPORT_ORIGIN_UNKNOWN")
	}
	if selfimprovementtransport.CheckReadOnly(transport) != nil {
		return "", fmt.Errorf("DELIVERY_TRANSPORT_ORIGIN_MISMATCH")
	}
	return reason, nil
}

func readBundle(raw []byte) (Bundle, error) {
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return Bundle{}, err
	}
	if len(z.File) > 128 {
		return Bundle{}, fmt.Errorf("delivery member population oversized")
	}
	files := map[string][]byte{}
	for _, entry := range z.File {
		if entry.Name != "output/first.json" && entry.Name != "transport/receipt.json" && entry.Name != "output/delivery.json" {
			continue
		}
		if _, found := files[entry.Name]; found {
			return Bundle{}, fmt.Errorf("duplicate delivery member")
		}
		stream, err := entry.Open()
		if err != nil {
			return Bundle{}, err
		}
		data, readErr := io.ReadAll(io.LimitReader(stream, (8<<20)+1))
		closeErr := stream.Close()
		if readErr != nil || closeErr != nil || len(data) > 8<<20 {
			return Bundle{}, fmt.Errorf("delivery member unavailable or oversized")
		}
		files[entry.Name] = data
	}
	return Bundle{Candidate: files["output/first.json"], Transport: files["transport/receipt.json"], Manifest: files["output/delivery.json"]}, nil
}
