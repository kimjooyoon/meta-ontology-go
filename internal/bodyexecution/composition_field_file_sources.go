package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

type CompositionFieldFileSources struct {
	Schema               string                        `json:"schema"`
	SelectedSourceSHA256 string                        `json:"selected_source_sha256"`
	FieldSourcesSHA256   string                        `json:"field_sources_sha256"`
	Bindings             []CompositionFieldFileBinding `json:"bindings"`
	SHA256               string                        `json:"sha256"`
}

type CompositionFieldFileBinding struct {
	Choice       *CompositionFieldChoiceSource      `json:"choice,omitempty"`
	Kind         string                             `json:"kind"`
	SiteIndex    int                                `json:"site_index"`
	Target       CompositionLocalFileLocation       `json:"target"`
	Value        CompositionLocalFileLocation       `json:"value"`
	Construction *bodycodegen.SourceFillInputOrigin `json:"construction,omitempty"`
}

func RecordFieldFileSources(runtime *CompositionRuntime, bindings []CompositionFieldFileBinding) {
	runtime.FieldFileSources = nil
	if runtime.FieldSources == nil || len(bindings) == 0 {
		return
	}
	runtime.FieldFileSources = &CompositionFieldFileSources{Schema: "gooo/native-field-file-sources/v1", SelectedSourceSHA256: runtime.SelectedSourceSHA256, FieldSourcesSHA256: runtime.FieldSources.SHA256, Bindings: bindings}
	runtime.FieldFileSources.SHA256 = fieldFileSourcesDigest(*runtime.FieldFileSources)
}

func fieldFileSourcesDigest(sources CompositionFieldFileSources) string {
	sources.SHA256 = ""
	return compositionDigest(sources)
}
