package provenance

import "testing"

func TestGatewayPolicyDigestIgnoresAllowlistOrder(t *testing.T) {
	left := GatewayPolicy{AllowedHosts: []string{"api.example.com", "registry.example.com"}}
	right := GatewayPolicy{AllowedHosts: []string{"registry.example.com", "api.example.com"}}
	if left.Digest() != right.Digest() {
		t.Fatalf("policy digests differ: %q != %q", left.Digest(), right.Digest())
	}
}

func TestGatewayPolicyDigestNormalizesCanonicalHosts(t *testing.T) {
	left := GatewayPolicy{AllowedHosts: []string{"API.EXAMPLE.COM.", "registry.example.com"}}
	right := GatewayPolicy{AllowedHosts: []string{"registry.example.com.", "api.example.com"}}
	if left.Digest() != right.Digest() {
		t.Fatalf("canonical policy digests differ: %q != %q", left.Digest(), right.Digest())
	}
}

func TestGatewayPolicyDigestRetainsInvalidEntryProvenance(t *testing.T) {
	base := GatewayPolicy{AllowedHosts: []string{"api.example.com"}}
	withInvalid := GatewayPolicy{AllowedHosts: []string{"api.example.com", "https://registry.example.com"}}
	if base.Digest() == withInvalid.Digest() {
		t.Fatal("invalid allowlist entry disappeared from policy provenance")
	}
}

func TestGatewayPolicyObservesAllowedHost(t *testing.T) {
	observation := (GatewayPolicy{AllowedHosts: []string{"api.example.com"}}).ObserveHost("API.EXAMPLE.COM.")
	if observation.Status != GatewayObservationAllowed || observation.Reason != "GATEWAY_HOST_ALLOWLISTED" {
		t.Fatalf("observation = %#v", observation)
	}
	if observation.PolicyDigest == "" || observation.Host != "api.example.com" {
		t.Fatalf("observation identity = %#v", observation)
	}
}

func TestGatewayPolicyKeepsDeniedAndInvalidDistinct(t *testing.T) {
	policy := GatewayPolicy{AllowedHosts: []string{"api.example.com"}}
	if observation := policy.ObserveHost("other.example.com"); observation.Status != GatewayObservationDenied {
		t.Fatalf("denied observation = %#v", observation)
	}
	if observation := policy.ObserveHost("https://api.example.com"); observation.Status != GatewayObservationUnknown || observation.Reason != "GATEWAY_HOST_INVALID" {
		t.Fatalf("invalid observation = %#v", observation)
	}
}
