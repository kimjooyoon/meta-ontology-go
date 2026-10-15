package bodyexecution

import "fmt"

func validateLocalFileSources(runtime CompositionRuntime) error {
	set := runtime.LocalFileSources
	if set == nil {
		return nil
	}
	if set.Schema != "gooo/native-local-file-sources/v1" || set.SelectedSourceSHA256 != runtime.SelectedSourceSHA256 || set.SHA256 != localFileSourcesDigest(*set) || len(set.Bindings) != len(runtime.LocalSites)+len(runtime.LocalUseSites) {
		return fmt.Errorf("local file-source identity or population differs")
	}
	seen := map[[2]int]bool{}
	for _, binding := range set.Bindings {
		kind, limit := 0, len(runtime.LocalSites)
		if binding.Kind == "use" {
			kind, limit = 1, len(runtime.LocalUseSites)
		} else if binding.Kind != "write" {
			return fmt.Errorf("unknown local file-source subject")
		}
		key := [2]int{kind, binding.SiteIndex}
		if binding.SiteIndex < 0 || binding.SiteIndex >= limit || seen[key] {
			return fmt.Errorf("local file-source site index differs")
		}
		seen[key] = true
		for _, location := range []CompositionLocalFileLocation{binding.Declaration, binding.Value} {
			if err := validateInputFileLocation(location); err != nil {
				return err
			}
		}
		if err := validateInputFileConstruction(binding.Construction); err != nil {
			return err
		}
	}
	return nil
}
