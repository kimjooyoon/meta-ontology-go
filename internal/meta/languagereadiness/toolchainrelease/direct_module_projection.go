package toolchainrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type DirectModuleProjectedFile struct {
	Path         string `json:"path"`
	InputSHA256  string `json:"input_sha256"`
	OutputSHA256 string `json:"output_sha256"`
}

type DirectModuleCandidate struct {
	Schema                string                      `json:"schema"`
	PlanSHA256            string                      `json:"plan_sha256"`
	Decision              moduleDecision              `json:"gooo_decision"`
	ProjectedOutputs      int                         `json:"projected_outputs"`
	ChangedOutputs        int                         `json:"changed_outputs"`
	VerifiedOutputs       int                         `json:"verified_outputs"`
	VerifiedElapsedMicros int64                       `json:"verification_elapsed_microseconds"`
	Verification          string                      `json:"verification_scope"`
	BuildStatus           string                      `json:"build_status"`
	Application           string                      `json:"application_status"`
	Files                 []DirectModuleProjectedFile `json:"files"`
	Provenance            map[string]any              `json:"provenance"`
}

// ProjectDirectModuleCandidate writes the Gooo-declared outputs only to
// a new directory outside the source tree, after a fresh exact-plan replay.
func ProjectDirectModuleCandidate(root, destination, module, proposed string) (_ DirectModuleCandidate, err error) {
	started := time.Now()
	impact, err := ObserveDirectModuleImpact(root, module, proposed)
	if err != nil {
		return DirectModuleCandidate{}, err
	}
	if impact.Decision.Status != "READY_FOR_DETACHED_PROJECTION" ||
		impact.ObservedOutputs != len(impact.Outputs) || len(impact.Unresolved) != 0 {
		return DirectModuleCandidate{}, fmt.Errorf("Gooo did not authorize complete direct-module projection: %s", impact.Decision.Status)
	}
	policy, err := readPayloadPolicy(root, Targets()[0])
	if err != nil {
		return DirectModuleCandidate{}, err
	}
	before := map[string][]byte{}
	for _, name := range impact.Outputs {
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if readErr != nil || payloadDigest(data) != impact.Inputs[name] {
			return DirectModuleCandidate{}, fmt.Errorf("source preimage changed: %s", name)
		}
		before[name] = data
	}
	newLicense, err := os.ReadFile(filepath.Join(impact.Download.dir, "LICENSE"))
	if err != nil || payloadDigest(newLicense) != impact.Download.LicenseSHA {
		return DirectModuleCandidate{}, fmt.Errorf("downloaded license preimage changed")
	}
	if err := verifyClosedModuleGraph(impact.Download.dir, module); err != nil {
		return DirectModuleCandidate{}, err
	}
	after, err := projectDirectModuleImages(before, policy.Members, impact, newLicense)
	if err != nil {
		return DirectModuleCandidate{}, err
	}
	fresh, err := ObserveDirectModuleImpact(root, module, proposed)
	if err != nil || fresh.PlanSHA256 != impact.PlanSHA256 ||
		fresh.Decision.Status != impact.Decision.Status || len(fresh.Unresolved) != 0 {
		return DirectModuleCandidate{}, fmt.Errorf("direct-module observation changed before materialization")
	}
	return writeDirectModuleCandidate(root, destination, impact, after, started)
}

func projectDirectModuleImages(before map[string][]byte, members requiredMembers,
	impact DirectModuleImpact, newLicense []byte) (map[string][]byte, error) {
	module, proposed := impact.Module, impact.ProposedVersion
	moduleAfter, err := replaceOne(before["go.mod"], module+" "+impact.CurrentVersion,
		module+" "+proposed)
	if err != nil {
		return nil, err
	}
	sumAfter, err := rewriteDirectSums(before["go.sum"], module, impact.CurrentVersion, proposed,
		impact.Download.Sum, impact.Download.GoModSum)
	if err != nil {
		return nil, err
	}
	policyAfter, expected, err := projectPayloadModuleFields(before[payloadSourcePath], members,
		module, proposed, impact.Download.Ref, impact.Download.LicenseSHA)
	if err != nil {
		return nil, err
	}
	if err := verifyProjectedPayload(policyAfter, expected); err != nil {
		return nil, err
	}
	docAfter, err := replacePayloadModuleTable(before[payloadDocumentPath], renderPayloadModuleTable(expected))
	if err != nil {
		return nil, err
	}
	after := map[string][]byte{"go.mod": moduleAfter, "go.sum": sumAfter,
		payloadSourcePath: policyAfter, impact.Outputs[3]: newLicense,
		payloadDocumentPath: docAfter}
	for _, rule := range impact.ActiveRules {
		projected, err := replaceOne(after[rule.Path], rule.Prefix+impact.CurrentVersion,
			rule.Prefix+impact.ProposedVersion)
		if err != nil {
			return nil, fmt.Errorf("active Gooo rule %s: %w", rule.ID, err)
		}
		after[rule.Path] = projected
	}
	if bytes.Equal(before["go.mod"], moduleAfter) || bytes.Equal(before["go.sum"], sumAfter) ||
		bytes.Equal(before[payloadSourcePath], policyAfter) || bytes.Equal(before[payloadDocumentPath], docAfter) {
		return nil, fmt.Errorf("required version projections did not change")
	}
	return after, nil
}

