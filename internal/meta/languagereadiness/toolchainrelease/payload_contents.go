package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type archiveMember struct {
	evidence ArchiveMemberEvidence
	data     []byte
}

type releasePayload struct {
	policy  payloadPolicy
	members []archiveMember
}

func buildReleasePayload(root string, target Target, head string) (releasePayload, error) {
	policy, err := readPayloadPolicy(root, target)
	if err != nil {
		return releasePayload{}, err
	}
	if err := verifyPayloadModules(root, policy.Members); err != nil {
		return releasePayload{}, err
	}
	m := policy.Members
	rootModule := "github.com/kimjooyoon/meta-ontology-go"
	entries := []struct {
		role, name, expected, module, version, ref string
	}{
		{"root-license", m.RootLicensePath, m.RootLicenseSHA256, rootModule, head, head},
		{"runtime-license", m.RuntimeLicensePath, m.RuntimeLicenseSHA256, m.RuntimeModule, m.RuntimeVersion, m.RuntimeRef},
		{"jev-license", m.JevLicensePath, m.JevLicenseSHA256, m.JevModule, m.JevVersion, m.JevRef},
	}
	result := releasePayload{policy: policy}
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.name)))
		if err != nil || len(raw) == 0 || len(raw) > 64<<10 || payloadDigest(raw) != entry.expected {
			return releasePayload{}, fmt.Errorf("TOOLCHAIN_RELEASE_LICENSE_SOURCE_MISMATCH: %s", entry.name)
		}
		sourcePath := "LICENSE"
		if entry.role == "root-license" {
			sourcePath = entry.name
		}
		result.members = append(result.members, archiveMember{ArchiveMemberEvidence{
			Role: entry.role, Name: entry.name, Digest: entry.expected, Bytes: int64(len(raw)),
			Mode: "0644", SourcePath: entry.name, Module: entry.module, Version: entry.version,
			Ref: entry.ref, SourceURL: "https://github.com/" + strings.TrimPrefix(entry.module, "github.com/") +
				"/blob/" + entry.ref + "/" + sourcePath,
		}, raw})
	}
	notice := releaseNotice(result.members, policy, head)
	result.members = append(result.members, archiveMember{ArchiveMemberEvidence{
		Role: "notice", Name: m.NoticePath, Digest: payloadDigest(notice), Bytes: int64(len(notice)),
		Mode: "0644", SourcePath: payloadSourcePath, Module: rootModule, Version: head,
		Ref: head, SourceURL: "https://github.com/kimjooyoon/meta-ontology-go/blob/" + head + "/" + payloadSourcePath,
	}, notice})
	return result, nil
}

func releaseNotice(documents []archiveMember, policy payloadPolicy, head string) []byte {
	var out strings.Builder
	fmt.Fprintf(&out, "Gooo release document origins\ncompiler revision: %s\npayload policy: %s (%s)\n\n",
		head, payloadSourcePath, policy.SourceSHA256)
	for _, item := range documents {
		entry := item.evidence
		fmt.Fprintf(&out, "%s\nmodule: %s\nversion: %s\nrevision: %s\nsource: %s\nsha256: %s\n\n",
			entry.Name, entry.Module, entry.Version, entry.Ref, entry.SourceURL, strings.TrimPrefix(entry.Digest, "sha256:"))
	}
	return []byte(out.String())
}

func (payload releasePayload) withBinary(binary []byte, head string) ([]archiveMember, PayloadEvidence) {
	rootModule := "github.com/kimjooyoon/meta-ontology-go"
	main := archiveMember{ArchiveMemberEvidence{
		Role: "binary", Name: payload.policy.Members.BinaryPath, Digest: payloadDigest(binary),
		Bytes: int64(len(binary)), Mode: "0755", SourcePath: "cmd/gooo", Module: rootModule,
		Version: head, Ref: head, SourceURL: "https://github.com/kimjooyoon/meta-ontology-go/tree/" + head + "/cmd/gooo",
	}, binary}
	members := append([]archiveMember{main}, payload.members...)
	evidence := PayloadEvidence{SourceSHA256: payload.policy.SourceSHA256,
		ProjectionSHA256: payload.policy.ProjectionSHA256, Members: make([]ArchiveMemberEvidence, 0, len(members))}
	for _, member := range members {
		evidence.Members = append(evidence.Members, member.evidence)
	}
	return members, evidence
}

func expectedPayloadEvidence(payload releasePayload, head string,
	binary ReplayEvidence) PayloadEvidence {
	_, expected := payload.withBinary(nil, head)
	expected.Members[0].Digest, expected.Members[0].Bytes = binary.Digest, binary.Bytes
	return expected
}

func samePayloadEvidence(a, b PayloadEvidence) bool {
	return reflect.DeepEqual(a, b)
}
