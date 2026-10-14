package bodyexecution

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFillLargePage11739(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	raw, err := os.ReadFile(filepath.Join(root, "examples/native-origin-reader/local-fill11591-runtime.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record CompositionRuntime
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	// This is a synthetic paging record assembled from retained metadata. It
	// makes no claim that its 4096 writes were executed by the saved producer.
	write := record.Traces[0].LocalWrites[0]
	fragment := record.LocalSites[write.SiteIndex].Source.Construction.Fragments[0]
	origin := record.LocalSites[write.SiteIndex].Source.Construction
	origin.Fragments = nil
	for range 17 {
		origin.Fragments = append(origin.Fragments, fragment)
	}
	for index := range record.Traces {
		record.Traces[index].LocalWrites, record.Traces[index].LocalUses = nil, nil
		record.Traces[index].LocalElements = nil
	}
	for range 4096 {
		record.Traces[0].LocalWrites = append(record.Traces[0].LocalWrites, write)
	}
	record.LocalObservation, err = validateLocalObservations(record.Traces, record.LocalSites, record.LocalUseSites, record.ReadSites, record.ConstructionSites)
	if err != nil {
		t.Fatal("synthetic paging population", err)
	}
	for _, offset := range []int{65536, 69630} {
		rows, page, err := SelectRecordedNativeOrigins(record, NativeOriginQuery{Kind: "local-fill-fragments", Offset: offset, Limit: 2})
		if err != nil || len(rows) != 2 || page.TotalRows != 69632 || page.Truncated != (offset == 65536) {
			t.Fatal("fragment page after scalar-use limit", offset, err, page)
		}
		for index, row := range rows {
			var fact LocalFillFragmentFact
			if err := json.Unmarshal(row, &fact); err != nil || fact.AssignmentIndex != (offset+index)/17 {
				t.Fatal("page position changed", err, fact)
			}
		}
		if page.NextOffset != nil && ValidateNativeOriginQuery(NativeOriginQuery{Kind: page.Query.Kind, Offset: *page.NextOffset, Limit: 2}) != nil {
			t.Fatal("next page cannot be requested")
		}
	}
	t.Log("synthetic 69632fragment rows; middle/final pages after65536 exact; next offset usable; producer/model executions0")
}
