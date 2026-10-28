package syntaxregistration

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func validProfileRequirement(requirement string) bool {
	switch requirement {
	case "", "ordinary", "text-fields", "implicit-ports", "ordered-collections",
		"entity-string", "entity-boolean", "entity-integer", "entity-optional",
		"entity-many", "entity-text-bound":
		return true
	default:
		return false
	}
}

// Eligibility comes from the validated kernel capabilities. The source policy
// consumes this observed fact; the host checks its selected result separately.
func profileSatisfiesRequirement(profile, requirement string) bool {
	if requirement == "" {
		return true
	}
	switch requirement {
	case "ordinary", "text-fields", "implicit-ports", "ordered-collections":
		return profile == requirement
	}
	support, known := syntax.EntityFieldsSupportForProfileID(profile)
	if !known {
		return false
	}
	capabilities, err := support.Profile.Capabilities()
	if err != nil {
		return false
	}
	switch requirement {
	case "entity-string":
		return true
	case "entity-boolean":
		return capabilities.Boolean
	case "entity-integer":
		return capabilities.Integer
	case "entity-optional":
		return capabilities.Optional
	case "entity-many":
		return capabilities.Many
	case "entity-text-bound":
		return capabilities.TextBound
	default:
		return false
	}
}
