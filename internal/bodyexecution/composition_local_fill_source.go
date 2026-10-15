package bodyexecution

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func compositionLocalFillOrigin(prior Composition, origin bodycodegen.BodyLocalOrigin) (*bodycodegen.SourceFillLocalOrigin, string) {
	var selected *bodycodegen.SourceFillLocalOrigin
	merge := func(fill *bodycodegen.IRBodyFillReceipt) bool {
		if fill == nil || fill.SourceOrigin == nil {
			return true
		}
		link := fill.SourceOrigin.LocalOrigin(origin)
		if link == nil {
			return true
		}
		if selected != nil && !reflect.DeepEqual(selected, link) {
			return false
		}
		selected = link
		return true
	}
	for _, step := range prior.Steps {
		if !merge(step.Generation.Report.BodyFill) {
			return nil, "AMBIGUOUS_ASSEMBLY_INPUT"
		}
	}
	for _, preparation := range prior.Preparations {
		if !merge(preparation.Generation.Report.BodyFill) {
			return nil, "AMBIGUOUS_ASSEMBLY_INPUT"
		}
	}
	return selected, ""
}
