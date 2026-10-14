package bodyexecution

import "fmt"

// Check recorded references and populations without re-executing or attesting
// the program that wrote them. Partial observations retain their original scope.
func validateRecordedNativeOrigins(runtime CompositionRuntime) error {
	switch runtime.Schema {
	case "gooo/body-composition-runtime/v1", "gooo/body-composition-runtime/v2", "gooo/body-composition-runtime/v3":
	default:
		return fmt.Errorf("origin input requires a supported composition runtime")
	}
	if runtime.Stage == "" || len(runtime.Traces) < 1 || len(runtime.Traces) > 128 || runtime.ReadObservation == nil {
		return fmt.Errorf("origin input requires a stage, 1..128 traces and recorded native observations")
	}
	summary := runtime.ReadObservation
	if summary.Schema != "gooo/native-read-summary/v1" && summary.Schema != "gooo/native-read-summary/v2" {
		return fmt.Errorf("unsupported native observation summary")
	}
	if len(runtime.ReadSites) > 4096 || len(runtime.ConstructionSites) > 4096 {
		return fmt.Errorf("recorded origin site limit exceeded")
	}
	for _, site := range runtime.ConstructionSites {
		if len(site.Elements) > 256 {
			return fmt.Errorf("recorded constructor exceeds 256 elements")
		}
	}
	counts := CompositionReadSummary{}
	previous, reserved := -1, 0
	for _, trace := range runtime.Traces {
		if trace.CaseIndex <= previous || trace.CaseIndex > 127 {
			return fmt.Errorf("recorded cases must have distinct increasing indices in 0..127")
		}
		if err := validateRecordedOriginTrace(runtime, trace, &counts, &reserved); err != nil {
			return fmt.Errorf("recorded case %d: %w", trace.CaseIndex, err)
		}
		previous = trace.CaseIndex
	}
	if summary.Schema == "gooo/native-read-summary/v1" && (len(runtime.ConstructionSites) > 0 || counts.RecordedConstructions > 0 || counts.ConstructionDropped > 0 || counts.ConstructedReads > 0) {
		return fmt.Errorf("v1 observation cannot contain construction origins")
	}
	if err := validateRecordedLocalOrigins(runtime); err != nil {
		return err
	}
	if err := validateLocalFileSources(runtime); err != nil {
		return err
	}
	if err := validateRecordedFieldObservations(runtime); err != nil {
		return err
	}
	return compareRecordedOriginCounts(*summary, counts)
}

func compareRecordedOriginCounts(saved, counts CompositionReadSummary) error {
	if saved.RecordedReads != counts.RecordedReads || saved.ResolvedReads != counts.ResolvedReads || saved.UnresolvedReads != counts.UnresolvedReads || saved.ConstructedReads != counts.ConstructedReads || saved.RecordedConstructions != counts.RecordedConstructions || saved.InterruptedConstructions != counts.InterruptedConstructions || saved.DroppedReads != counts.DroppedReads || saved.RegistrationDropped != counts.RegistrationDropped || saved.ConstructionDropped != counts.ConstructionDropped {
		return fmt.Errorf("recorded native summary differs from trace populations")
	}
	status := "COMPLETE"
	if counts.UnresolvedReads > 0 || counts.DroppedReads > 0 || counts.RegistrationDropped > 0 || counts.ConstructionDropped > 0 || counts.InterruptedConstructions > 0 {
		status = "PARTIAL"
	}
	if saved.Status != status {
		return fmt.Errorf("recorded native summary status differs from partial observations")
	}
	return nil
}

func addRecordedOriginDrops(total *int, value int) error {
	if value < 0 || value > int(^uint(0)>>1)-*total {
		return fmt.Errorf("recorded origin drop count is negative or overflows")
	}
	*total += value
	return nil
}
