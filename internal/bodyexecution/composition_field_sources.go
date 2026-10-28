package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

// Source coordinates live once per static site. Dynamic journals and their
// existing metadata keep their original identity and wire representation.
type CompositionFieldSources struct {
	Schema               string                          `json:"schema"`
	SelectedSourceSHA256 string                          `json:"selected_source_sha256"`
	Bindings             []CompositionFieldSourceBinding `json:"bindings"`
	SHA256               string                          `json:"sha256"`
}

// Derived fields bind an existing source-local operation to a declared member.
// They do not assert that the source contained a field-assignment expression.
type CompositionFieldDerivedOrigin struct {
	Kind     string                      `json:"kind"`
	RecordID string                      `json:"record_id"`
	FieldID  string                      `json:"field_id"`
	Local    bodycodegen.BodyLocalOrigin `json:"local"`
}

type CompositionFieldSourceBinding struct {
	Derivation *CompositionFieldDerivedOrigin `json:"derivation,omitempty"`
	Kind       string                         `json:"kind"`
	SiteIndex  int                            `json:"site_index"`
	Origin     *bodycodegen.BodyFieldOrigin   `json:"origin,omitempty"`
	Source     bodycodegen.BodyFieldSource    `json:"source"`
}

func fieldSourcesDigest(sources CompositionFieldSources) string {
	sources.SHA256 = ""
	return compositionDigest(sources)
}

func (p *compositionFieldProjection) recordSources(prior Composition, bindings []CompositionFieldSourceBinding) {
	p.sources = &CompositionFieldSources{Schema: "gooo/native-field-sources/v1", SelectedSourceSHA256: prior.SelectedSourceSHA256, Bindings: bindings}
	for _, binding := range bindings {
		if binding.Derivation != nil {
			p.sources.Schema = "gooo/native-field-sources/v2"
			break
		}
	}
	p.sources.SHA256 = fieldSourcesDigest(*p.sources)
}
