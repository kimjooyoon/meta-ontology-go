package bodyexecution

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateInputFileLocation(location CompositionLocalFileLocation) error {
	if location.View != "workspace_input_file" {
		return fmt.Errorf("input file-source view differs")
	}
	switch location.Status {
	case "UNMAPPED":
		if location.Reason == "" || location.Start != 0 || location.End != 0 {
			return fmt.Errorf("unmapped input file-source claims a range")
		}
	case "MAPPED":
		if location.Reason != "" || location.PackagePath == "" || location.Filename == "" || location.ActivityID == "" || location.SourceSHA256 == "" || location.Start < 0 || location.End <= location.Start {
			return fmt.Errorf("mapped input file-source is incomplete")
		}
	default:
		return fmt.Errorf("unknown input file-source mapping status")
	}
	return nil
}

func validateInputFileConstruction(c *bodycodegen.SourceFillInputOrigin) error {
	if c == nil {
		return nil
	}
	if c.Schema != "gooo/source-fill-input-origin/v1" || c.View != "workspace_input_file" || c.PackagePath == "" || c.Filename == "" || c.ActivityID == "" || c.SourceSHA256 == "" {
		return fmt.Errorf("assembly file-source is incomplete")
	}
	if c.Status != "MAPPED" && c.Status != "PARTIAL" && c.Status != "UNMAPPED" {
		return fmt.Errorf("unknown assembly file-source status")
	}
	for _, fragment := range c.Fragments {
		if fragment.Status != "MAPPED" && fragment.Status != "UNMAPPED" {
			return fmt.Errorf("unknown assembly fragment status")
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
			return fmt.Errorf("unknown assembly fragment kind")
		}
		if fragment.Status == "MAPPED" && (start < 0 || end <= start) {
			return fmt.Errorf("mapped assembly fragment has no range")
		}
	}
	return nil
}
