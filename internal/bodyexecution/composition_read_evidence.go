package bodyexecution

func observedCompositionReadEvidence(prior Composition) (string, string, []CompositionFaultSite, []CompositionConditionSite, []CompositionReadSite, []CompositionReadInput, []CompositionConstructionSite, error) {
	o, err := observedCompositionLocalEvidence(prior)
	return o.projection, o.driver, o.faults, o.conditions, o.reads, o.inputs, o.constructions, err
}

type compositionObservedSource struct {
	projection, driver string
	faults             []CompositionFaultSite
	conditions         []CompositionConditionSite
	reads              []CompositionReadSite
	inputs             []CompositionReadInput
	constructions      []CompositionConstructionSite
	locals             []CompositionLocalSite
	uses               []CompositionLocalUseSite
	fields             *compositionFieldProjection
}

func observedCompositionLocalEvidence(prior Composition) (o compositionObservedSource, err error) {
	o.projection, o.driver, o.faults, o.conditions, err = observedCompositionEvidence(prior)
	if err != nil {
		return o, err
	}
	o.reads, err = compositionReadSites(prior)
	if err != nil {
		return o, err
	}
	o.constructions, err = compositionConstructionSites(prior)
	if err != nil {
		return o, err
	}
	o.locals, err = compositionLocalSites(prior)
	if err != nil {
		return o, err
	}
	o.uses, err = compositionLocalUseSites(prior, o.locals)
	if err != nil {
		return o, err
	}
	fieldBindings, err := compositionFieldBindings(prior)
	if err != nil {
		return o, err
	}
	if len(o.reads)+len(o.constructions)+len(o.locals)+len(fieldBindings.bindings) == 0 {
		return o, nil
	}
	o.inputs = compositionReadInputs(prior)
	prefix := readObservationPrefix(prior.Source)
	o.projection, err = observedReadProjectionWithLocals(o.projection, prefix, prior, o.reads, o.constructions, o.locals, o.uses, o.faults)
	if err != nil {
		return o, err
	}
	if len(fieldBindings.bindings) > 0 {
		o.projection, o.fields, err = observedCompositionFieldProjectionWithFaults(o.projection, prefix, prior, o.faults)
		if err != nil {
			return o, err
		}
	}
	observer := readObserver
	if len(o.constructions) > 0 {
		observer = constructionReadObserver
	}
	if len(o.locals) > 0 {
		observer, err = localReadObserver()
		if err != nil {
			return o, err
		}
	}
	if o.fields != nil && len(o.fields.sites) > 0 {
		observer, err = fieldReadObserver()
		if err != nil {
			return o, err
		}
	}
	o.driver, err = observedReadDriverWithObserver(o.driver, prefix, prior, observer)
	return o, err
}
