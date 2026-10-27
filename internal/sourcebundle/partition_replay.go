package sourcebundle

import (
	"fmt"
	"reflect"
)

// ValidatePartition reconstructs a proposal from current bytes using its recorded
// profile, then compares all semantics, physical origins and projected bytes.
// It does not discover consumers or authorize publication.
func ValidatePartition(sources []Source, proposal Partition) error {
	if len(proposal.Files) != 2 {
		return fmt.Errorf("source partition requires exactly two output files")
	}
	var observed Partition
	var err error
	if proposal.EntityFieldsSupport == nil {
		observed, err = PreparePartition(sources, proposal.Files[0].Filename,
			proposal.BoundaryID, proposal.Files[1].Filename)
	} else {
		observed, err = PreparePartitionWithEntityFieldsSupport(sources, proposal.Files[0].Filename,
			proposal.BoundaryID, proposal.Files[1].Filename, *proposal.EntityFieldsSupport)
	}
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(observed, proposal) {
		return fmt.Errorf("source partition proposal differs from current source replay")
	}
	return nil
}
