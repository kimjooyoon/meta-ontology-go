package generation

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

const contractTypesFile = "operation-input-contract-types.gooo"
const contractActivitiesFile = "operation-input-contract-activities.gooo"

// Partition the actual authored contract; no second ABI fixture can drift.
func partitionedOperationContract(t *testing.T) fstest.MapFS {
	t.Helper()
	raw := string(operationContractFixture(t))
	cut := strings.Index(raw, "activity CollapseAssignReturn")
	if cut < 0 {
		t.Fatal("actual contract no longer contains the partition boundary")
	}
	header := "package operationinputcontract\nnamespace operationinputcontract\n\n"
	return fstest.MapFS{
		contractTypesFile:      {Data: []byte(raw[:cut])},
		contractActivitiesFile: {Data: []byte(header + raw[cut:])},
	}
}

func TestNativeContractConsumesPartitionedPhysicalSources(t *testing.T) {
	files := partitionedOperationContract(t)
	before, err := loadOperationInputContract()
	if err != nil {
		t.Fatal(err)
	}
	after, err := loadOperationInputContractFS(files)
	if err != nil {
		t.Fatal(err)
	}
	if before.SourceDigest == after.SourceDigest || before.SemanticDigest != after.SemanticDigest {
		t.Fatal("partition must change source identity and preserve semantic identity")
	}
	if len(after.Bindings) != 4 || len(after.ObligationFacts) != 9 || len(after.PolicyFacts) != 2 ||
		!reflect.DeepEqual(before.Bindings, after.Bindings) || !reflect.DeepEqual(before.Facts, after.Facts) ||
		!reflect.DeepEqual(before.ObligationFacts, after.ObligationFacts) ||
		!reflect.DeepEqual(before.PolicyFacts, after.PolicyFacts) {
		t.Fatal("partition changed the admitted native contract")
	}
	assertPhysicalContractOrigins(t, after.SourceOrigins, files)
	evidence, err := ExtractFunctionInputContractEvidence()
	if err != nil || !reflect.DeepEqual(evidence.SourceOrigins, before.SourceOrigins) {
		t.Fatalf("executor admission lost embedded source origins: %v", err)
	}
}

func assertPhysicalContractOrigins(t *testing.T, origins sourcebundle.Provenance, files fstest.MapFS) {
	t.Helper()
	if len(origins.Members) != 2 || len(origins.Declarations) != 26 {
		t.Fatalf("missing physical origins: members=%d declarations=%d", len(origins.Members), len(origins.Declarations))
	}
	for _, member := range origins.Members {
		raw := files[member.Filename].Data
		if member.SHA256 != fmt.Sprintf("%x", sha256.Sum256(raw)) || member.ByteLength != len(raw) {
			t.Fatalf("member is not bound to actual bytes: %+v", member)
		}
	}
	counts, ids := map[string]int{}, map[string]bool{}
	for _, origin := range origins.Declarations {
		span := origin.Span
		file, ok := files[span.File]
		if !ok || !span.Valid() || span.Start < 0 || span.End > len(file.Data) || span.Start >= span.End {
			t.Fatalf("origin does not locate a real source range: %+v", origin)
		}
		declaration := string(file.Data[span.Start:span.End])
		kind := "activity "
		if span.File == contractTypesFile {
			kind = "entity "
			if !strings.Contains(declaration, `id "`+origin.ID+`"`) {
				t.Fatalf("entity origin changed stable identity: %+v", origin)
			}
		}
		if origin.ID == "" || ids[origin.ID] || !strings.HasPrefix(declaration, kind) ||
			span.StartLine != strings.Count(string(file.Data[:span.Start]), "\n")+1 {
			t.Fatalf("wrong physical declaration origin: %+v", origin)
		}
		ids[origin.ID] = true
		counts[span.File]++
	}
	if counts[contractTypesFile] != 11 || counts[contractActivitiesFile] != 15 {
		t.Fatalf("declarations were attributed to the wrong partition: %v", counts)
	}
}

