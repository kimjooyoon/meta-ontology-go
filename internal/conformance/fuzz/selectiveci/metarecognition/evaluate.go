package metarecognition

func Evaluate(c Case) (Outcome, Outcome) {
	return evaluateGooo(c), evaluateBaseline(c)
}

func evaluateBaseline(c Case) Outcome {
	b := c.Baseline
	state, reason, ids := ClosedSound, ReasonExactBinding, []string{}
	switch b.Subject {
	case SubjectBinding:
		if !b.DirectivePresent {
			state, reason, ids = UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.ObservedFile}
		} else if !b.RegistryPresent || !b.SourceMapPresent || b.Ambiguous {
			state, reason, ids = UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.StableID}
		} else if b.BoundID != b.StableID || b.ExpectedFile != b.ObservedFile || b.ExpectedBlob != b.ObservedBlob {
			state, reason, ids = FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}
		} else {
			ids = []string{b.StableID}
			if b.DeclarationName != "" && b.DeclarationName != "Order" {
				reason = ReasonRenameBinding
			}
		}
	case SubjectGraph:
		if len(b.UnknownIDs) > 0 {
			state, reason, ids = UnknownFullSuiteRequired, ReasonUnknownGraph, b.UnknownIDs
		} else if len(b.MissedIDs) > 0 {
			state, reason, ids = FailClosedUnsound, ReasonMissedObligation, b.MissedIDs
		}
	case SubjectSoundness:
		state, reason, ids = baselineSoundness(b)
	case SubjectPath:
		if b.Path.Duplicate {
			state, reason, ids = FailClosedUnsound, ReasonDuplicateReceipt, b.Path.IDs
		} else if b.Path.Conflict {
			state, reason, ids = FailClosedUnsound, ReasonConflictingReceipt, b.Path.IDs
		}
	case SubjectResource:
		if !b.Resource.Valid || b.Resource.Overflow {
			state, reason, ids = UnknownFullSuiteRequired, ReasonInvalidResource, []string{"receipt-1"}
		}
	}
	return baselineOutcome(state, reason, ids, b, b.SelectedCommands)
}

func baselineSoundness(b BaselineConfig) (State, Reason, []string) {
	if externalMissing(b.External) {
		return UnknownFullSuiteRequired, ReasonExternalMissing, []string{externalInputID(b.External)}
	}
	for _, command := range b.Commands {
		if command.GlobalGuard && !command.Selected {
			return FailClosedUnsound, ReasonGlobalGuard, []string{command.ID}
		}
	}
	for _, command := range b.Commands {
		if command.Selected && (command.FullStatus != command.SelectedStatus || command.FullDigest != command.SelectedDigest) {
			return FailClosedUnsound, ReasonSelectedDrift, []string{command.ID}
		}
	}
	for _, command := range b.Commands {
		if command.FullFailure && !command.Selected {
			return FailClosedUnsound, ReasonOmittedFailure, []string{command.ID}
		}
	}
	for _, command := range b.Commands {
		if command.Impacted && !command.Selected {
			return FailClosedUnsound, ReasonOmittedFailure, []string{command.ID}
		}
	}
	if b.Obligation.Authority != Authoritative && !b.Obligation.Selected {
		return UnknownFullSuiteRequired, ReasonNonAuthoritative, []string{b.Obligation.ID}
	}
	return ClosedSound, ReasonExactBinding, nil
}

func baselineOutcome(state State, reason Reason, ids []string, b BaselineConfig, selected int) Outcome {
	if selected > b.FullCommands {
		selected = b.FullCommands
	}
	work := Work{Selected: selected, Full: b.FullCommands, ProvRecords: b.ProvRecords, ProvPaths: b.ProvPaths}
	work.Units = work.Selected
	return Outcome{State: state, Reason: reason, LocalizedIDs: sorted(ids), Work: work}
}
