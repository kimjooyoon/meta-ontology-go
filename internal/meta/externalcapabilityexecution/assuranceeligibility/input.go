package assuranceeligibility

import _ "embed"

//go:embed evidence/assurance.json
var assuranceEvidence []byte

type Input struct {
	SubjectSHA string
	Payloads   map[string][]byte
}

func NewInput(subjectSHA string, parentReport, parentObservation, parentSuite,
	capabilityReport, capabilityObservation, capabilitySuite []byte) Input {
	return Input{SubjectSHA: subjectSHA, Payloads: map[string][]byte{
		AssuranceName: append([]byte(nil), assuranceEvidence...),
		ParentReportName: append([]byte(nil), parentReport...),
		ParentObservationName: append([]byte(nil), parentObservation...),
		ParentSuiteName: append([]byte(nil), parentSuite...),
		CapabilityReportName: append([]byte(nil), capabilityReport...),
		CapabilityObservationName: append([]byte(nil), capabilityObservation...),
		CapabilitySuiteName: append([]byte(nil), capabilitySuite...),
	}}
}

func EmbeddedAssurance() []byte { return append([]byte(nil), assuranceEvidence...) }

func (input Input) clone() Input {
	result := Input{SubjectSHA: input.SubjectSHA, Payloads: make(map[string][]byte, len(input.Payloads))}
	for name, payload := range input.Payloads {
		result.Payloads[name] = append([]byte(nil), payload...)
	}
	return result
}

func available(input Input) bool {
	for _, name := range artifactNames {
		if len(input.Payloads[name]) == 0 {
			return false
		}
	}
	return true
}
