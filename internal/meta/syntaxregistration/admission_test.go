package syntaxregistration

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func admissionFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/admission-v111.go.fixture")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "c8ec0446341b1a72731b7801034813a3d193fe8c38a38f55637e8abf0f0079b2" {
		t.Fatal("frozen v111 original predicate source changed")
	}
	return raw
}

func admissionSource(t *testing.T, units map[string][]byte) *goSource {
	t.Helper()
	source, err := parseSourceUnits(units, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := bindContractSource(contractSource)
	if err != nil {
		t.Fatal(err)
	}
	_, activities := registrationRoles()
	for index, name := range activities {
		if name == "GenerateDenominatorAdmission" {
			source.activity = binding.activities[index]
		}
	}
	if source.activity == "" {
		t.Fatal("Gooo denominator admission activity is missing")
	}
	return source
}

func admissionSets(t *testing.T, source *goSource) []admissionSet {
	t.Helper()
	sets, err := readAdmissionSets(source)
	if err != nil {
		t.Fatal(err)
	}
	return sets
}

func TestAdmissionNormalizationKeepsFrozenNativeOutcomes(t *testing.T) {
	legacy := admissionSource(t, map[string][]byte{closureRoot + "denominator.go": admissionFixture(t)})
	baseline := admissionSets(t, legacy)
	source := currentAdmissionSource(t)
	for _, set := range admissionSets(t, source) {
		if len(set.helpers) == 0 {
			set.normalize(source)
		}
	}
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	assertAdmissionNativeOracle(t, legacy, baseline, changes)
	writeAdmissionProposal(t, source, changes)
}

func currentAdmissionSource(t *testing.T) *goSource {
	t.Helper()
	root := "../../.."
	paths, err := filepath.Glob(filepath.Join(root, closureRoot, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	units := map[string][]byte{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		units[filepath.ToSlash(relative)] = raw
	}
	return admissionSource(t, units)
}

func assertAdmissionNativeOracle(t *testing.T, legacy *goSource, baseline []admissionSet, changes map[string][]byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/admission_oracle.go.fixture")
	if err != nil {
		t.Fatal(err)
	}
	keys, versions := []string{}, []string{}
	for _, row := range baseline[0].rows {
		keys = append(keys, row.key)
		versions = append(versions, strconv.Itoa(row.version))
	}
	text := strings.NewReplacer("LEGACY_DIGEST", legacy.text(baseline[0].root),
		"LEGACY_HEADER", legacy.text(baseline[1].root), "DIGEST_VALUES", strings.Join(keys, ","),
		"HEADER_VERSIONS", strings.Join(versions, ",")).Replace(string(raw))
	directory := t.TempDir()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	overlay := writeAdmissionOverlay(t, directory, root, changes, text)
	nativeCommand(t, root, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-overlay", overlay,
		"./internal/meta/languageassurance/verticalsliceclosureshadow", "-run", "^TestBoundedAdmissionOracle$", "-count=1")
	t.Logf("frozen v111 original predicates: %d digests and %d headers plus changed/unknown values matched native Go", len(keys), len(versions))
}

func writeAdmissionOverlay(t *testing.T, directory, root string, changes map[string][]byte, oracle string) string {
	t.Helper()
	files := maps.Clone(changes)
	files[closureRoot+"admission_projection_oracle_test.go"] = []byte(oracle)
	replacements := map[string]string{}
	for index, path := range sortedPaths(files) {
		target := filepath.Join(directory, fmt.Sprintf("member%d.go", index))
		if err := os.WriteFile(target, files[path], 0600); err != nil {
			t.Fatal(err)
		}
		replacements[filepath.Join(root, path)] = target
	}
	raw, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeAdmissionProposal(t *testing.T, source *goSource, changes map[string][]byte) {
	t.Helper()
	if path := os.Getenv("GOOO_ADMISSION_NORMALIZATION_PROPOSAL"); path != "" {
		members := []Member{}
		for _, name := range sortedPaths(changes) {
			raw := changes[name]
			members = append(members, Member{Path: name, BeforeDigest: digest(source.units[name]), AfterDigest: digest(raw),
				Content: raw, ActivityIDs: []string{source.activity}})
		}
		raw, err := json.MarshalIndent(members, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAdmissionAppendPreservesEveryPriorRow(t *testing.T) {
	path := closureRoot + "denominator.go"
	units := map[string][]byte{path: admissionFixture(t)}
	for step := range 25 {
		source := admissionSource(t, units)
		before := admissionSets(t, source)
		if err := generateAdmission(source, 112+step, 147+step); err != nil {
			t.Fatal(step, err)
		}
		changes, err := source.finish()
		if err != nil {
			t.Fatal(err)
		}
		maps.Copy(units, changes)
		after := admissionSets(t, admissionSource(t, units))
		for index, set := range after {
			if len(set.rows) != len(before[index].rows)+1 || !reflect.DeepEqual(set.rows[:len(set.rows)-1], before[index].rows) {
				t.Fatal("append changed historical admission", step, index)
			}
			assertAdmissionHelperBounds(t, admissionSource(t, units), set)
		}
	}
}

func assertAdmissionHelperBounds(t *testing.T, source *goSource, set admissionSet) {
	t.Helper()
	for _, helper := range set.helpers {
		text := source.text(helper)
		if len(strings.Split(text, "\n")) > 40 {
			t.Fatal("admission helper exceeded 40 physical lines", helper.Name)
		}
		for _, line := range strings.Split(text, "\n") {
			if len(line) > 120 {
				t.Fatal("admission helper exceeded 120 columns", helper.Name)
			}
		}
	}
}
