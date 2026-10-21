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
	OutsideScope     []string          `json:"outside_scope_current_references"`
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
	dir        string
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
	GoMod               string `json:"go_mod"`
	GoSum               string `json:"go_sum"`
	Policy              string `json:"policy"`
	License             string `json:"license"`
	Document            string `json:"document"`
	OutsideReadme       string `json:"outside_readme"`
	OutsideScalarSource string `json:"outside_scalar_source"`
	OutsideScalarModel  string `json:"outside_scalar_model"`
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	source, paths, outside, err := readAffectedSurfaces(ctx, root, module, licensePath)
	if err != nil {
		return DirectModuleImpact{}, err
	}
	result := DirectModuleImpact{Schema: "gooo/direct-module-impact/v1", Module: module,
		CurrentVersion: current, ProposedVersion: proposed, Inputs: map[string]string{},
		Outputs: paths[:5], OutsideScope: outside, Unresolved: []string{}}
	if err := readImpactPreimages(root, append(paths, outside...), &result); err != nil {
		return DirectModuleImpact{}, err
	}
	for _, name := range outside {
		result.Unresolved = append(result.Unresolved, "active current reference outside candidate: "+name)
	}
	currentBound, currentDownload := observeCurrentModule(root, module, current, currentRef,
		licensePath, m, &result)
	originBound, sumsBound, licenseBound := observeProposedModule(module, proposed,
		current, currentDownload, &result)
	if err := assessDirectModule(ctx, source, policy, currentBound, originBound,
		sumsBound, licenseBound, &result); err != nil {
		return DirectModuleImpact{}, err
	}
	result.Provenance = directModuleProvenance(result)
	return result, nil
}

func readAffectedSurfaces(ctx context.Context, root, module, licensePath string) ([]byte, []string, []string, error) {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil {
		return nil, nil, nil, err
	}
	input, _ := json.Marshal([]any{map[string]any{"module_path": module}})
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"RequiredSurfaces", input)
	if err != nil || value.ActivityID != "urn:gooo:direct-module-impact:required-surfaces" ||
		value.SourceSHA256 != payloadDigest(source) {
		return nil, nil, nil, fmt.Errorf("Gooo affected surfaces cannot be replayed: %v", err)
	}
	var surfaces affectedSurfaces
	if err := json.Unmarshal(value.Value, &surfaces); err != nil {
		return nil, nil, nil, err
	}
	if surfaces.GoMod != "go.mod" || surfaces.GoSum != "go.sum" ||
		surfaces.Policy != payloadSourcePath || surfaces.License != licensePath ||
		surfaces.Document != payloadDocumentPath {
		return nil, nil, nil, fmt.Errorf("Gooo affected surfaces do not match release payload structure")
	}
	outside := []string{}
	if module == "github.com/kimjooyoon/gooo-decision-runtime" {
		if surfaces.OutsideReadme != "README.md" ||
			surfaces.OutsideScalarSource != "docs/source-scalar-criteria.md" ||
			surfaces.OutsideScalarModel != "docs/scalar-source-model.ko.md" {
			return nil, nil, nil, fmt.Errorf("Gooo active runtime references differ from observed structure")
		}
		outside = []string{surfaces.OutsideReadme, surfaces.OutsideScalarSource, surfaces.OutsideScalarModel}
	} else if surfaces.OutsideReadme != "" || surfaces.OutsideScalarSource != "" || surfaces.OutsideScalarModel != "" {
		return nil, nil, nil, fmt.Errorf("Gooo Jev scope contains unexpected active references")
	}
	return source, []string{surfaces.GoMod, surfaces.GoSum, surfaces.Policy,
		surfaces.License, surfaces.Document, directModuleSourcePath}, outside, nil
}

func readImpactPreimages(root string, paths []string, result *DirectModuleImpact) error {
	for _, name := range paths {
		info, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(name)))
		if statErr != nil || !info.Mode().IsRegular() {
			result.Unresolved = append(result.Unresolved, "missing regular source: "+name)
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if readErr != nil {
			return readErr
		}
		result.Inputs[name] = payloadDigest(raw)
	}
	return nil
}

func observeCurrentModule(root, module, current, currentRef, licensePath string,
	m requiredMembers, result *DirectModuleImpact) (bool, moduleDownload) {
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
	return currentBound, currentDownload
}

