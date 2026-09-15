package predecessorbinding

func transitionIndicators(value BindingTransition) []TransitionIndicator {
	return []TransitionIndicator{
		{ID: "dynamic-binding-bps", Class: "OUTCOME", Before: value.BeforeBPS,
			After: value.AfterBPS, Delta: value.BPSDelta, Unit: "BASIS_POINT"},
		{ID: "dynamic-coordinates", Class: "DRIVER", Before: value.BeforeDynamic,
			After: value.AfterDynamic, Delta: value.DynamicDelta, Unit: "COORDINATE"},
		{ID: "static-coordinates", Class: "GUARDRAIL", Before: value.BeforeStatic,
			After: value.AfterStatic, Delta: value.StaticDelta, Unit: "COORDINATE"},
		{ID: "unknown-coordinates", Class: "GUARDRAIL", After: value.Unknown,
			Delta: value.Unknown, Unit: "COORDINATE"},
		{ID: "repository-writes", Class: "GUARDRAIL", After: value.RepositoryWrites,
			Delta: value.RepositoryWrites, Unit: "REPOSITORY_WRITE"},
	}
}
