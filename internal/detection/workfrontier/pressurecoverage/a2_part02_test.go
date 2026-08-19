package pressurecoverage

var a2PrecedenceCases = []a2PrecedenceCase{
	{
		name:  "malformed A1 input",
		edit:  func(input *Input) { input.RequiredPressureIDs[0] = "p x" },
		wantD: DecisionFailClosed,
		wantR: ReasonInvalidInput,
	},
	{
		name:  "blank binding",
		edit:  func(input *Input) { input.PolicyDigest = "" },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "stale binding",
		edit:  func(input *Input) { input.PolicyDigest = "stale" },
		wantD: DecisionUnknown,
		wantR: ReasonSnapshotMismatch,
	},
	{
		name:  "zero K",
		bind:  true,
		edit:  func(input *Input) { input.RequestedK = 0 },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "zero minimum",
		bind:  true,
		edit:  func(input *Input) { input.MinimumIndependent = 0 },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "K below floor",
		bind:  true,
		edit:  func(input *Input) { input.RequestedK = 1 },
		wantD: DecisionFailClosed,
		wantR: ReasonPolicyFloorViolation,
	},
	{
		name:  "minimum below floor",
		bind:  true,
		edit:  func(input *Input) { input.MinimumIndependent = 1 },
		wantD: DecisionFailClosed,
		wantR: ReasonPolicyFloorViolation,
	},
	{
		name:  "minimum above K",
		bind:  true,
		edit:  func(input *Input) { input.MinimumIndependent = 3 },
		wantD: DecisionFailClosed,
		wantR: ReasonPolicyFloorViolation,
	},
	{
		name:  "empty required",
		bind:  true,
		edit:  func(input *Input) { input.RequiredPressureIDs = nil },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "missing record",
		bind:  true,
		edit:  func(input *Input) { input.PressureRecords = input.PressureRecords[1:] },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "blank group",
		bind:  true,
		edit:  func(input *Input) { input.PressureRecords[0].IndependenceGroupID = "" },
		wantD: DecisionUnknown,
		wantR: ReasonApplicabilityOrGroupUnproven,
	},
	{
		name:  "blank applicability",
		bind:  true,
		edit:  func(input *Input) { input.PressureRecords[0].ApplicabilityRuleID = "" },
		wantD: DecisionUnknown,
		wantR: ReasonApplicabilityOrGroupUnproven,
	},
	{
		name: "same group",
		bind: true,
		edit: func(input *Input) {
			for index := range input.PressureRecords {
				input.PressureRecords[index].IndependenceGroupID = "group-a"
			}
		},
		wantD: DecisionUnknown,
		wantR: ReasonIndependentGroupShortfall,
	},
}
