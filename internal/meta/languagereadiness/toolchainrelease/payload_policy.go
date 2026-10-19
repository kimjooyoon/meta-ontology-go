package toolchainrelease

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const payloadSourcePath = "internal/meta/languagereadiness/toolchainrelease/payload.gooo"

//go:embed payload.gooo
var payloadSource []byte

type requiredMembers struct {
	BinaryPath           string `json:"binary_path"`
	RootLicensePath      string `json:"root_license_path"`
	RootLicenseSHA256    string `json:"root_license_sha256"`
	RuntimeLicensePath   string `json:"runtime_license_path"`
	RuntimeLicenseSHA256 string `json:"runtime_license_sha256"`
	RuntimeModule        string `json:"runtime_module"`
	RuntimeVersion       string `json:"runtime_version"`
	RuntimeRef           string `json:"runtime_ref"`
	JevLicensePath       string `json:"jev_license_path"`
	JevLicenseSHA256     string `json:"jev_license_sha256"`
	JevModule            string `json:"jev_module"`
	JevVersion           string `json:"jev_version"`
	JevRef               string `json:"jev_ref"`
	NoticePath           string `json:"notice_path"`
}

type payloadPolicy struct {
	Members          requiredMembers
	SourceSHA256     string
	ProjectionSHA256 string
}

func payloadDigest(raw []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
}

func readPayloadPolicy(root string, target Target) (payloadPolicy, error) {
	current, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(payloadSourcePath)))
	if err != nil || !bytes.Equal(current, payloadSource) {
		return payloadPolicy{}, fmt.Errorf("TOOLCHAIN_RELEASE_PAYLOAD_SOURCE_STALE")
	}
	input, _ := json.Marshal([]any{struct {
		Windows bool `json:"windows"`
	}{target.GOOS == "windows"}})
	value, err := bodycodegen.EvaluateFixedValue(context.Background(), payloadSourcePath,
		current, "RequiredMembers", input)
	if err != nil {
		return payloadPolicy{}, fmt.Errorf("TOOLCHAIN_RELEASE_PAYLOAD_SOURCE_INVALID: %w", err)
	}
	if value.ActivityID != "gooo://release-payload/required-members" ||
		value.SourceSHA256 != payloadDigest(current) || value.ProjectionSHA256 == "" {
		return payloadPolicy{}, fmt.Errorf("TOOLCHAIN_RELEASE_PAYLOAD_ORIGIN_MISMATCH")
	}
	members, err := decodeStrict[requiredMembers](value.Value)
	if err != nil || !validRequiredMembers(members, target) {
		return payloadPolicy{}, fmt.Errorf("TOOLCHAIN_RELEASE_PAYLOAD_MEMBERS_INVALID")
	}
	return payloadPolicy{members, value.SourceSHA256, value.ProjectionSHA256}, nil
}

func validRequiredMembers(m requiredMembers, target Target) bool {
	paths := []string{m.BinaryPath, m.RootLicensePath, m.RuntimeLicensePath, m.JevLicensePath, m.NoticePath}
	seen := map[string]bool{}
	for _, name := range paths {
		if !fs.ValidPath(name) || name == "." || path.Clean(name) != name || seen[name] {
			return false
		}
		seen[name] = true
	}
	if m.BinaryPath != binaryName(target) || m.RootLicensePath != "LICENSE" ||
		m.NoticePath != "NOTICE" || !strings.HasSuffix(m.RuntimeLicensePath, "/LICENSE") ||
		!strings.HasSuffix(m.JevLicensePath, "/LICENSE") ||
		m.RuntimeModule != "github.com/kimjooyoon/gooo-decision-runtime" ||
		m.JevModule != "github.com/kimjooyoon/gooo-jev" ||
		m.RuntimeVersion == "" || m.JevVersion == "" ||
		len(m.RuntimeRef) != 40 || len(m.JevRef) != 40 {
		return false
	}
	for _, digest := range []string{m.RootLicenseSHA256, m.RuntimeLicenseSHA256, m.JevLicenseSHA256} {
		if len(digest) != len("sha256:")+64 || !strings.HasPrefix(digest, "sha256:") {
			return false
		}
	}
	return true
}

func verifyPayloadModules(root string, members requiredMembers) error {
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return err
	}
	want := map[string]string{members.RuntimeModule: members.RuntimeVersion, members.JevModule: members.JevVersion}
	actual := map[string]string{}
	inRequire := false
	for _, original := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(original)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if strings.HasPrefix(line, "replace ") || line == "replace (" {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MODULE_REPLACEMENT_UNBOUND")
		}
		if !inRequire && !strings.HasPrefix(line, "require ") {
			continue
		}
		if strings.HasPrefix(line, "require ") {
			line = strings.TrimPrefix(line, "require ")
		}
		if strings.Contains(line, "// indirect") {
			continue
		}
		line, _, _ = strings.Cut(line, "//")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 || actual[fields[0]] != "" {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MODULE_GRAPH_UNBOUND")
		}
		actual[fields[0]] = fields[1]
	}
	if inRequire || len(actual) != len(want) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_MODULE_GRAPH_UNBOUND")
	}
	for module, version := range want {
		if actual[module] != version {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MODULE_GRAPH_UNBOUND")
		}
	}
	return nil
}
