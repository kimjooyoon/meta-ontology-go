package toolchainlsp

import "maps"

func evaluateRuntime() (map[string]observation, runtimeStats, error) {
	firstRaw, first, err := executeServerSession()
	if err != nil {
		return nil, runtimeStats{}, err
	}
	replayRaw, _, err := executeServerSession()
	if err != nil {
		return nil, runtimeStats{}, err
	}
	observations, stats := observeServer(firstRaw, first, replayRaw)
	couplingObservations, couplingStats, err := observeCoupling()
	if err != nil {
		return nil, runtimeStats{}, err
	}
	maps.Copy(observations, couplingObservations)
	stats.NavigationPaths += couplingStats.NavigationPaths
	stats.FailClosedPaths += couplingStats.FailClosedPaths
	stats.NonstandardWireFields += couplingStats.NonstandardWireFields
	stats.StaleLeaks += couplingStats.StaleLeaks
	stats.UnknownLeaks += couplingStats.UnknownLeaks
	stats.FailClosedLeaks += couplingStats.FailClosedLeaks
	return observations, stats, nil
}