func writeDirectModuleCandidate(root, destination string, impact DirectModuleImpact,
	after map[string][]byte, started time.Time) (_ DirectModuleCandidate, err error) {
	dest, err := detachedModuleDestination(root, destination)
	if err != nil {
		return DirectModuleCandidate{}, err
	}
	if err := os.Mkdir(dest, 0o755); err != nil {
		return DirectModuleCandidate{}, err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(dest)
		}
	}()
	result := DirectModuleCandidate{Schema: "gooo/direct-module-detached-projection/v5",
		PlanSHA256: impact.PlanSHA256, Decision: impact.Decision,
		ProjectedOutputs: len(impact.Outputs), Files: []DirectModuleProjectedFile{},
		Verification: "candidate_file_bytes", BuildStatus: "not_run", Application: "not_applied"}
	graph, activityID, moduleID := moduleProjectionGraph(impact)
	projectedPolicyID := "urn:gooo:candidate-file:" + payloadSourcePath + ":" + payloadDigest(after[payloadSourcePath])
	for _, name := range impact.Outputs {
		path := filepath.Join(dest, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return DirectModuleCandidate{}, err
		}
		if err = os.WriteFile(path, after[name], 0o644); err != nil {
			return DirectModuleCandidate{}, err
		}
		out := DirectModuleProjectedFile{name, impact.Inputs[name], payloadDigest(after[name])}
		written, readErr := os.ReadFile(path)
		if readErr != nil || payloadDigest(written) != out.OutputSHA256 {
			return DirectModuleCandidate{}, fmt.Errorf("candidate output did not verify: %s", name)
		}
		result.VerifiedOutputs++
		result.Files = append(result.Files, out)
		if out.InputSHA256 != out.OutputSHA256 {
			result.ChangedOutputs++
		}
		graph = append(graph, projectedModuleEntity(out, name == impact.Outputs[3],
			moduleID, activityID, projectedPolicyID, sourceRuleIDsForPath(impact, name)))
	}
	result.Provenance = map[string]any{"@context": impact.Provenance["@context"], "@graph": graph}
	result.VerifiedElapsedMicros = time.Since(started).Microseconds()
	receipt, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return DirectModuleCandidate{}, err
	}
	if err = os.WriteFile(filepath.Join(dest, "candidate-receipt.json"), append(receipt, '\n'), 0o644); err != nil {
		return DirectModuleCandidate{}, err
	}
	return result, nil
}

func moduleProjectionGraph(impact DirectModuleImpact) ([]any, string, string) {
	graph := append([]any{}, impact.Provenance["@graph"].([]any)...)
	activityID := "urn:gooo:direct-module-impact:projection:" + impact.PlanSHA256
	moduleID := "urn:gooo:module-download:" + impact.Module + "@" + impact.ProposedVersion
	used := []string{"urn:gooo:direct-module-impact:decision:" + impact.PlanSHA256, moduleID}
	for _, name := range impact.Outputs {
		used = append(used, repositoryFileID(name, impact.Inputs[name]))
	}
	graph = append(graph, map[string]any{"@id": activityID, "@type": "prov:Activity",
		"prov:used": provReferences(used)})
	return graph, activityID, moduleID
}

func projectedModuleEntity(out DirectModuleProjectedFile, fromModule bool, moduleID,
	activityID, projectedPolicyID string, ruleIDs []string) map[string]any {
	derived := []string{repositoryFileID(out.Path, out.InputSHA256), moduleID}
	if fromModule {
		derived = []string{moduleID}
	} else if out.Path == payloadDocumentPath {
		derived = append(derived, projectedPolicyID)
	}
	derived = append(derived, ruleIDs...)
	return map[string]any{"@id": "urn:gooo:candidate-file:" + out.Path + ":" + out.OutputSHA256,
		"@type": "prov:Entity", "gooo:path": out.Path, "gooo:sha256": out.OutputSHA256,
		"prov:wasDerivedFrom": provReferences(derived),
		"prov:wasGeneratedBy": map[string]string{"@id": activityID}}
}

