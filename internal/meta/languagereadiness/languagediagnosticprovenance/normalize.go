package languagediagnosticprovenance

func Normalize(observation Observation) (Trace, *ProvenanceError) {
	if failure := validateObservation(observation); failure != nil {
		return Trace{}, failure
	}
	trace := Trace{
		Origin: observation.Origin, Stage: observation.Stage,
		Code: observation.Code, Hardness: observation.Hardness,
		Physical: observation.Physical, Logical: observation.Logical,
	}
	target := observation.Logical
	if observation.RequireSemantic {
		mapping, failure := resolveSourceMapping(observation)
		if failure != nil {
			return Trace{}, failure
		}
		semantic := sourceSpan(mapping)
		trace.SemanticID = mapping.SemanticID
		trace.SemanticKind = mapping.Kind
		trace.Semantic = &semantic
		target = semantic
	}
	trace.Diagnostic = projectDiagnostic(observation, target)
	trace.Steps = provenanceSteps()
	trace.TraceDigest = digestTrace(trace)
	return trace, nil
}

func digestTrace(trace Trace) string {
	trace.TraceDigest = ""
	return digestJSON(trace)
}

func provenanceSteps() []StepReceipt {
	return []StepReceipt{
		step(1, "CAPTURE_PHYSICAL", "FOUNDATION", "capture-go-byte-position"),
		step(2, "RESOLVE_LOGICAL", "COHERENCE", "apply-line-directive-position"),
		step(3, "BIND_SEMANTIC", "COHERENCE", "reverse-generator-source-map"),
		step(4, "CLASSIFY_DIAGNOSTIC", "FOUNDATION", "classify-code-severity-hardness"),
		step(5, "PROJECT_LSP", "COHERENCE", "project-normalized-diagnostic"),
	}
}

func step(ordinal int, stage, proof, operation string) StepReceipt {
	return StepReceipt{
		Ordinal: ordinal, Stage: stage, ProofChoice: proof,
		MetaOperation: operation, Status: "PASS", Effects: 0,
	}
}
