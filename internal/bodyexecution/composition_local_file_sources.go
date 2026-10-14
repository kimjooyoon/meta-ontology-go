package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

// File annotations explain static sites once. Native event rows refer to those
// sites, so repeated writes and uses do not copy file text or locations.
type CompositionLocalFileSources struct {
	Schema               string                        `json:"schema"`
	SelectedSourceSHA256 string                        `json:"selected_source_sha256"`
	Bindings             []CompositionLocalFileBinding `json:"bindings"`
	SHA256               string                        `json:"sha256"`
}

type CompositionLocalFileBinding struct {
	Kind         string                             `json:"kind"`
	SiteIndex    int                                `json:"site_index"`
	Declaration  CompositionLocalFileLocation       `json:"declaration"`
	Value        CompositionLocalFileLocation       `json:"value"`
	Construction *bodycodegen.SourceFillInputOrigin `json:"construction,omitempty"`
}

type CompositionLocalFileLocation struct {
	Status       string `json:"status"`
	Reason       string `json:"reason,omitempty"`
	View         string `json:"view"`
	PackagePath  string `json:"package_path"`
	Filename     string `json:"filename"`
	ActivityID   string `json:"activity_id"`
	SourceSHA256 string `json:"source_sha256"`
	Start        int    `json:"start_byte"`
	End          int    `json:"end_byte"`
	Expression   string `json:"expression,omitempty"`
}

// These additive annotations do not change an existing lexical site identity,
// native counts, or the program's result. Missing history stays unannotated.
func RecordLocalFileSources(runtime *CompositionRuntime, bindings []CompositionLocalFileBinding) {
	runtime.LocalFileSources = nil
	if len(bindings) == 0 {
		return
	}
	runtime.LocalFileSources = &CompositionLocalFileSources{Schema: "gooo/native-local-file-sources/v1",
		SelectedSourceSHA256: runtime.SelectedSourceSHA256, Bindings: bindings}
	runtime.LocalFileSources.SHA256 = localFileSourcesDigest(*runtime.LocalFileSources)
}

func localFileSourcesDigest(sources CompositionLocalFileSources) string {
	sources.SHA256 = ""
	return compositionDigest(sources)
}
