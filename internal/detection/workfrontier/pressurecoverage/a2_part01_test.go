package pressurecoverage

const (
	a2Snapshot   = "sha256:875f886774b6c7127f6b59ee5cb5facaf5825a36f708900ba63235d7db2e9b8f"
	a2Policy     = "sha256:a4d888b25b683488a6751d9dcc487043002be60441122fe8a87afc54b809fa49"
	a2Registry   = "sha256:e0d1d311e52cc85a3ff82cf7b49b299fa7dfa7bcf876eb0995e838188dd24c57"
	a2Toolchain  = "sha256:9f2f29d60c221e56bf389b6721e08875db5f7d6b14b30d9c25b8fc73e6908cb2"
	a2InputHash  = "sha256:c6dbf237e8c44a55c79836153a6880fabdfaaa3d8e872dc07e99e337cd2e4fd3"
	a2ResultHash = "sha256:620f8ba049427fcc6dff30f4592ee10acbdaf931d5075056b9e35128bb41afa5"
	a2ReplayHash = "sha256:de2a612f99485c5711b709dc804e4d289c88a1d283b9e55427691b8ce2a29a78"
)

type a2PrecedenceCase struct {
	name  string
	bind  bool
	edit  func(*Input)
	wantD Decision
	wantR Reason
}
