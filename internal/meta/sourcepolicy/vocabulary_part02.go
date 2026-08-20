package sourcepolicy

// Dimension is also the stable metric ID emitted into CI artifacts.
type Dimension string

const (
	DimensionGoFileLines       Dimension = "gooo.metric.source.go-file-lines.v1"
	DimensionGoooFileLines     Dimension = "gooo.metric.source.gooo-file-lines.v1"
	DimensionFunctionLines     Dimension = "gooo.metric.source.function-lines.v1"
	DimensionGoFiles           Dimension = "gooo.metric.source.go-files.v1"
	DimensionGoooFiles         Dimension = "gooo.metric.source.gooo-files.v1"
	DimensionGoLines           Dimension = "gooo.metric.source.go-lines.v1"
	DimensionGoooLines         Dimension = "gooo.metric.source.gooo-lines.v1"
	DimensionDirectFiles       Dimension = "gooo.metric.layout.direct-files.v1"
	DimensionDirectFolders     Dimension = "gooo.metric.layout.direct-folders.v1"
	DimensionRecursiveFiles    Dimension = "gooo.metric.layout.recursive-files.v1"
	DimensionRecursiveFolders  Dimension = "gooo.metric.layout.recursive-folders.v1"
	DimensionDirectEntries     Dimension = "gooo.metric.layout.direct-entries.v1"
	DimensionDirectoryKinds    Dimension = "gooo.metric.layout.entry-kinds.v1"
	DimensionRefactorDuplicate Dimension = "gooo.metric.refactor.duplicate-body.v1"
	DimensionRefactorReturn    Dimension = "gooo.metric.refactor.single-return.v1"
	DimensionRefactorAssign    Dimension = "gooo.metric.refactor.assign-return.v1"
	DimensionFixDelta          Dimension = "gooo.metric.conformance.go-fix-delta.v1"
	DimensionToolchain         Dimension = "gooo.metric.conformance.toolchain.v1"
)
