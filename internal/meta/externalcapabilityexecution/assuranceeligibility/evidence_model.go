package assuranceeligibility

type evidence struct {
	Assurance             assuranceReport
	ParentReport          parentReport
	ParentObservation     parentObservation
	ParentSuite           parentSuite
	CapabilityReport      capabilityReport
	CapabilityObservation capabilityObservation
	CapabilitySuite       capabilitySuite
}
