package coupling

import (
	"testing"
)

func TestLiveQueryRequiresEveryContributingLocation(t *testing.T) {
	required := []string{"code://billing/pay-order", "owner://billing/pay-order", "term://pay-order", "evidence://coupling"}
	for _, missing := range required {
		t.Run(missing, func(t *testing.T) {
			request := liveRequest()
			locations := make([]SourceLocation, 0, len(request.Locations.Locations)-1)
			for _, location := range request.Locations.Locations {
				if location.StableID != missing {
					locations = append(locations, location)
				}
			}
			request.Locations.Locations = locations
			result := liveAdapter(t).Resolve(request)
			if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticLiveMissingLocations {
				t.Fatalf("missing %s result = %#v", missing, result)
			}
		})
	}
}
func TestLiveQueryRejectsDuplicateAndMismatchedSourceMapLocations(t *testing.T) {
	request := liveRequest()
	request.Locations.Locations = append(request.Locations.Locations, request.Locations.Locations[1])
	result := liveAdapter(t).Resolve(request)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticAmbiguous {
		t.Fatalf("duplicate location result = %#v", result)
	}

	request = liveRequest()
	request.Locations.Locations[1].SourceMapDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	result = liveAdapter(t).Resolve(request)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticStaleSnapshot {
		t.Fatalf("mismatched source-map result = %#v", result)
	}

	request = liveRequest()
	request.Locations.Locations[0].SourceMapID = "sourcemap://other"
	result = liveAdapter(t).Resolve(request)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticStaleSnapshot {
		t.Fatalf("mismatched origin source-map ID result = %#v", result)
	}
}
