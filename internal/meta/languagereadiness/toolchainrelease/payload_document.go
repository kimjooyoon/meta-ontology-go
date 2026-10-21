package toolchainrelease

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const payloadDocumentPath = "docs/language/release-payload-origins.md"
const payloadTableBegin = "<!-- RELEASE-PAYLOAD-MODULES:BEGIN -->"
const payloadTableEnd = "<!-- RELEASE-PAYLOAD-MODULES:END -->"

type PayloadDocumentProjection struct {
	Schema             string         `json:"schema"`
	Changed            bool           `json:"changed"`
	Written            bool           `json:"written"`
	PolicySourceSHA256 string         `json:"policy_source_sha256"`
	ModuleSHA256       string         `json:"module_sha256"`
	InputSHA256        string         `json:"input_sha256"`
	OutputSHA256       string         `json:"output_sha256"`
	Provenance         map[string]any `json:"provenance,omitempty"`
}

// ProjectPayloadModuleDocument derives the current-module table from the Gooo
// release policy. Check mode makes no write or generation claim.
func ProjectPayloadModuleDocument(root string, write bool) (PayloadDocumentProjection, error) {
	moduleBefore, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return PayloadDocumentProjection{}, err
	}
	payload, err := buildReleasePayload(root, Targets()[0], "document-projection")
	if err != nil {
		return PayloadDocumentProjection{}, err
	}
	moduleAfter, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || !bytes.Equal(moduleBefore, moduleAfter) {
		return PayloadDocumentProjection{}, fmt.Errorf("release module source changed during projection")
	}
	policy := payload.policy
	path := filepath.Join(root, filepath.FromSlash(payloadDocumentPath))
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return PayloadDocumentProjection{}, fmt.Errorf("release payload document is not a regular file")
	}
	before, err := os.ReadFile(path)
	if err != nil {
		return PayloadDocumentProjection{}, err
	}
	after, err := replacePayloadModuleTable(before, renderPayloadModuleTable(policy.Members))
	if err != nil {
		return PayloadDocumentProjection{}, err
	}
	result := PayloadDocumentProjection{Schema: "gooo/release-payload-document-projection/v1",
		Changed: !bytes.Equal(before, after), PolicySourceSHA256: policy.SourceSHA256,
		ModuleSHA256: payloadDigest(moduleBefore), InputSHA256: payloadDigest(before), OutputSHA256: payloadDigest(after)}
	if !write || !result.Changed {
		return result, nil
	}
	if err := verifyPayloadDocumentPreimages(root, policy, result); err != nil {
		return PayloadDocumentProjection{}, err
	}
	if err := writePayloadDocument(path, after, info.Mode().Perm()); err != nil {
		return PayloadDocumentProjection{}, err
	}
	result.Written = true
	result.Provenance = payloadDocumentProvenance(result, policy.Members)
	return result, nil
}

func verifyPayloadDocumentPreimages(root string, policy payloadPolicy, result PayloadDocumentProjection) error {
	for _, input := range []struct{ path, sha string }{
		{payloadSourcePath, policy.SourceSHA256}, {"go.mod", result.ModuleSHA256},
		{policy.Members.RuntimeLicensePath, policy.Members.RuntimeLicenseSHA256},
		{policy.Members.JevLicensePath, policy.Members.JevLicenseSHA256},
		{payloadDocumentPath, result.InputSHA256},
	} {
		current, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(input.path)))
		if readErr != nil || payloadDigest(current) != input.sha {
			return fmt.Errorf("release document preimage changed: %s", input.path)
		}
	}
	return nil
}

func writePayloadDocument(path string, after []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".release-payload-doc-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(after); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), path)
}

