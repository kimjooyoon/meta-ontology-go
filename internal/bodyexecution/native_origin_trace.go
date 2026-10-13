package bodyexecution

import "fmt"

func validateRecordedOriginTrace(runtime CompositionRuntime, trace CompositionTrace, counts *CompositionReadSummary, reserved *int) error {
	if !trace.ReadsObserved && (len(trace.Reads) > 0 || len(trace.Constructions) > 0 || trace.ReadsDropped != 0 || trace.ReadRegistrationDropped != 0 || trace.ConstructionsDropped != 0) {
		return fmt.Errorf("trace contains unobserved native origins")
	}
	for _, pair := range []struct {
		total *int
		value int
	}{{&counts.DroppedReads, trace.ReadsDropped}, {&counts.RegistrationDropped, trace.ReadRegistrationDropped}, {&counts.ConstructionDropped, trace.ConstructionsDropped}} {
		if err := addRecordedOriginDrops(pair.total, pair.value); err != nil {
			return err
		}
	}
	counts.RecordedReads += len(trace.Reads)
	counts.RecordedConstructions += len(trace.Constructions)
	if counts.RecordedReads > readObservationLimit || counts.RecordedConstructions > constructionObservationLimit {
		return fmt.Errorf("recorded native observation limit exceeded")
	}
	for _, created := range trace.Constructions {
		if err := validateRecordedConstruction(runtime, trace, created, reserved); err != nil {
			return err
		}
		if !created.Completed {
			counts.InterruptedConstructions++
		}
	}
	for index, event := range trace.Reads {
		if err := validateRecordedRead(runtime, trace, index, event); err != nil {
			return err
		}
		switch {
		case event.Origin != nil:
			counts.ResolvedReads++
		case event.Construction != nil:
			counts.ConstructedReads++
		default:
			counts.UnresolvedReads++
		}
	}
	return nil
}

func recordedOriginDelivery(trace CompositionTrace, id string) (int, *CompositionDelivery) {
	for i := range trace.Deliveries {
		if trace.Deliveries[i].ActivityID == id {
			return i, &trace.Deliveries[i]
		}
	}
	return -1, nil
}

func validateRecordedRead(runtime CompositionRuntime, trace CompositionTrace, index int, event CompositionReadObservation) error {
	if event.SiteIndex < 0 || event.SiteIndex >= len(runtime.ReadSites) || event.Length < 1 || event.Length > 256 || event.Index < 0 || event.Index >= event.Length {
		return fmt.Errorf("recorded read site or index differs")
	}
	site := runtime.ReadSites[event.SiteIndex]
	root, delivery := recordedOriginDelivery(trace, site.RootActivityID)
	if delivery == nil || len(delivery.BlockedBy) > 0 {
		return fmt.Errorf("recorded read has no active root delivery")
	}
	if event.Construction != nil {
		return validateRecordedConstructedRead(runtime, trace, index, root, site, event)
	}
	if origin := event.Origin; origin != nil {
		if origin.Input < 0 || origin.Input >= len(runtime.ReadInputs) || origin.Index < 0 || origin.Index >= 256 || event.Reason != "" {
			return fmt.Errorf("recorded input element origin differs")
		}
		input := runtime.ReadInputs[origin.Input]
		owner, _ := recordedOriginDelivery(trace, input.ActivityID)
		if owner < 0 || owner > root || input.ElementType != site.ElementType {
			return fmt.Errorf("recorded input origin type or delivery differs")
		}
	} else if event.Reason != "NO_REGISTERED_INPUT_ELEMENT" {
		return fmt.Errorf("recorded unresolved read requires its original reason")
	}
	return nil
}

func validateRecordedConstructedRead(runtime CompositionRuntime, trace CompositionTrace, index, root int, site CompositionReadSite, event CompositionReadObservation) error {
	origin := event.Construction
	if origin.Instance < 0 || origin.Instance >= len(trace.Constructions) || origin.Index < 0 || event.Origin != nil || event.Reason != "" {
		return fmt.Errorf("recorded construction origin is ambiguous or out of range")
	}
	created := trace.Constructions[origin.Instance]
	owner := runtime.ConstructionSites[created.SiteIndex]
	ownerIndex, _ := recordedOriginDelivery(trace, owner.RootActivityID)
	if !created.Completed || created.ReadEnd > index || origin.Index >= len(created.Elements) || ownerIndex < 0 || ownerIndex > root || owner.ElementType != site.ElementType {
		return fmt.Errorf("recorded read precedes its constructed origin or differs in type")
	}
	return nil
}
