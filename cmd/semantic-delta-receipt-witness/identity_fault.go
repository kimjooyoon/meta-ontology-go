package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

const (
	identityFaultSchema   = "gooo/semantic-delta-claim-identity-fault/v1"
	identityFaultID       = "raw-only-stable-id-recreation"
	identityFaultTarget   = "alternate-observation"
	identityFaultMutation = "stable_id_only"
	identityFaultRule     = "replace each alternate stable_id with gooo://semantic-delta/identity-fault/claim/<sha256(rule|alternate-before-raw-digest|alternate-after-raw-digest|original-stable-id)>"
)

type identityFaultArtifact struct {
	Schema   string `json:"schema"`
	FaultID  string `json:"fault_id"`
	Target   string `json:"target"`
	Mutation string `json:"mutation"`
	Rule     string `json:"rule"`
}

type identityFaultEvidence struct {
	ArtifactPath   string `json:"artifact_path"`
	ArtifactBytes  int    `json:"artifact_bytes"`
	ArtifactDigest string `json:"artifact_digest"`
	FaultID        string `json:"fault_id"`
	Target         string `json:"target"`
	Mutation       string `json:"mutation"`
	Rule           string `json:"rule"`
}

func readIdentityFaultArtifact(path string) (identityFaultArtifact, identityFaultEvidence, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("read identity fault artifact: %w", err)
	}
	var artifact identityFaultArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("decode identity fault artifact: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("identity fault artifact has trailing data")
	}
	if artifact.Schema != identityFaultSchema || artifact.FaultID != identityFaultID || artifact.Target != identityFaultTarget || artifact.Mutation != identityFaultMutation || artifact.Rule != identityFaultRule {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("identity fault artifact contract mismatch")
	}
	return artifact, identityFaultEvidence{ArtifactPath: path, ArtifactBytes: len(raw), ArtifactDigest: bytesDigest(raw), FaultID: artifact.FaultID, Target: artifact.Target, Mutation: artifact.Mutation, Rule: artifact.Rule}, nil
}

func mutateProducerIdentityFault(records []producer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact) []producer.ClaimIdentityRecord {
	result := append([]producer.ClaimIdentityRecord(nil), records...)
	for index := range result {
		result[index].StableID = producerFaultStableID(artifact.Rule, observation, result[index].StableID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StableID < result[j].StableID })
	return result
}

func mutateConsumerIdentityFault(records []consumer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact) []consumer.ClaimIdentityRecord {
	result := append([]consumer.ClaimIdentityRecord(nil), records...)
	for index := range result {
		result[index].StableID = consumerFaultStableID(artifact.Rule, observation, result[index].StableID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StableID < result[j].StableID })
	return result
}

func producerFaultStableID(rule string, observation evolutionSourcePair, original string) string {
	material := strings.Join([]string{rule, observation.BeforeRawDigest, observation.AfterRawDigest, original}, "|")
	digest := strings.TrimPrefix(bytesDigest([]byte(material)), "sha256:")
	return "gooo://semantic-delta/identity-fault/claim/" + digest
}

func consumerFaultStableID(rule string, observation evolutionSourcePair, original string) string {
	material := strings.Join([]string{rule, observation.BeforeRawDigest, observation.AfterRawDigest, original}, "|")
	digest := strings.TrimPrefix(bytesDigest([]byte(material)), "sha256:")
	return "gooo://semantic-delta/identity-fault/claim/" + digest
}

func producerFaultOnlyStableIDChanges(original, faulted []producer.ClaimIdentityRecord) bool {
	return identityFaultSnapshotsOnlyStableIDChanges(producerRecordSnapshots(original), producerRecordSnapshots(faulted))
}

func consumerFaultOnlyStableIDChanges(original, faulted []consumer.ClaimIdentityRecord) bool {
	return identityFaultSnapshotsOnlyStableIDChanges(consumerRecordSnapshots(original), consumerRecordSnapshots(faulted))
}

func identityFaultSnapshotsOnlyStableIDChanges(original, faulted []claimIdentityRecordSnapshot) bool {
	if len(original) == 0 || len(original) != len(faulted) {
		return false
	}
	left := append([]claimIdentityRecordSnapshot(nil), original...)
	right := append([]claimIdentityRecordSnapshot(nil), faulted...)
	sort.Slice(left, func(i, j int) bool { return identityFaultSnapshotKey(left[i]) < identityFaultSnapshotKey(left[j]) })
	sort.Slice(right, func(i, j int) bool { return identityFaultSnapshotKey(right[i]) < identityFaultSnapshotKey(right[j]) })
	for index := range left {
		if identityFaultSnapshotKey(left[index]) != identityFaultSnapshotKey(right[index]) || left[index].StableID == right[index].StableID {
			return false
		}
	}
	return true
}

func identityFaultSnapshotKey(record claimIdentityRecordSnapshot) string {
	return strings.Join([]string{record.Kind, record.RelationRole, record.NormalizedProposition, record.PropositionDigest, record.TargetAddress, record.TargetAddressDigest, record.PreservationOf, record.BeforeSourcePath, record.AfterSourcePath, record.EvidenceBeforeRawDigest, record.EvidenceAfterRawDigest, record.EvidenceBeforeSemanticDigest, record.EvidenceAfterSemanticDigest}, "\x00")
}
