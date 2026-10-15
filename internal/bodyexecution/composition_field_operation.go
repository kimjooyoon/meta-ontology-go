package bodyexecution

import "fmt"

func validateFieldOperation(site CompositionFieldSite) error {
	if site.DependencyMode != "compound_update" {
		if site.Operation != "" {
			return fmt.Errorf("field operation without compound dependency mode")
		}
		return nil
	}
	if site.Kind != "field_assignment" {
		return fmt.Errorf("compound update requires a field assignment")
	}
	switch site.Operation {
	case "+=", "-=", "*=", "/=":
		return nil
	}
	return fmt.Errorf("unsupported compound field operation")
}

func validateCompoundFieldReadWindow(write []int, uses [][]int, p *compositionFieldProjection) error {
	site := p.sites[write[1]]
	if site.DependencyMode != "compound_update" || write[5] > 0 {
		return nil
	}
	if write[3] == write[4] {
		return fmt.Errorf("compound field update has no previous-value read")
	}
	read := uses[write[3]]
	readSite := p.uses[read[1]]
	if read[2] != write[2] || readSite.Binding != site.Binding || readSite.Field != site.Field {
		return fmt.Errorf("compound field update previous-value identity differs")
	}
	return nil
}