func observeProposedModule(module, proposed, current string, currentDownload moduleDownload,
	result *DirectModuleImpact) (bool, bool, bool) {
	if proposed == current {
		result.Download = currentDownload
	} else {
		download, err := downloadModule(module, proposed)
		if err != nil {
			result.Unresolved = append(result.Unresolved, "proposed module download: "+err.Error())
		}
		result.Download = download
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
	return originBound, sumsBound, licenseBound
}

func assessDirectModule(ctx context.Context, source []byte, policy payloadPolicy,
	currentBound, originBound, sumsBound, licenseBound bool, result *DirectModuleImpact) error {
	core, err := json.Marshal(struct {
		Module, Current, Proposed string
		Inputs                    map[string]string
		OutsideScope              []string
		Download                  moduleDownload
	}{result.Module, result.CurrentVersion, result.ProposedVersion,
		result.Inputs, result.OutsideScope, result.Download})
	if err != nil {
		return err
	}
	result.PlanSHA256 = payloadDigest(core)
	result.ObservedOutputs = len(result.Outputs) - missingDeclaredOutputs(result)
	input := map[string]any{"module_path": result.Module, "current_version": result.CurrentVersion,
		"proposed_version": result.ProposedVersion, "go_mod_sha256": result.Inputs["go.mod"],
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
		return err
	}
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"AssessModuleProposal", encoded)
	if err != nil || value.ActivityID != directModuleActivityID ||
		value.SourceSHA256 != result.Inputs[directModuleSourcePath] {
		return fmt.Errorf("Gooo module decision source cannot be replayed: %v", err)
	}
	if err := json.Unmarshal(value.Value, &result.Decision); err != nil {
		return err
	}
	if result.Decision.ModulePath != result.Module || result.Decision.ProposedVersion != result.ProposedVersion ||
		result.Decision.PlanSHA256 != result.PlanSHA256 || result.Decision.ObservedOutputs != result.ObservedOutputs {
		return fmt.Errorf("Gooo module decision differs from observed proposal")
	}
	return nil
}

func unresolvedOutputs(result *DirectModuleImpact) int {
	n := missingDeclaredOutputs(result) + len(result.OutsideScope)
	for _, reason := range result.Unresolved {
		if reason == "release document projection differs from Gooo payload" {
			n++
		}
	}
	return n
}

func missingDeclaredOutputs(result *DirectModuleImpact) int {
	n := 0
	for _, name := range result.Outputs {
		if result.Inputs[name] == "" {
			n++
		}
	}
	return n
}

func hasModuleSumPair(raw []byte, module, version string, download moduleDownload) bool {
	full, mod := 0, 0
	for line := range strings.SplitSeq(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != module {
			continue
		}
		if fields[1] == version {
			if fields[2] != download.Sum {
				return false
			}
			full++
		}
		if fields[1] == version+"/go.mod" {
			if fields[2] != download.GoModSum {
				return false
			}
			mod++
		}
	}
	return full == 1 && mod == 1
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
		if failed, ok := err.(*exec.ExitError); ok && len(failed.Stderr) != 0 {
			detail := strings.TrimSpace(string(failed.Stderr))
			if len(detail) > 300 {
				detail = detail[:300]
			}
			return moduleDownload{}, fmt.Errorf("go mod download failed: %s", detail)
		}
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
		Ref: observed.Origin.Hash, Sum: observed.Sum, GoModSum: observed.GoModSum, dir: observed.Dir}
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
	graph = append(graph, map[string]any{"@id": activity, "@type": "prov:Activity",
		"prov:used": provReferences(used)})
	graph = append(graph, map[string]any{"@id": "urn:gooo:direct-module-impact:decision:" + result.PlanSHA256,
		"@type": "prov:Entity", "prov:wasGeneratedBy": map[string]string{"@id": activity},
		"gooo:status": result.Decision.Status, "gooo:cochangeScope": result.Outputs,
		"gooo:outsideScope": result.OutsideScope})
	return map[string]any{"@context": map[string]string{"prov": "http://www.w3.org/ns/prov#",
		"gooo": "https://github.com/kimjooyoon/meta-ontology-go/ns#"}, "@graph": graph}
}

func provReferences(ids []string) []any {
	references := make([]any, 0, len(ids))
	for _, id := range ids {
		references = append(references, map[string]string{"@id": id})
	}
	return references
}
