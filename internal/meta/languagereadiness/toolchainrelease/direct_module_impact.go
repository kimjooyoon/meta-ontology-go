package toolchainrelease

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const directModuleSourcePath = "examples/direct-module-impact/plan.gooo"
const directModuleActivityID = "urn:gooo:direct-module-impact:assess"

type DirectModuleImpact struct {
	Schema           string            `json:"schema"`
	Module           string            `json:"module"`
	CurrentVersion   string            `json:"current_version"`
	ProposedVersion  string            `json:"proposed_version"`
	Inputs           map[string]string `json:"input_sha256"`
	Outputs          []string          `json:"cochange_outputs"`
	ObservedOutputs  int               `json:"observed_outputs"`
	ProjectedOutputs int               `json:"projected_outputs"`
	Unresolved       []string          `json:"unresolved"`
	Download         moduleDownload    `json:"download"`
	PlanSHA256       string            `json:"plan_sha256"`
	Decision         moduleDecision    `json:"gooo_decision"`
	Provenance       map[string]any    `json:"provenance"`
}

type moduleDownload struct {
	Path       string `json:"path"`
	Version    string `json:"version"`
	VCS        string `json:"vcs"`
	SourceURL  string `json:"source_url"`
	Ref        string `json:"ref"`
	Sum        string `json:"sum"`
	GoModSum   string `json:"go_mod_sum"`
	LicenseSHA string `json:"license_sha256"`
}

type moduleDecision struct {
	ModulePath      string `json:"module_path"`
	ProposedVersion string `json:"proposed_version"`
	PlanSHA256      string `json:"plan_sha256"`
	DownloadRef     string `json:"download_ref"`
	ObservedOutputs int    `json:"observed_outputs"`
	Status          string `json:"status"`
}

type affectedSurfaces struct {
	GoMod    string `json:"go_mod"`
	GoSum    string `json:"go_sum"`
	Policy   string `json:"policy"`
	License  string `json:"license"`
	Document string `json:"document"`
}

