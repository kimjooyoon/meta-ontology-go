package verticalsliceclosureshadow

func applyBoundaryLinks(result *BoundaryResult, artifact artifactEnvelope,
	artifacts map[string]artifactEnvelope, raw map[string][]byte,
	results map[string]*BoundaryResult, head string) {
	if result.Status != StatusSatisfied {
		return
	}
	switch result.ID {
	case "syntax":
		result.LinksSatisfied = 1
	case "semantics":
		if dependencyReady(result, results["syntax"]) {
			result.LinksSatisfied = semanticLinks(artifact, artifacts["syntax"], raw["syntax"])
		}
	case "binding":
		if dependencyReady(result, results["semantics"]) {
			result.LinksSatisfied = bindingLinks(artifact, artifacts["semantics"], raw["semantics"])
		}
	case "use-cases":
		if dependencyReady(result, results["syntax"]) &&
			artifact.Source.ConceptArtifactDigest ==
				artifacts["syntax"].Source.ConceptArtifactDigest {
			result.LinksSatisfied = 1
		}
	case "toolchain":
		if dependencyReady(result, results["syntax"]) &&
			dependencyReady(result, results["semantics"]) &&
			dependencyReady(result, results["use-cases"]) {
			result.LinksSatisfied = toolchainLinks(artifact.Surfaces, head)
		}
	case "release":
		result.LinksSatisfied = releaseLinks(artifact)
	}
	if result.Status == StatusSatisfied && result.LinksSatisfied != result.LinksTotal {
		setBlocked(result, boundaryReasonLink)
	}
}

func dependencyReady(target, dependency *BoundaryResult) bool {
	if dependency != nil && dependency.Status == StatusSatisfied {
		return true
	}
	if dependency == nil || dependency.Status == StatusUnknown {
		setUnknown(target, boundaryReasonDependency, false)
	} else {
		setBlocked(target, boundaryReasonLink)
	}
	return false
}
