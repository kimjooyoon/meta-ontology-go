package bodyexecution

import (
	"fmt"
	"reflect"
)

func validateRecordedLocalOrigins(r CompositionRuntime) error {
	if r.LocalObservation == nil {
		if len(r.LocalSites)+len(r.LocalUseSites) > 0 {
			return fmt.Errorf("recorded local sites require a local summary")
		}
		for _, t := range r.Traces {
			if len(t.LocalWrites)+len(t.LocalUses)+len(t.LocalElements) > 0 || t.LocalFrames != 0 || t.LocalFrameDropped != 0 || t.LocalWriteDropped != 0 || t.LocalUseDropped != 0 {
				return fmt.Errorf("recorded local rows require a local summary")
			}
		}
		return nil
	}
	counts, err := validateLocalObservations(r.Traces, r.LocalSites, r.LocalUseSites, r.ReadSites, r.ConstructionSites)
	if err != nil {
		return err
	}
	actual := *r.LocalObservation
	actual.SHA256, actual.Scope = "", ""
	counts.SHA256, counts.Scope = "", ""
	if !reflect.DeepEqual(actual, *counts) {
		return fmt.Errorf("recorded local summary differs from actual populations")
	}
	return nil
}