// ObserveDirectModuleImpact makes a read-only co-change plan. The Go adapter
// observes source and module bytes; the checked Gooo activity owns admission.
func ObserveDirectModuleImpact(root, module, proposed string) (DirectModuleImpact, error) {
	if proposed == "" || strings.ContainsAny(proposed, "@/") ||
		len(strings.Fields(proposed)) != 1 || strings.Fields(proposed)[0] != proposed {
		return DirectModuleImpact{}, fmt.Errorf("provide one exact module version")
	}
	policy, err := readPayloadPolicy(root, Targets()[0])
	if err != nil {
		return DirectModuleImpact{}, err
	}
	m := policy.Members
	current, licensePath, currentRef := "", "", ""
	switch module {
	case m.RuntimeModule:
		current, licensePath, currentRef = m.RuntimeVersion, m.RuntimeLicensePath, m.RuntimeRef
	case m.JevModule:
		current, licensePath, currentRef = m.JevVersion, m.JevLicensePath, m.JevRef
	default:
		return DirectModuleImpact{}, fmt.Errorf("module is outside the Gooo release payload")
	}
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil {
		return DirectModuleImpact{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	surfacesInput, _ := json.Marshal([]any{map[string]any{"module_path": module}})
	surfacesValue, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"RequiredSurfaces", surfacesInput)
	if err != nil || surfacesValue.ActivityID != "urn:gooo:direct-module-impact:required-surfaces" {
		return DirectModuleImpact{}, fmt.Errorf("Gooo affected surfaces cannot be replayed: %v", err)
	}
	var surfaces affectedSurfaces
	if err := json.Unmarshal(surfacesValue.Value, &surfaces); err != nil {
		return DirectModuleImpact{}, err
	}
	if surfaces.GoMod != "go.mod" || surfaces.GoSum != "go.sum" ||
		surfaces.Policy != payloadSourcePath || surfaces.License != licensePath ||
		surfaces.Document != payloadDocumentPath {
		return DirectModuleImpact{}, fmt.Errorf("Gooo affected surfaces do not match release payload structure")
	}
	paths := []string{surfaces.GoMod, surfaces.GoSum, surfaces.Policy,
		surfaces.License, surfaces.Document, directModuleSourcePath}
	result := DirectModuleImpact{Schema: "gooo/direct-module-impact/v1", Module: module,
		CurrentVersion: current, ProposedVersion: proposed, Inputs: map[string]string{},
		Outputs: paths[:5], Unresolved: []string{}}
	for _, name := range paths {
		info, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(name)))
		if statErr != nil || !info.Mode().IsRegular() {
			result.Unresolved = append(result.Unresolved, "missing regular source: "+name)
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if readErr != nil {
			return DirectModuleImpact{}, readErr
		}
		result.Inputs[name] = payloadDigest(raw)
	}
	currentBound := verifyPayloadModules(root, m) == nil
	currentBound = currentBound && currentRef != "" && result.Inputs[licensePath] != ""
	currentDownload, currentErr := downloadModule(module, current)
	if currentErr != nil {
		currentBound = false
		result.Unresolved = append(result.Unresolved, "current module download: "+currentErr.Error())
	}
	currentBound = currentBound && currentDownload.Path == module &&
		currentDownload.Version == current && currentDownload.VCS == "git" &&
		currentDownload.SourceURL == "https://"+module && currentDownload.Ref == currentRef &&
		currentDownload.LicenseSHA == result.Inputs[licensePath]
	if module == m.RuntimeModule {
		currentBound = currentBound && result.Inputs[licensePath] == m.RuntimeLicenseSHA256
	} else {
		currentBound = currentBound && result.Inputs[licensePath] == m.JevLicenseSHA256
	}
	if sum, err := os.ReadFile(filepath.Join(root, "go.sum")); err == nil {
		currentBound = currentBound && hasModuleSumPair(sum, module, current, currentDownload)
	} else {
		currentBound = false
	}
	if !currentBound {
		result.Unresolved = append(result.Unresolved, "current pin, sums or license differs from Gooo payload")
	}
	if doc, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(payloadDocumentPath))); err == nil {
		projected, projectionErr := replacePayloadModuleTable(doc, renderPayloadModuleTable(m))
		if projectionErr != nil || !bytes.Equal(doc, projected) {
			result.Unresolved = append(result.Unresolved, "release document projection differs from Gooo payload")
		}
	}
	if proposed == current {
		result.Download = currentDownload
	} else {
		result.Download, err = downloadModule(module, proposed)
		if err != nil {
			result.Unresolved = append(result.Unresolved, "proposed module download: "+err.Error())
		}
	}
	originBound := result.Download.VCS == "git" && result.Download.SourceURL == "https://"+module &&
		len(result.Download.Ref) == 40 && isHex(result.Download.Ref)
	if suffix := pseudoRevision(proposed); suffix != "" {
		originBound = originBound && strings.HasPrefix(result.Download.Ref, suffix)
	}
	sumsBound := strings.HasPrefix(result.Download.Sum, "h1:") &&
		strings.HasPrefix(result.Download.GoModSum, "h1:")
	licenseBound := strings.HasPrefix(result.Download.LicenseSHA, "sha256:")
	if !originBound || !sumsBound || !licenseBound {
		result.Unresolved = append(result.Unresolved, "proposed origin, sums or license remains unbound")
	}
	core, err := json.Marshal(struct {
		Module, Current, Proposed string
		Inputs                    map[string]string
		Download                  moduleDownload
	}{module, current, proposed, result.Inputs, result.Download})
	if err != nil {
		return DirectModuleImpact{}, err
	}
	result.PlanSHA256 = payloadDigest(core)
	result.ObservedOutputs = len(result.Outputs) - unresolvedOutputs(result)
	input := map[string]any{"module_path": module, "current_version": current,
		"proposed_version": proposed, "go_mod_sha256": result.Inputs["go.mod"],
		"policy_sha256": policy.SourceSHA256, "plan_sha256": result.PlanSHA256,
		"download_module_path": result.Download.Path, "download_version": result.Download.Version,
		"download_ref": result.Download.Ref, "module_sum": result.Download.Sum,
		"mod_sum": result.Download.GoModSum, "license_sha256": result.Download.LicenseSHA,
		"current_pin_consistent": currentBound, "origin_bound": originBound,
		"sums_bound": sumsBound, "license_bound": licenseBound,
		"observed_outputs": result.ObservedOutputs, "unresolved_outputs": unresolvedOutputs(result),
		"historical_writes": 0}
	encoded, err := json.Marshal([]any{input})
	if err != nil {
		return DirectModuleImpact{}, err
	}
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"AssessModuleProposal", encoded)
	if err != nil || value.ActivityID != directModuleActivityID ||
		value.SourceSHA256 != result.Inputs[directModuleSourcePath] {
		return DirectModuleImpact{}, fmt.Errorf("Gooo module decision source cannot be replayed: %v", err)
	}
	if err := json.Unmarshal(value.Value, &result.Decision); err != nil {
		return DirectModuleImpact{}, err
	}
	if result.Decision.ModulePath != module || result.Decision.ProposedVersion != proposed ||
		result.Decision.PlanSHA256 != result.PlanSHA256 || result.Decision.ObservedOutputs != result.ObservedOutputs {
		return DirectModuleImpact{}, fmt.Errorf("Gooo module decision differs from observed proposal")
	}
	result.Provenance = directModuleProvenance(result)
	return result, nil
}

