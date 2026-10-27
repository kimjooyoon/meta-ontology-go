package embeduse_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle/embeduse"
)

func TestPartitionPlanFindsAValidButIncompleteConsumer(t *testing.T) {
	root := fixture(t)
	plan, err := embeduse.PreparePartition(t.Context(), root, []string{"contract.gooo"},
		"contract.gooo", "gooo://contract/result", "contract-extra.gooo")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"family": "PRESERVED", "single": "CHANGED", "inside": "CHANGED", "outside": "PRESERVED"}
	if len(plan.Consumers) != len(want) || len(plan.Partition.Files) != 2 {
		t.Fatal("consumer obligations or source outputs missing")
	}
	for _, change := range plan.Consumers {
		consumer := plan.Before.Consumers[change.Consumer]
		if change.State != want[consumer.Name] || change.Before == nil || change.After == nil || change.Reason != "" {
			t.Fatalf("wrong consumer result: %s %+v", consumer.Name, change)
		}
		if change.State == "CHANGED" {
			if len(change.Before.Declarations) != 3 || len(change.After.Declarations) != 1 || len(change.LostIDs) != 2 || len(change.NewIDs) != 0 {
				t.Fatal("valid partial source was accepted as equivalent", change)
			}
		} else if change.Before.SemanticDigest != change.After.SemanticDigest || len(change.LostIDs) != 0 {
			t.Fatal("complete source view lost its meaning", change)
		}
	}
	assertOriginalFiles(t, root)
	replay, err := embeduse.PreparePartition(t.Context(), root, []string{"contract.gooo"},
		"contract.gooo", "gooo://contract/result", "contract-extra.gooo")
	if err != nil || !reflect.DeepEqual(plan, replay) {
		t.Fatal("partition planning changed on replay", err)
	}
}

func TestPartitionPlanRetainsUnassessedAndNewConsumers(t *testing.T) {
	root := fixture(t)
	appendFixture(t, root, "consumer.go", "\n//go:embed contract*.gooo other.gooo\nvar mixed embed.FS\n"+
		"//go:embed *extra.gooo\nvar newcomer embed.FS\n")
	writeFixture(t, root, "other.gooo", "package other\nnamespace other\nentity Other id \"urn:other\"\n")
	writeFixture(t, root, "existing-extra.gooo", header+"entity Extra id \"urn:extra\"\n")
	plan, err := embeduse.PreparePartition(t.Context(), root, []string{"contract.gooo"},
		"contract.gooo", "gooo://contract/result", "contract-extra.gooo")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, change := range plan.Consumers {
		name := plan.Before.Consumers[change.Consumer].Name
		if name != "mixed" && name != "newcomer" {
			continue
		}
		found[name] = true
		if change.State != "UNASSESSED" || change.Reason == "" || change.After != nil {
			t.Fatalf("incomplete source scope reported as preserved: %s %+v", name, change)
		}
	}
	if len(found) != 2 {
		t.Fatal("newly affected or unsupported consumer was omitted")
	}
}

func TestPartitionPlanRefusesUnobservedSourcesAndExistingDestinations(t *testing.T) {
	root := fixture(t)
	writeFixture(t, root, "unread.gooo", first+tail)
	for _, row := range []struct{ file, into string }{
		{"unread.gooo", "contract-extra.gooo"}, {"contract.gooo", "unread.gooo"},
	} {
		plan, err := embeduse.PreparePartition(t.Context(), root, []string{row.file}, row.file,
			"gooo://contract/result", row.into)
		if err == nil || !reflect.DeepEqual(plan, embeduse.PartitionPlan{}) {
			t.Fatal("unobserved source or occupied destination produced a plan", row, err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	plan, err := embeduse.PreparePartition(ctx, root, []string{"contract.gooo"}, "contract.gooo",
		"gooo://contract/result", "contract-extra.gooo")
	if err == nil || !reflect.DeepEqual(plan, embeduse.PartitionPlan{}) {
		t.Fatal("cancelled partition produced a plan")
	}
	assertOriginalFiles(t, root)
}

func appendFixture(t *testing.T, root, name, extra string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, name, string(raw)+extra)
}

func writeFixture(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPartitionPlanBindsExactCopiedDeclarationBytes(t *testing.T) {
	root := fixture(t)
	plan, err := embeduse.PreparePartition(t.Context(), root, []string{"contract.gooo"},
		"contract.gooo", "gooo://contract/result", "contract-extra.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range plan.Partition.After.Provenance.Declarations {
		for _, file := range plan.Partition.Files {
			if origin.Span.File == file.Filename && !strings.Contains(first+tail, file.Source[origin.Span.Start:origin.Span.End]) {
				t.Fatal("declaration body no longer points to original bytes")
			}
		}
	}
}
