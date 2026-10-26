package bodylibrary

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type GenerationOptions struct {
	StageExports []StageExport
	Origin       *GenerationContext
}

// GenerationContext describes observations supplied by the generating frontend.
// The library compiler supplies the lowered-source and typed-plan digests itself.
type GenerationContext struct {
	CompilerVersion      string             `json:"compiler_version"`
	CompilerSourceSHA    string             `json:"compiler_source_sha"`
	CompilerSourceStatus string             `json:"compiler_source_status"`
	GoVersion            string             `json:"go_version"`
	WorkspaceManifestSHA string             `json:"workspace_manifest_sha256"`
	WorkspaceImageSHA    string             `json:"workspace_image_sha256"`
	Sources              []GenerationSource `json:"sources"`
}

type GenerationSource struct {
	PackagePath string `json:"package_path"`
	Filename    string `json:"filename"`
	SourceSHA   string `json:"source_json_string_sha256"`
}

func appendGenerationOrigin(source, sourceDigest string, plan bodyexecution.CompositionPlan, options GenerationOptions) (string, error) {
	entryID := ""
	for _, activity := range plan.Activities {
		if activity.Name == plan.EntryActivity {
			entryID = activity.ID
		}
	}
	origin := struct {
		Schema       string            `json:"schema"`
		EntryID      string            `json:"entry_activity_id"`
		SourceSHA    string            `json:"lowered_source_sha256"`
		TypedPlanSHA string            `json:"typed_plan_sha256"`
		Context      GenerationContext `json:"context"`
		StageExports []StageExport     `json:"stage_exports"`
	}{"gooo/library-generation-origin/v1", entryID, sourceDigest, plan.TypedPlanSHA256, *options.Origin, options.StageExports}
	raw, err := json.Marshal(origin)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("urn:gooo:library:generation-origin:%x", sha256.Sum256(raw))
	declaration := fmt.Sprintf("\n//gooo:generated:start id=%q kind=\"generation-origin\"\n"+
		"const GoooGenerationOrigin = %s\n//gooo:generated:end id=%q kind=\"generation-origin\"\n",
		id, strconv.Quote(string(raw)), id)
	return formatLibraryProjection(source + declaration)
}
