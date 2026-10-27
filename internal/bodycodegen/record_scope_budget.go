package bodycodegen

import (
	"fmt"
	"go/types"
)

// ValidateRecordUseBudget also checks a realized candidate, whose typed body can
// reference records absent from its original hole or baseline. Declarations
// alone are not uses. The naming view must match the checked projection so a
// source display name cannot impersonate another record's generated name.
func ValidateRecordUseBudget(information *types.Info, records []RecordType, native bool) error {
	if information == nil {
		return fmt.Errorf("record budget requires typed use information")
	}
	used := map[string]bool{}
	for _, value := range information.Types {
		if named, ok := types.Unalias(value.Type).(*types.Named); ok {
			used[named.Obj().Name()] = true
		}
	}
	_, err := checkedRecordScope(records, used, "typed body", native)
	return err
}

func checkedRecordScope(declarations []RecordType, used map[string]bool, scope string, native bool) ([]RecordType, error) {
	var selected []RecordType
	for _, record := range declarations {
		name := record.Name
		if native {
			name = record.GoName
		}
		if !used[name] {
			continue
		}
		if len(selected) >= recordTypeLimit {
			return nil, fmt.Errorf("record %q (%s) is record %d in the %s; body profile supports at most %d",
				record.Name, record.ID, len(selected)+1, scope, recordTypeLimit)
		}
		selected = append(selected, record)
	}
	return selected, nil
}
