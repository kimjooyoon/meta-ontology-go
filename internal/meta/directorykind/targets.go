package directorykind

import "sort"

func kindTargets(source SourceMetrics) ([]SourceIndicator, int, int) {
	targets := make([]SourceIndicator, 0)
	applicable, rootExemptions := 0, 0
	for _, indicator := range source.Meta.Indicators {
		if indicator.Subject == "." && indicator.Applicability == "NOT_APPLICABLE" {
			rootExemptions++
		}
		if indicator.MetaOperation != "separate-directory-kinds" ||
			indicator.Applicability != "APPLICABLE" || !indicator.Blocking {
			continue
		}
		applicable++
		if !indicator.Satisfied {
			targets = append(targets, indicator)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Subject < targets[j].Subject })
	return targets, applicable, rootExemptions
}

func directoryMetric(source SourceMetrics, subject string) (SourceDirectory, bool) {
	for _, directory := range source.Directories {
		if directory.Path == subject {
			return directory, true
		}
	}
	return SourceDirectory{}, false
}