func TestNativeContractRejectsPartialOrChangedPartition(t *testing.T) {
	mutations := map[string]func(fstest.MapFS){
		"missing member":         func(files fstest.MapFS) { delete(files, contractTypesFile) },
		"syntax in later member": func(files fstest.MapFS) { files[contractTypesFile].Data = []byte("?") },
		"different headers": func(files fstest.MapFS) {
			files[contractTypesFile].Data = []byte(strings.Replace(string(files[contractTypesFile].Data),
				"namespace operationinputcontract", "namespace other", 1))
		},
		"duplicate across members": func(files fstest.MapFS) {
			files[contractActivitiesFile].Data = append(files[contractActivitiesFile].Data,
				[]byte("\nentity FunctionSource id \"gooo://meta-operation-input-contract/entity/function-input\"\n")...)
		},
		"unresolved import scope": func(files fstest.MapFS) {
			files[contractTypesFile].Data = []byte(strings.Replace(string(files[contractTypesFile].Data),
				"\n\n", "\nimport \"other/package\"\n\n", 1))
		},
		"changed obligation chain": func(files fstest.MapFS) {
			files[contractActivitiesFile].Data = []byte(strings.Replace(string(files[contractActivitiesFile].Data),
				"ProveControlFlow(ReturnShapeObligation)", "ProveControlFlow(FunctionSource)", 1))
		},
		"extra activity": func(files fstest.MapFS) {
			files[contractActivitiesFile].Data = append(files[contractActivitiesFile].Data,
				[]byte("\nactivity Extra(FunctionSource) -> OperationResult\n")...)
		},
		"extra policy": func(files fstest.MapFS) {
			files[contractActivitiesFile].Data = append(files[contractActivitiesFile].Data,
				[]byte("\npolicy Extra id \"urn:example:extra-policy\" { state \"READY\" }\n")...)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			files := partitionedOperationContract(t)
			if _, err := loadOperationInputContractFS(files); err != nil {
				t.Fatal(err)
			}
			mutate(files)
			got, err := loadOperationInputContractFS(files)
			if err == nil || !reflect.DeepEqual(got, operationInputContract{}) {
				t.Fatalf("invalid bundle inherited successful admission: err=%v", err)
			}
			if name == "extra policy" && !strings.Contains(err.Error(), "unknown operation input contract declaration") {
				t.Fatalf("valid policy did not reach the native ABI boundary: %v", err)
			}
		})
	}
	for _, files := range []fs.FS{nil, fstest.MapFS{}} {
		if got, err := loadOperationInputContractFS(files); err == nil || !reflect.DeepEqual(got, operationInputContract{}) {
			t.Fatal("missing source bundle admitted a contract")
		}
	}
}

func TestNativeContractCacheBindsNamesBytesAndMemberOrder(t *testing.T) {
	files := partitionedOperationContract(t)
	sources, err := sourcebundle.Read(files, []string{contractTypesFile, contractActivitiesFile})
	if err != nil {
		t.Fatal(err)
	}
	var cache operationContractParseCache
	first, err := cache.load(sources, [32]byte{})
	if err != nil {
		t.Fatal(err)
	}
	variants := [][]sourcebundle.Source{
		{{Filename: "renamed.gooo", Content: sources[0].Content}, sources[1]},
		{{Filename: sources[0].Filename, Content: sources[0].Content + "\n"}, sources[1]},
		{sources[1], sources[0]},
	}
	for i, variant := range variants {
		got, err := cache.load(variant, [32]byte{})
		if err != nil || got.SourceDigest == first.SourceDigest || got.SemanticDigest != first.SemanticDigest ||
			cache.parses != uint64(i+2) {
			t.Fatalf("source identity did not cause fresh admission: variant=%d err=%v parses=%d", i, err, cache.parses)
		}
	}
}