func sourceRuleIDsForPath(impact DirectModuleImpact, name string) []string {
	ids := []string{}
	for _, rule := range impact.ActiveRules {
		if rule.Path == name {
			ids = append(ids, rule.ID)
		}
	}
	for _, rule := range impact.PreservedSpans {
		if rule.Path == name {
			ids = append(ids, rule.ID)
		}
	}
	return ids
}

func replaceOne(data []byte, old, next string) ([]byte, error) {
	if bytes.Count(data, []byte(old)) != 1 {
		return nil, fmt.Errorf("expected exactly one source span: %s", old)
	}
	return bytes.Replace(data, []byte(old), []byte(next), 1), nil
}

func verifyClosedModuleGraph(dir, module string) error {
	raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return fmt.Errorf("downloaded module graph is unavailable: %w", err)
	}
	moduleSeen, goSeen := false, false
	for line := range strings.SplitSeq(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		switch {
		case line == "module "+module && !moduleSeen:
			moduleSeen = true
		case strings.HasPrefix(line, "go ") && !goSeen && len(strings.Fields(line)) == 2:
			goSeen = true
		default:
			return fmt.Errorf("proposed module has an unmodeled graph directive")
		}
	}
	if !moduleSeen || !goSeen {
		return fmt.Errorf("proposed module graph lacks exact module and Go declarations")
	}
	return nil
}

func rewriteDirectSums(data []byte, module, old, next, sum, modSum string) ([]byte, error) {
	lines := strings.Split(string(data), "\n")
	full, mod := 0, 0
	for index, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != module {
			continue
		}
		switch fields[1] {
		case old:
			lines[index] = module + " " + next + " " + sum
			full++
		case old + "/go.mod":
			lines[index] = module + " " + next + "/go.mod " + modSum
			mod++
		case next, next + "/go.mod":
			return nil, fmt.Errorf("proposed sum already occurs in the current graph")
		}
	}
	if full != 1 || mod != 1 {
		return nil, fmt.Errorf("current module sums are not a unique pair")
	}
	return []byte(strings.Join(lines, "\n")), nil
}

func projectPayloadModuleFields(source []byte, before requiredMembers, module, version,
	ref, licenseSHA string) ([]byte, requiredMembers, error) {
	type field struct{ key, old, next string }
	var fields []field
	after := before
	if module == before.RuntimeModule {
		fields = []field{{"runtime_version", before.RuntimeVersion, version},
			{"runtime_ref", before.RuntimeRef, ref},
			{"runtime_license_sha256", before.RuntimeLicenseSHA256, licenseSHA}}
		after.RuntimeVersion, after.RuntimeRef, after.RuntimeLicenseSHA256 = version, ref, licenseSHA
	} else if module == before.JevModule {
		fields = []field{{"jev_version", before.JevVersion, version},
			{"jev_ref", before.JevRef, ref},
			{"jev_license_sha256", before.JevLicenseSHA256, licenseSHA}}
		after.JevVersion, after.JevRef, after.JevLicenseSHA256 = version, ref, licenseSHA
	} else {
		return nil, requiredMembers{}, fmt.Errorf("module is absent from Gooo payload")
	}
	projected := source
	for _, item := range fields {
		var err error
		projected, err = replaceOne(projected, item.key+": "+strconv.Quote(item.old),
			item.key+": "+strconv.Quote(item.next))
		if err != nil {
			return nil, requiredMembers{}, err
		}
	}
	return projected, after, nil
}

func verifyProjectedPayload(source []byte, expected requiredMembers) error {
	input, _ := json.Marshal([]any{map[string]any{"windows": false}})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	value, err := bodycodegen.EvaluateFixedValue(ctx, payloadSourcePath, source, "RequiredMembers", input)
	if err != nil || value.ActivityID != "gooo://release-payload/required-members" {
		return fmt.Errorf("projected Gooo payload cannot be evaluated: %v", err)
	}
	actual, err := decodeStrict[requiredMembers](value.Value)
	if err != nil || actual != expected || !validRequiredMembers(actual, Targets()[0]) {
		return fmt.Errorf("projected Gooo payload does not produce the requested members")
	}
	return nil
}

func detachedModuleDestination(root, destination string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootAbs, err = filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", err
	}
	destAbs, err := filepath.Abs(destination)
	if err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(destAbs))
	if err != nil {
		return "", fmt.Errorf("candidate parent must exist: %w", err)
	}
	destAbs = filepath.Join(parent, filepath.Base(destAbs))
	relative, err := filepath.Rel(rootAbs, destAbs)
	if err != nil || relative == "." ||
		(relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return "", fmt.Errorf("candidate destination must be outside the source tree")
	}
	return destAbs, nil
}