func renderPayloadModuleTable(m requiredMembers) string {
	return fmt.Sprintf("| Module | Pinned source | License text SHA-256 |\n| --- | --- | --- |\n"+
		"| `%s` | `%s` / `%s` | `%s` |\n| `%s` | `%s` / `%s` | `%s` |",
		path.Base(m.RuntimeModule), m.RuntimeVersion, m.RuntimeRef,
		strings.TrimPrefix(m.RuntimeLicenseSHA256, "sha256:"),
		path.Base(m.JevModule), m.JevVersion, m.JevRef,
		strings.TrimPrefix(m.JevLicenseSHA256, "sha256:"))
}

func replacePayloadModuleTable(document []byte, table string) ([]byte, error) {
	begin, end := []byte(payloadTableBegin), []byte(payloadTableEnd)
	if bytes.Count(document, begin) != 1 || bytes.Count(document, end) != 1 {
		return nil, fmt.Errorf("release payload document has no unique module table markers")
	}
	start := bytes.Index(document, begin) + len(begin)
	stop := bytes.Index(document, end)
	if stop < start {
		return nil, fmt.Errorf("release payload module table markers are reversed")
	}
	after := append(bytes.Clone(document[:start]), '\n')
	after = append(after, table...)
	after = append(after, '\n')
	return append(after, document[stop:]...), nil
}

func payloadDocumentProvenance(result PayloadDocumentProjection, members requiredMembers) map[string]any {
	sourceID := payloadDocumentEntityID("source", payloadSourcePath, result.PolicySourceSHA256)
	moduleID := payloadDocumentEntityID("module", "go.mod", result.ModuleSHA256)
	runtimeID := payloadDocumentEntityID("license", members.RuntimeLicensePath, members.RuntimeLicenseSHA256)
	jevID := payloadDocumentEntityID("license", members.JevLicensePath, members.JevLicenseSHA256)
	inputID := payloadDocumentEntityID("document-input", payloadDocumentPath, result.InputSHA256)
	outputID := payloadDocumentEntityID("document-output", payloadDocumentPath, result.OutputSHA256)
	activityID := "urn:gooo:release-payload:document-projection:" + result.OutputSHA256
	return map[string]any{"@context": map[string]string{"prov": "http://www.w3.org/ns/prov#",
		"gooo": "urn:gooo:release-payload:vocab:"}, "@graph": []any{
		map[string]any{"@id": sourceID, "@type": "prov:Entity", "gooo:path": payloadSourcePath,
			"gooo:content_sha256": result.PolicySourceSHA256},
		map[string]any{"@id": moduleID, "@type": "prov:Entity", "gooo:path": "go.mod",
			"gooo:content_sha256": result.ModuleSHA256},
		map[string]any{"@id": runtimeID, "@type": "prov:Entity", "gooo:path": members.RuntimeLicensePath,
			"gooo:content_sha256": members.RuntimeLicenseSHA256},
		map[string]any{"@id": jevID, "@type": "prov:Entity", "gooo:path": members.JevLicensePath,
			"gooo:content_sha256": members.JevLicenseSHA256},
		map[string]any{"@id": inputID, "@type": "prov:Entity", "gooo:path": payloadDocumentPath,
			"gooo:content_sha256": result.InputSHA256},
		map[string]any{"@id": activityID, "@type": "prov:Activity", "prov:used": []any{
			map[string]string{"@id": sourceID}, map[string]string{"@id": moduleID},
			map[string]string{"@id": runtimeID}, map[string]string{"@id": jevID},
			map[string]string{"@id": inputID}}},
		map[string]any{"@id": outputID, "@type": "prov:Entity", "gooo:path": payloadDocumentPath,
			"gooo:content_sha256": result.OutputSHA256,
			"prov:wasGeneratedBy": map[string]string{"@id": activityID},
			"prov:wasDerivedFrom": []any{map[string]string{"@id": sourceID}, map[string]string{"@id": inputID}}},
	}}
}

func payloadDocumentEntityID(role, path, sha string) string {
	return "urn:gooo:release-payload:" + role + ":" + strings.TrimPrefix(payloadDigest([]byte(path)), "sha256:") +
		":" + strings.TrimPrefix(sha, "sha256:")
}