func unresolvedOutputs(result DirectModuleImpact) int {
	n := 0
	for _, name := range result.Outputs {
		if result.Inputs[name] == "" {
			n++
		}
	}
	for _, reason := range result.Unresolved {
		if reason == "release document projection differs from Gooo payload" {
			n++
		}
	}
	return n
}

func hasModuleSumPair(raw []byte, module, version string, download moduleDownload) bool {
	full, mod := false, false
	for line := range strings.SplitSeq(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != module {
			continue
		}
		full = full || fields[1] == version && fields[2] == download.Sum
		mod = mod || fields[1] == version+"/go.mod" && fields[2] == download.GoModSum
	}
	return full && mod
}

func isHex(value string) bool {
	for _, ch := range value {
		if !strings.ContainsRune("0123456789abcdef", ch) {
			return false
		}
	}
	return true
}

func pseudoRevision(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) < 3 || len(parts[len(parts)-1]) != 12 {
		return ""
	}
	return parts[len(parts)-1]
}

func downloadModule(module, version string) (moduleDownload, error) {
	dir, err := os.MkdirTemp("", "gooo-module-impact-*")
	if err != nil {
		return moduleDownload{}, err
	}
	defer os.RemoveAll(dir)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.invalid/gooo-impact\n\ngo 1.27.3\n"), 0600); err != nil {
		return moduleDownload{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "mod", "download", "-json", module+"@"+version)
	command.Dir = dir
	raw, err := command.Output()
	if err != nil {
		return moduleDownload{}, fmt.Errorf("go mod download failed: %w", err)
	}
	var observed struct {
		Path, Version, Sum, GoModSum, Dir, Error string
		Origin                                   struct{ VCS, URL, Hash string }
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		return moduleDownload{}, err
	}
	if observed.Error != "" || observed.Path != module || observed.Version != version {
		return moduleDownload{}, fmt.Errorf("download did not resolve the exact module version")
	}
	result := moduleDownload{Path: observed.Path, Version: observed.Version,
		VCS: observed.Origin.VCS, SourceURL: observed.Origin.URL,
		Ref: observed.Origin.Hash, Sum: observed.Sum, GoModSum: observed.GoModSum}
	licensePath := filepath.Join(observed.Dir, "LICENSE")
	info, err := os.Lstat(licensePath)
	if observed.Dir != "" && filepath.IsAbs(observed.Dir) && err == nil && info.Mode().IsRegular() {
		license, err := os.ReadFile(licensePath)
		if err == nil && len(license) > 0 && len(license) <= 64<<10 {
			result.LicenseSHA = fmt.Sprintf("sha256:%x", sha256.Sum256(license))
		}
	}
	return result, nil
}

func directModuleProvenance(result DirectModuleImpact) map[string]any {
	graph := []any{}
	used := []string{}
	paths := make([]string, 0, len(result.Inputs))
	for name := range result.Inputs {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	for _, name := range paths {
		digest := result.Inputs[name]
		id := "urn:gooo:repository-file:" + name + ":" + digest
		graph = append(graph, map[string]any{"@id": id, "@type": "prov:Entity",
			"gooo:path": name, "gooo:sha256": digest})
		used = append(used, id)
	}
	if result.Download.Path != "" {
		moduleID := "urn:gooo:module-download:" + result.Module + "@" + result.ProposedVersion
		graph = append(graph, map[string]any{"@id": moduleID, "@type": "prov:Entity",
			"gooo:ref": result.Download.Ref, "gooo:vcs": result.Download.VCS,
			"gooo:source_url": result.Download.SourceURL, "gooo:sum": result.Download.Sum,
			"gooo:go_mod_sum":     result.Download.GoModSum,
			"gooo:license_sha256": result.Download.LicenseSHA})
		used = append(used, moduleID)
	}
	activity := "urn:gooo:direct-module-impact:observation:" + result.PlanSHA256
	graph = append(graph, map[string]any{"@id": activity, "@type": "prov:Activity", "prov:used": used})
	graph = append(graph, map[string]any{"@id": "urn:gooo:direct-module-impact:decision:" + result.PlanSHA256,
		"@type": "prov:Entity", "prov:wasGeneratedBy": activity,
		"gooo:status": result.Decision.Status, "gooo:cochangeScope": result.Outputs})
	return map[string]any{"@context": map[string]string{"prov": "http://www.w3.org/ns/prov#",
		"gooo": "https://github.com/kimjooyoon/meta-ontology-go/ns#"}, "@graph": graph}
}
