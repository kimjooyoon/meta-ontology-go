package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// GatewayPolicy is a declarative outbound host boundary. It describes an
// observation policy only; it does not authenticate a workload or grant a
// capability by itself.
type GatewayPolicy struct {
	AllowedHosts []string `json:"allowed_hosts"`
}

// GatewayObservationStatus is deliberately separate from identity freshness
// and authorization. A caller must combine these observations through an
// explicit policy decision outside this type.
type GatewayObservationStatus string

const (
	GatewayObservationAllowed GatewayObservationStatus = "ALLOWED"
	GatewayObservationDenied  GatewayObservationStatus = "DENIED"
	GatewayObservationUnknown GatewayObservationStatus = "UNKNOWN"
)

// GatewayObservation is detached evidence for one host lookup.
type GatewayObservation struct {
	PolicyDigest string                   `json:"policy_digest"`
	Host         string                   `json:"host"`
	Status       GatewayObservationStatus `json:"status"`
	Reason       string                   `json:"reason"`
}

// Digest identifies the host policy independently of allowlist order.
func (policy GatewayPolicy) Digest() string {
	hosts := append([]string(nil), policy.AllowedHosts...)
	sort.Strings(hosts)
	sum := sha256.Sum256([]byte(strings.Join(hosts, "\x00")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ObserveHost records whether a syntactically acceptable host is explicitly
// present in the allowlist. Invalid or empty hosts remain UNKNOWN rather than
// being silently denied or promoted to an authorization result.
func (policy GatewayPolicy) ObserveHost(host string) GatewayObservation {
	normalized, valid := normalizeGatewayHost(host)
	observation := GatewayObservation{PolicyDigest: policy.Digest(), Host: normalized}
	if !valid {
		observation.Status = GatewayObservationUnknown
		observation.Reason = "GATEWAY_HOST_INVALID"
		return observation
	}
	for _, allowed := range policy.AllowedHosts {
		if candidate, ok := normalizeGatewayHost(allowed); ok && candidate == normalized {
			observation.Status = GatewayObservationAllowed
			observation.Reason = "GATEWAY_HOST_ALLOWLISTED"
			return observation
		}
	}
	observation.Status = GatewayObservationDenied
	observation.Reason = "GATEWAY_HOST_NOT_ALLOWLISTED"
	return observation
}

func normalizeGatewayHost(host string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(host))
	normalized = strings.TrimSuffix(normalized, ".")
	if normalized == "" || strings.ContainsAny(normalized, "/:*? \t\r\n") {
		return "", false
	}
	return normalized, true
}
