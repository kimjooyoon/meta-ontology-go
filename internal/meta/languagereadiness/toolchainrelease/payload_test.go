package toolchainrelease

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const payloadTestHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func payloadTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"go.mod", "LICENSE", payloadSourcePath,
		"licenses/gooo-decision-runtime/LICENSE", "licenses/gooo-jev/LICENSE"} {
		raw, err := os.ReadFile(filepath.Join("../../../..", filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestGoooReleasePayloadBindsDirectModulesAndTextOrigins(t *testing.T) {
	root := payloadTestRoot(t)
	for _, target := range Targets() {
		payload, err := buildReleasePayload(root, target, payloadTestHead)
		if err != nil {
			t.Fatal(target.ID, err)
		}
		if payload.policy.Members.BinaryPath != binaryName(target) ||
			payload.policy.SourceSHA256 != payloadDigest(payloadSource) ||
			payload.policy.ProjectionSHA256 == "" || len(payload.members) != 4 ||
			!bytes.Contains(payload.members[3].data, []byte(payload.policy.Members.JevRef)) {
			t.Fatalf("source-owned payload differs for %s: %+v", target.ID, payload.policy)
		}
	}
	for _, row := range []struct {
		name string
		edit func([]byte) []byte
	}{
		{"changed license", func(b []byte) []byte { return append(b, '\n') }},
		{"missing license", func([]byte) []byte { return nil }},
	} {
		t.Run(row.name, func(t *testing.T) {
			name := filepath.Join(root, "licenses/gooo-jev/LICENSE")
			old, _ := os.ReadFile(name)
			defer os.WriteFile(name, old, 0o644)
			if err := os.WriteFile(name, row.edit(old), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := buildReleasePayload(root, Targets()[0], payloadTestHead); err == nil {
				t.Fatal("changed source text retained release authority")
			}
		})
	}
	modulePath := filepath.Join(root, "go.mod")
	module, _ := os.ReadFile(modulePath)
	if err := os.WriteFile(modulePath, append(module, []byte("\nrequire example.org/extra v1.0.0\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := buildReleasePayload(root, Targets()[0], payloadTestHead); err == nil {
		t.Fatal("new direct module kept the old archive obligation")
	}
	if err := os.WriteFile(modulePath, module, 0o644); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, filepath.FromSlash(payloadSourcePath))
	if err := os.WriteFile(sourcePath, append(bytes.Clone(payloadSource), '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readPayloadPolicy(root, Targets()[0]); err == nil {
		t.Fatal("changed Gooo policy retained executable origin")
	}
}

func TestReleaseArchiveReaderChecksEveryMember(t *testing.T) {
	root := payloadTestRoot(t)
	for _, target := range Targets() {
		t.Run(target.ID, func(t *testing.T) {
			payload, err := buildReleasePayload(root, target, payloadTestHead)
			if err != nil {
				t.Fatal(err)
			}
			members, evidence := payload.withBinary([]byte("native binary fixture"), payloadTestHead)
			path := filepath.Join(t.TempDir(), archiveName(target))
			if err := createArchive(path, target, members); err != nil {
				t.Fatal(err)
			}
			if err := verifyArchiveMembers(path, target, evidence.Members); err != nil {
				t.Fatal(err)
			}
			digest, size, err := digestFile(path)
			if err != nil {
				t.Fatal(err)
			}
			receipt := PlatformReceipt{Schema: PlatformReceiptSchema, Platform: target,
				HeadSHA: payloadTestHead, Binary: ReplayEvidence{Name: binaryName(target),
					Digest: evidence.Members[0].Digest, Bytes: evidence.Members[0].Bytes},
				Archive: ReplayEvidence{Name: archiveName(target), Digest: digest, Bytes: size}, Payload: evidence}
			if _, ok := verifyReceiptArchive(root, filepath.Dir(path), receipt); !ok {
				t.Fatal("complete archive was refused")
			}
			changed := append([]archiveMember(nil), members...)
			changed[2].data = []byte("changed license")
			if err := createArchive(path, target, changed); err != nil {
				t.Fatal(err)
			}
			receipt.Archive.Digest, receipt.Archive.Bytes, _ = digestFile(path)
			if _, ok := verifyReceiptArchive(root, filepath.Dir(path), receipt); ok {
				t.Fatal("changed license passed with a recomputed archive checksum")
			}
			for label, altered := range map[string][]archiveMember{
				"missing license": append(append([]archiveMember(nil), members[:2]...), members[3:]...),
				"missing notice":  members[:len(members)-1],
				"extra":           append(append([]archiveMember(nil), members...), members[1]),
			} {
				if err := createArchive(path, target, altered); err != nil {
					t.Fatal(err)
				}
				if err := verifyArchiveMembers(path, target, evidence.Members); err == nil {
					t.Fatal(label, "member inventory passed")
				}
			}
			changed = append([]archiveMember(nil), members...)
			changed[1].evidence.Mode = "0755"
			if err := createArchive(path, target, changed); err != nil {
				t.Fatal(err)
			}
			if err := verifyArchiveMembers(path, target, evidence.Members); err == nil {
				t.Fatal("changed document mode passed")
			}
			if err := createArchive(path, target, members); err != nil {
				t.Fatal(err)
			}
			receipt.Archive.Digest, receipt.Archive.Bytes, _ = digestFile(path)
			receipt.Payload.Members[1].SourceURL = strings.ReplaceAll(
				receipt.Payload.Members[1].SourceURL, "github.com", "example.org")
			if _, ok := verifyReceiptArchive(root, filepath.Dir(path), receipt); ok {
				t.Fatal("changed source URL passed")
			}
		})
	}
}
