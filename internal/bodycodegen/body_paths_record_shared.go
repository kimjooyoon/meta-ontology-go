package bodycodegen

import "github.com/kimjooyoon/gooo-decision-runtime/jointdecision"

func loadCompilerSharedThree(name, feature string) (typedPathModel, error) {
	if feature == jointdecision.RecordSharedFeatureVersion {
		model, err := jointdecision.LoadRecordSharedThree(name)
		return typedPathModel{three: model}, err
	}
	model, err := jointdecision.LoadSharedThree(name)
	return typedPathModel{three: model}, err
}
