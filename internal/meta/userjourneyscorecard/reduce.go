package userjourneyscorecard

import (
	"reflect"
	"sort"
)

func (s *inspection) reduceSamples() {
	known := map[string]JourneyDefinition{}
	for _, definition := range s.contract.Journeys {
		known[definition.Operation] = definition
		if definition.MetaOperation != "" && knownProof(definition.ProofChoice) {
			s.metaBindings++
		} else {
			s.unknowns++
			s.lowerResolution = true
		}
	}
	groups := map[string][]Sample{}
	for _, sample := range s.profile.Samples {
		definition, ok := known[sample.Operation]
		if !ok || !reflect.DeepEqual(sample.Arguments, definition.Arguments) || sample.WallMS < 0 || sample.MaxRSSKiB <= 0 || !validDigest(sample.StdoutDigest) || !validDigest(sample.StderrDigest) {
			s.unknowns++
			s.lowerResolution = true
			continue
		}
		groups[sample.Operation] = append(groups[sample.Operation], sample)
		s.samplesObserved++
	}
	for _, definition := range s.contract.Journeys {
		samples := groups[definition.Operation]
		sort.Slice(samples, func(i, j int) bool { return samples[i].Sequence < samples[j].Sequence })
		stats, complete := summarizeJourney(definition, samples, s.contract)
		if !complete {
			s.unknowns++
			s.lowerResolution = true
		}
		if stats.Successful == s.contract.SamplesPerJourney && stats.OutputReplay {
			s.journeysPassed++
		}
		if stats.OutputReplay {
			s.outputReplays++
		}
		if stats.EnvelopePassed {
			s.envelopesPassed++
		}
		if stats.WallMaxMS > s.contract.WallMSLimit {
			s.wallViolations++
		}
		if stats.RSSMaxKiB > s.contract.MaxRSSKiBLimit {
			s.rssViolations++
		}
		s.stats = append(s.stats, stats)
	}
}

func validDigest(value string) bool {
	return len(value) == 71 && value[:7] == "sha256:" && stringsHex(value[7:])
}

func stringsHex(value string) bool {
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
