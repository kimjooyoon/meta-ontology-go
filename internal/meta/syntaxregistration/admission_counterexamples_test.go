package syntaxregistration

import (
	"maps"
	"reflect"
	"strings"
	"testing"
)

func normalizedAdmissionFixture(t *testing.T) map[string][]byte {
	t.Helper()
	source := admissionSource(t, map[string][]byte{closureRoot + "denominator.go": admissionFixture(t)})
	for _, set := range admissionSets(t, source) {
		set.normalize(source)
	}
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	return changes
}

func TestAdmissionUnsupportedPredicatesHaveNoPartialEdits(t *testing.T) {
	raw := string(admissionFixture(t))
	for _, altered := range []string{
		strings.Replace(raw, "digest != DenominatorDigest", "digest == DenominatorDigest", 1),
		strings.Replace(raw, "digest != DenominatorMigrationDigest", "digest != DenominatorDigest", 1),
		strings.Replace(raw, "value.Version == 21", "value.Version == 22", 1),
		strings.Replace(raw, "if target, known := historicalSyntaxTargetChunk0(version); known", "if target, known := historicalSyntaxTargetChunk0(version); !known", 1),
	} {
		assertAdmissionRefusal(t, raw, altered)
	}
}

func TestAdmissionRejectsLocallyShadowedDigestBindings(t *testing.T) {
	raw := string(admissionFixture(t))
	t.Run("function-local", func(t *testing.T) {
		altered := strings.Replace(raw, "digest := digestBytes(raw)",
			"digest := digestBytes(raw)\nconst DenominatorDigest = DenominatorMigrationDigest", 1)
		assertAdmissionRefusal(t, raw, altered)
	})
	t.Run("if-initializer", func(t *testing.T) {
		altered := strings.Replace(raw, "\tif (", "\tif DenominatorDigest := DenominatorMigrationDigest; (", 1)
		assertAdmissionRefusal(t, raw, altered)
	})
}

func TestAdmissionUnsupportedHelpersHaveNoPartialEdits(t *testing.T) {
	raw := string(normalizedAdmissionFixture(t)[closureRoot+"denominator.go"])
	for _, altered := range []string{
		strings.Replace(raw, "return true", "return false", 1),
		strings.Replace(raw, "return acceptedDenominatorDigestChunk1(digest)", "return acceptedDenominatorDigestChunk0(digest)", 1),
		strings.Replace(raw, "switch digest {", "switch {", 1),
		strings.Replace(raw, "case DenominatorMigrationDigest:", "case DenominatorDigest:", 1),
		strings.Replace(raw, "validHeader := acceptedDenominatorHeaderChunk0(value)", "validHeader := !acceptedDenominatorHeaderChunk0(value)", 1),
		raw + "\nfunc acceptedDenominatorDigestChunk99(digest string) bool { return false }\n",
		raw + "\nvar acceptedDenominatorDigestChunk99 string\n",
	} {
		assertAdmissionRefusal(t, raw, altered)
	}
}

func assertAdmissionRefusal(t *testing.T, original, altered string) {
	t.Helper()
	if original == altered {
		t.Fatal("admission counterexample did not alter source")
	}
	path := closureRoot + "denominator.go"
	source := admissionSource(t, map[string][]byte{path: []byte(altered)})
	if err := generateAdmission(source, 112, 147); err == nil || len(source.edits) != 0 {
		t.Fatal("unsupported admission produced partial edits", err, len(source.edits))
	}
	if string(source.units[path]) != altered {
		t.Fatal("admission changed original source")
	}
}

func TestAdmissionAppendPreservesEarlierPhysicalOwners(t *testing.T) {
	units := splitAdmissionOwners(t, normalizedAdmissionFixture(t))
	for version := 112; version <= 133; version++ {
		source := admissionSource(t, units)
		sets := admissionSets(t, source)
		owners := map[string]bool{}
		for _, set := range sets {
			path, _ := source.location(set.helpers[len(set.helpers)-1].Pos())
			owners[path] = true
			if err := set.appendVersion(source, version); err != nil {
				t.Fatal(err)
			}
		}
		changes, err := source.finish()
		if err != nil || !reflect.DeepEqual(sortedPaths(owners), sortedPaths(changes)) {
			t.Fatal("append changed an earlier physical owner", err, sortedPaths(changes))
		}
		maps.Copy(units, changes)
		admissionSets(t, admissionSource(t, units))
	}
}

func splitAdmissionOwners(t *testing.T, units map[string][]byte) map[string][]byte {
	t.Helper()
	source := admissionSource(t, units)
	result := maps.Clone(units)
	for _, set := range admissionSets(t, source) {
		for _, helper := range set.helpers {
			result[closureRoot+helper.Name.Name+".go"] = []byte("package verticalsliceclosureshadow\n" + source.text(helper))
			source.replace(helper, "")
		}
	}
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	maps.Copy(result, changes)
	return result
}
