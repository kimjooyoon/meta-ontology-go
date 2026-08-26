package artifact

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"

func BuildImprovement(
	beforeRaw []byte, before, after Receipt, headSHA string,
	baseline BaselineReference,
) (ImprovementArtifact, error) {
	result := sealImprovement(ImprovementArtifact{
		Schema: ImprovementArtifactSchema, Producer: improvementProducer,
		Consumer: improvementConsumer, MetaOperation: improvementOperation,
		Baseline:             FoundationBaseline(baseline),
		HeadSHA:              headSHA,
		BeforeArtifactDigest: before.ArtifactDigest,
		AfterArtifactDigest:  after.ArtifactDigest,
		BeforeSnapshotDigest: before.Snapshot.Digest,
		AfterSnapshotDigest:  after.Snapshot.Digest,
		Transition: improvement.Evaluate(
			before.TransitionInput, after.TransitionInput,
		),
		RepositoryWrites: 0,
	})
	if err := ValidateImprovement(beforeRaw, before, after, result); err != nil {
		return ImprovementArtifact{}, err
	}
	return result, nil
}
