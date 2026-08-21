package main

import "time"

const guardianObserverFreshnessWindow = 10 * time.Minute

func validObserverFreshness(observedAt, validUntil *string, now time.Time) bool {
	if observedAt == nil || validUntil == nil {
		return false
	}
	observed, err := time.Parse(time.RFC3339Nano, *observedAt)
	if err != nil {
		return false
	}
	expires, err := time.Parse(time.RFC3339Nano, *validUntil)
	if err != nil {
		return false
	}
	return expires.Sub(observed) == guardianObserverFreshnessWindow && !observed.After(now) && now.Before(expires)
}
