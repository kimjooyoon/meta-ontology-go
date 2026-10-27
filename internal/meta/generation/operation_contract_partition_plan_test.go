package generation

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle/embeduse"
)

func TestNativeContractAdmissionFollowsEachPlannedConsumer(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	raw := operationContractFixture(t)
	before, err := sourcebundle.Lower([]sourcebundle.Source{{Filename: "operation-input-contract.gooo", Content: string(raw)}})
	if err != nil {
		t.Fatal(err)
	}
	boundary, ok := before.IR.Graph.NodeByName(before.IR.Namespace, "CollapseAssignReturn")
	if !ok {
		t.Fatal("authored operation is absent")
	}
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod": "module example.test/planned-native-contract\n\ngo 1.27.3\n",
		"consumer.go": "package contract\nimport \"embed\"\n//go:embed operation-input-contract*.gooo\nvar family embed.FS\n" +
			"//go:embed operation-input-contract.gooo\nvar single []byte\n",
		"operation-input-contract.gooo": string(raw),
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	plan, err := embeduse.PreparePartition(t.Context(), root, []string{"operation-input-contract.gooo"},
		"operation-input-contract.gooo", boundary.ID.String(), contractActivitiesFile)
	if err != nil || len(plan.Consumers) != 2 {
		t.Fatal("native consumer plan failed", err)
	}
	original, err := loadOperationInputContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range plan.Consumers {
		consumer := plan.After.Consumers[change.Consumer]
		files := fstest.MapFS{}
		for _, index := range consumer.Members {
			for _, source := range plan.Partition.Sources {
				if source.Filename == plan.After.Files[index].Filename {
					files[source.Filename] = &fstest.MapFile{Data: []byte(source.Content)}
				}
			}
		}
		admitted, err := loadOperationInputContractFS(files)
		if consumer.Name == "single" {
			if change.State != "CHANGED" || len(change.LostIDs) != 12 || err == nil || !reflect.DeepEqual(admitted, operationInputContract{}) {
				t.Fatal("incomplete native contract was accepted", change, err)
			}
			continue
		}
		if err != nil || change.State != "PRESERVED" || len(admitted.SourceOrigins.Declarations) != 22 ||
			len(admitted.Bindings) != 4 || len(admitted.ObligationFacts) != 6 || len(admitted.PolicyFacts) != 2 ||
			!reflect.DeepEqual(original.Bindings, admitted.Bindings) ||
			!reflect.DeepEqual(original.ObligationFacts, admitted.ObligationFacts) ||
			!reflect.DeepEqual(original.PolicyFacts, admitted.PolicyFacts) {
			t.Fatal("complete planned contract changed native admission", change, err)
		}
	}
	current, err := os.ReadFile(filepath.Join(root, "operation-input-contract.gooo"))
	if err != nil || string(current) != string(raw) {
		t.Fatal("planning changed original native source")
	}
	if _, err := os.Stat(filepath.Join(root, contractActivitiesFile)); !os.IsNotExist(err) {
		t.Fatal("planning installed a source member")
	}
}
