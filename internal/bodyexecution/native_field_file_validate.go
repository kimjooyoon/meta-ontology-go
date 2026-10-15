package bodyexecution

import "fmt"

func validateFieldFileSources(r CompositionRuntime) error {
	set := r.FieldFileSources
	if set == nil {
		return nil
	}
	if set.Schema != "gooo/native-field-file-sources/v1" || set.SelectedSourceSHA256 != r.SelectedSourceSHA256 || set.FieldSourcesSHA256 != r.FieldSources.SHA256 || set.SHA256 != fieldFileSourcesDigest(*set) || len(set.Bindings) != len(r.FieldSources.Bindings) {
		return fmt.Errorf("field file-source identity or population differs")
	}
	seen := map[[2]int]bool{}
	for _, b := range set.Bindings {
		kind, limit := 0, len(r.FieldSites)
		if b.Kind == "use" {
			kind, limit = 1, len(r.FieldUseSites)
		} else if b.Kind != "write" {
			return fmt.Errorf("field file-source subject differs")
		}
		key := [2]int{kind, b.SiteIndex}
		if b.SiteIndex < 0 || b.SiteIndex >= limit || seen[key] {
			return fmt.Errorf("field file-source site differs")
		}
		seen[key] = true
		for _, location := range []CompositionLocalFileLocation{b.Target, b.Value} {
			if err := validateInputFileLocation(location); err != nil {
				return err
			}
		}
		if err := validateInputFileConstruction(b.Construction); err != nil {
			return err
		}
		if err := validateFieldChoiceSource(r, b); err != nil {
			return err
		}
	}
	return nil
}
