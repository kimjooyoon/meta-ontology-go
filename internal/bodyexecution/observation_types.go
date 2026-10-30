package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/nativeobservation"

// Keep the internal API while exposing one wire layout to Go consumers.
type ProcessObservation = nativeobservation.ProcessObservation
type ProcessTiming = nativeobservation.ProcessTiming
type ToolchainObservation = nativeobservation.ToolchainObservation
type ArtifactObservation = nativeobservation.ArtifactObservation
