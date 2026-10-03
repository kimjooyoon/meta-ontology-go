package bodyexecution

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func ownedScopeToolchain(receipt *completeness.CompletenessReceipt) (ToolchainObservation, error) {
	var reference ToolchainObservation
	raw, err := json.Marshal(receipt.Scope["owned_toolchain"])
	if err != nil || decode(raw, &reference, 1<<20) != nil {
		return reference, fmt.Errorf("incomplete owned toolchain")
	}
	scope, ok := receipt.Scope["runtime_scope"].(map[string]any)
	if !ok || reference.Schema != "gooo/owned-go-toolchain/v1" || !reference.BytesVerified || !reference.BuildInfoVerified ||
		!runtimeDigest(reference.KeySHA256) || !runtimeDigest(reference.GoToolSHA256) ||
		reference.GoToolSHA256 != scope["go_tool_sha256"] || reference.GoVersion != scope["go_version"] ||
		len(reference.SourceOutput) > 256 || strings.TrimSpace(string(reference.SourceOutput)) != reference.GoVersion {
		return reference, fmt.Errorf("owned toolchain identity or source output differs")
	}
	fields := strings.Fields(reference.GoVersion)
	if len(fields) != 4 || fields[0] != "go" || fields[1] != "version" || fields[2] != "go1.27.1" ||
		!strings.Contains(fields[3], "/") {
		return reference, fmt.Errorf("unsupported owned Go version")
	}
	bound, err := json.Marshal([]any{reference.KeySHA256, reference.GoToolSHA256, reference.GoVersion, reference.SourceCheck, reference.SourceOutput})
	p := reference.SourceCheck
	if err != nil || reference.SourceCheckSHA256 != digest(bound) || !p.Started || !p.Completed ||
		p.ExitCode == nil || *p.ExitCode != 0 || p.Canceled || p.TimedOut || p.OutputTruncated || p.DiagnosticsBytes != 0 ||
		p.WallNS <= 0 || p.UserNS < 0 || p.SystemNS < 0 || (p.PeakRSSBytes != nil && *p.PeakRSSBytes < 0) ||
		p.StdoutSHA256 != digest(reference.SourceOutput) || p.StderrSHA256 != digest(nil) {
		return reference, fmt.Errorf("owned version process binding differs")
	}
	return reference, nil
}
