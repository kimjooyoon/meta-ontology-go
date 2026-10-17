package cliobservations

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestCLIProducerSourceSettingsFresh1906(t *testing.T) {
	head := strings.Repeat("a", 40)
	for _, tc := range []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{"clean", []debug.BuildSetting{{Key: "vcs.revision", Value: head}, {Key: "vcs.modified", Value: "false"}}, head},
		{"dirty", []debug.BuildSetting{{Key: "vcs.revision", Value: head}, {Key: "vcs.modified", Value: "true"}}, "UNBOUND_LOCAL_SOURCE"},
		{"revision-only", []debug.BuildSetting{{Key: "vcs.revision", Value: head}}, "UNBOUND_LOCAL_SOURCE"},
		{"clean-only", []debug.BuildSetting{{Key: "vcs.modified", Value: "false"}}, "UNBOUND_LOCAL_SOURCE"},
		{"short-revision", []debug.BuildSetting{{Key: "vcs.revision", Value: "abc"}, {Key: "vcs.modified", Value: "false"}}, "UNBOUND_LOCAL_SOURCE"},
		{"no-build-settings", nil, "UNBOUND_LOCAL_SOURCE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := cliProducerSourceFromSettings(tc.settings); got != tc.want {
				t.Fatal("actual binary metadata was relabeled", got, tc.want)
			}
		})
	}
}

func TestNativeProducerBindingScopesFresh1906(t *testing.T) {
	head, other := strings.Repeat("a", 40), strings.Repeat("b", 40)
	for _, tc := range []struct {
		name, actual, built, work string
		pass                      bool
	}{
		{"original", head, head, head, true},
		{"wrong-runtime", other, head, head, false},
		{"wrong-work", other, other, head, false},
		{"unbound-derived", "UNBOUND_LOCAL_SOURCE", "UNBOUND_LOCAL_SOURCE", head, true},
		{"forged-original", head, "UNBOUND_LOCAL_SOURCE", head, false},
		{"missing-build", head, "", head, false},
		{"local-bound", head, head, "", true},
		{"local-unbound", "UNBOUND_LOCAL_SOURCE", "UNBOUND_LOCAL_SOURCE", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if textBoundProducerMatches1906(tc.actual, tc.built, tc.work) != tc.pass {
				t.Fatal("producer and admitted-work scopes were conflated", tc)
			}
		})
	}
}
