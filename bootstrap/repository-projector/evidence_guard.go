package main

import (
	"fmt"
	"sort"
	"strings"
)

func requireBlockingZero(report evidence) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			details := blockingSubjectDetails(report.Subjects, metric.ID)
			if details != "" {
				return fmt.Errorf("blocking indicator %s=%d subjects=%s", metric.ID, metric.Value, details)
			}
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}

func blockingSubjectDetails(subjects []subject, indicatorID string) string {
	details := make([]string, 0)
	for _, item := range subjects {
		if item.Indicator != indicatorID || item.Physical == "" {
			continue
		}
		details = append(details, fmt.Sprintf("%s(%d>%d)", item.Physical, item.Value, item.Limit))
	}
	sort.Strings(details)
	return strings.Join(details, ",")
}
