package toolchainrelease

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateRelativeV3Binding(prepared bodycodegen.Report, preflight *bodycodegen.TypedPathContextExport, head string) (*bodycodegen.BodyPathReceipt, error) {
	p := prepared.BodyPaths
	if prepared.ActivityID != "output-relations-v3://choose" || prepared.CompilerSourceSHA != head || p == nil || p.ContractRanking == nil || p.ModelRetention == nil ||
		p.ContractRanking.Schema != "gooo/source-relative-requirement-path-ranking/v3" || p.ContractRanking.Calls != 1 || !p.ContractRanking.Applied ||
		p.ContractRanking.PlanSHA != preflight.Context.OriginalPlanSHA || p.ContractRanking.CaseSHA != preflight.ContractCases.CaseSHA ||
		p.ContractRanking.SourceFeatures != contractdecision.SourceRelativeV3FeatureVersion ||
		p.ModelRetention.ArtifactSHA256 != preflight.ModelCompatibility.Model.ArtifactSHA256 ||
		p.ContractRanking.ModelFingerprint != strings.TrimPrefix(preflight.ModelCompatibility.Model.ModelFingerprint, "sha256:") ||
		p.ContractRanking.ConditionCount != 3 || "sha256:"+p.ContractRanking.ConditionFeatureSHA != preflight.ContractConditions.FeatureSHA ||
		p.ModelRetention.ModelFingerprint != preflight.ModelCompatibility.Model.ModelFingerprint || !reflect.DeepEqual(p.ModelContext, preflight.Context) ||
		p.Search.Selection.ModelCalls != 1 || p.Search.Selection.ExternalCalls != 0 || !p.Search.Selection.ExternalCallsKnown {
		return nil, fmt.Errorf("V3 initial source/model binding or one-call ordering differs")
	}
	for i, row := range preflight.Inputs {
		if row.InputSHA != "sha256:"+p.ContractRanking.FeatureSHA[i] {
			return nil, fmt.Errorf("V3 ranked input differs")
		}
	}
	return p, nil
}
