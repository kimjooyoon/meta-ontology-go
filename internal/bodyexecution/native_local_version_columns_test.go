package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"
)

func TestLocalVersionColumnsRetainOriginalNumbers(t *testing.T) {
	f, err := os.Open("../../examples/local-version-origins/recorded-runtime.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var runtime CompositionRuntime
	if err := json.NewDecoder(z).Decode(&runtime); err != nil {
		t.Fatal(err)
	}
	columns, report, err := SelectRecordedLocalVersions(runtime, NativeOriginQuery{Kind: "local-writes", Offset: 5, Limit: 3})
	if err != nil || report.TotalRows != 32 || report.ReturnedRows != 3 || report.ConsumerRows != 1 || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal(report, err)
	}
	for i := range columns.Assignments {
		write := runtime.Traces[0].LocalWrites[5+i]
		var declaration CompositionLocalFileLocation
		for _, binding := range runtime.LocalFileSources.Bindings {
			if binding.Kind == "write" && binding.SiteIndex == write.SiteIndex {
				declaration = binding.Declaration
			}
		}
		if columns.Cases[i] != 0 || columns.Assignments[i] != 5+i || columns.Frames[i] != write.Frame || !columns.DeclarationMapped[i] || columns.DeclarationSHA256[i] != declaration.SourceSHA256 || columns.DeclarationStarts[i] != declaration.Start || columns.DeclarationEnds[i] != declaration.End {
			t.Fatal("source or journal number changed", i)
		}
	}
	if _, _, err := SelectRecordedLocalVersions(runtime, NativeOriginQuery{Kind: "local-uses", Limit: 1}); err == nil {
		t.Fatal("different journal kind accepted")
	}
}

func TestOriginalFileVersionColumns(t *testing.T) {
	f, err := os.Open("../../examples/local-version-origins/recorded-runtime.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var runtime CompositionRuntime
	if err := json.NewDecoder(z).Decode(&runtime); err != nil {
		t.Fatal(err)
	}
	columns, report, err := SelectRecordedLocalVersions(runtime, NativeOriginQuery{Kind: "local-writes", Offset: 17, Limit: 2})
	if err != nil || columns.DeclarationScope != "original_file" || columns.FileSourcesSHA256 != runtime.LocalFileSources.SHA256 || report.InputShape != "gooo/local-version-columns/v2" {
		t.Fatal(columns, report, err)
	}
	for i := range columns.Cases {
		if columns.Cases[i] != 2 || columns.Assignments[i] != 1+i || columns.Frames[i] != 1 || columns.Packages[i] != "origin/versionproducer" || columns.Filenames[i] != "main.gooo" || columns.OriginalActivities[i] != "urn:gooo:version-producer:add" || columns.DeclarationSHA256[i] == columns.SourceSHA256 {
			t.Fatal("original and selected views were conflated", i, columns)
		}
	}
}
