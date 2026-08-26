package metricintervention

import (
	"fmt"
	"path"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func Predict(manifest metric.Manifest, plan metric.Plan) (metric.Delta, error) {
	if !metric.ValidManifest(manifest) || !metric.ValidPlan(plan) {
		return metric.Delta{}, fmt.Errorf("counterfactual plan foundation is invalid")
	}
	languages, directories, err := manifestIndex(manifest)
	if err != nil {
		return metric.Delta{}, err
	}
	var delta metric.Delta
	changedFiles, changedDirectories := map[string]bool{}, map[string]bool{}
	for _, mutation := range plan.Mutations {
		if !validMetricPath(mutation.Path) {
			return metric.Delta{}, fmt.Errorf("mutation path %q is invalid", mutation.Path)
		}
		for _, directory := range directoryChain(mutation.Path) {
			changedDirectories[directory] = true
		}
		changedFiles[mutation.Path] = true
		switch mutation.Kind {
		case "APPEND":
			err = predictAppend(languages, mutation, &delta)
		case "CREATE":
			err = predictCreate(languages, directories, mutation, &delta)
		default:
			err = fmt.Errorf("mutation kind %q is unsupported", mutation.Kind)
		}
		if err != nil {
			return metric.Delta{}, err
		}
	}
	delta.ChangedFiles = len(changedFiles)
	delta.ChangedDirectories = len(changedDirectories)
	return delta, nil
}
func predictAppend(languages map[string]string, mutation metric.Mutation, delta *metric.Delta) error {
	language, exists := languages[mutation.Path]
	if !exists {
		return fmt.Errorf("append target %q is absent", mutation.Path)
	}
	addLanguage(delta, language, 0, artifact.CountLines([]byte(mutation.Content)))
	return nil
}
func predictCreate(languages map[string]string, directories map[string]bool, mutation metric.Mutation, delta *metric.Delta) error {
	if _, exists := languages[mutation.Path]; exists {
		return fmt.Errorf("create target %q already exists", mutation.Path)
	}
	for _, directory := range directoryChain(mutation.Path)[1:] {
		if directories[directory] {
			continue
		}
		directories[directory] = true
		delta.RecursiveFolders++
		if path.Dir(directory) == "." {
			delta.DirectFolders++
		}
	}
	delta.RecursiveFiles++
	if path.Dir(mutation.Path) == "." {
		delta.DirectFiles++
	}
	language := languageOf(mutation.Path)
	languages[mutation.Path] = language
	addLanguage(delta, language, 1, artifact.CountLines([]byte(mutation.Content)))
	return nil
}
