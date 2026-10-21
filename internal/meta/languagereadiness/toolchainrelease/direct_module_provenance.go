package toolchainrelease

import (
	"net/url"
	"slices"
	"sort"
)

func directModuleProvenance(result DirectModuleImpact) map[string]any {
	graph := []any{}
	used := []string{}
	paths := make([]string, 0, len(result.Inputs))
	for name := range result.Inputs {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	for _, name := range paths {
		digest := result.Inputs[name]
		id := repositoryFileID(name, digest)
		file := map[string]any{"@id": id, "@type": "prov:Entity",
			"gooo:path": name, "gooo:sha256": digest}
		if slices.Contains(result.HistoricalScope, name) {
			headID := "urn:gooo:committed-file:" + result.ArchiveHead + ":" + url.PathEscape(name)
			graph = append(graph, map[string]any{"@id": headID, "@type": "prov:Entity",
				"gooo:commit": result.ArchiveHead, "gooo:path": name, "gooo:sha256": digest})
			origins := []string{headID}
			for _, rule := range result.HistoricalRules {
				if rule.Path == name {
					origins = append(origins, rule.ID)
					break
				}
			}
			file["prov:wasDerivedFrom"] = provReferences(origins)
			file["gooo:referenceRole"] = "historical"
		}
		graph = append(graph, file)
		used = append(used, id)
	}
	for _, rule := range result.ActiveRules {
		graph = append(graph, map[string]any{"@id": rule.ID, "@type": "prov:Entity",
			"gooo:path": rule.Path, "gooo:prefix": rule.Prefix,
			"prov:wasDerivedFrom": map[string]string{
				"@id": repositoryFileID(directModuleSourcePath, result.Inputs[directModuleSourcePath])}})
		used = append(used, rule.ID)
	}
	for _, rule := range result.HistoricalRules {
		graph = append(graph, map[string]any{"@id": rule.ID, "@type": "prov:Entity",
			"gooo:path": rule.Path, "gooo:sha256": rule.SHA256,
			"prov:wasDerivedFrom": map[string]string{
				"@id": repositoryFileID(directModuleSourcePath, result.Inputs[directModuleSourcePath])}})
		used = append(used, rule.ID)
	}
	if result.Download.Path != "" {
		moduleID := "urn:gooo:module-download:" + result.Module + "@" + result.ProposedVersion
		graph = append(graph, map[string]any{"@id": moduleID, "@type": "prov:Entity",
			"gooo:ref": result.Download.Ref, "gooo:vcs": result.Download.VCS,
			"gooo:source_url": result.Download.SourceURL, "gooo:sum": result.Download.Sum,
			"gooo:go_mod_sum":     result.Download.GoModSum,
			"gooo:license_sha256": result.Download.LicenseSHA})
		used = append(used, moduleID)
	}
	activity := "urn:gooo:direct-module-impact:observation:" + result.PlanSHA256
	graph = append(graph, map[string]any{"@id": activity, "@type": "prov:Activity",
		"prov:used": provReferences(used), "gooo:referenceScanScope": result.ScanScope,
		"gooo:trackedPathsSHA256":    result.TrackedPathsSHA,
		"gooo:trackedFilesScanned":   result.TrackedFiles,
		"gooo:trackedReferenceFiles": result.ReferenceFiles})
	graph = append(graph, map[string]any{"@id": "urn:gooo:direct-module-impact:decision:" + result.PlanSHA256,
		"@type": "prov:Entity", "prov:wasGeneratedBy": map[string]string{"@id": activity},
		"gooo:status": result.Decision.Status, "gooo:cochangeScope": result.Outputs,
		"gooo:outsideScope":        result.OutsideScope,
		"gooo:historicalScope":     result.HistoricalScope,
		"gooo:archiveHeadCommit":   result.ArchiveHead,
		"gooo:archiveRoots":        result.ArchiveRoots,
		"gooo:activeVersionRules":  result.ActiveRules,
		"gooo:historicalFileRules": result.HistoricalRules})
	return map[string]any{"@context": map[string]string{"prov": "http://www.w3.org/ns/prov#",
		"gooo": "https://github.com/kimjooyoon/meta-ontology-go/ns#"}, "@graph": graph}
}

func provReferences(ids []string) []any {
	references := make([]any, 0, len(ids))
	for _, id := range ids {
		references = append(references, map[string]string{"@id": id})
	}
	return references
}

func repositoryFileID(name, digest string) string {
	return "urn:gooo:repository-file:" + url.PathEscape(name) + ":" + digest
}
