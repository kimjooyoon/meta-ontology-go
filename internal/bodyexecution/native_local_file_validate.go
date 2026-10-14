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
			if location.View != "workspace_input_file" {
				return fmt.Errorf("local file-source view differs")
			}
			switch location.Status {
			case "UNMAPPED":
				if location.Reason == "" || location.Start != 0 || location.End != 0 {
					return fmt.Errorf("unmapped local file-source claims a range")
				}
			case "MAPPED":
				if location.Reason != "" || location.PackagePath == "" || location.Filename == "" || location.ActivityID == "" || location.SourceSHA256 == "" || location.Start < 0 || location.End <= location.Start {
					return fmt.Errorf("mapped local file-source is incomplete")
				}
			default:
				return fmt.Errorf("unknown local file-source mapping status")
			}
		}
		if c := binding.Construction; c != nil {
			if c.Schema != "gooo/source-fill-input-origin/v1" || c.View != "workspace_input_file" || c.PackagePath == "" || c.Filename == "" || c.ActivityID == "" || c.SourceSHA256 == "" {
				return fmt.Errorf("local assembly file-source is incomplete")
			}
			if c.Status != "MAPPED" && c.Status != "PARTIAL" && c.Status != "UNMAPPED" {
				return fmt.Errorf("unknown local assembly file-source status")
			}
			for _, fragment := range c.Fragments {
				if fragment.Status != "MAPPED" && fragment.Status != "UNMAPPED" {
					return fmt.Errorf("unknown local assembly fragment status")
				}
				var start, end int
				switch fragment.Kind {
				case "retained_template":
					if fragment.TemplateSource != nil {
						start, end = fragment.TemplateSource.Start, fragment.TemplateSource.End
					}
				case "declared_candidate":
					if fragment.ExpressionSource != nil {
						start, end = fragment.ExpressionSource.Start, fragment.ExpressionSource.End
					}
				case "derived_grammar":
					if fragment.GrammarSource != nil {
						start, end = fragment.GrammarSource.Start, fragment.GrammarSource.End
					}
				default:
					return fmt.Errorf("unknown local assembly fragment kind")
				}
				if fragment.Status == "MAPPED" && (start < 0 || end <= start) {
					return fmt.Errorf("mapped local assembly fragment has no range")
				}
			}
		}
	}
	return nil
}
