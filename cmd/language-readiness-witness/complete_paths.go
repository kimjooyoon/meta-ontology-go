package main

func completePaths(cfg config, target string) []string {
	paths := []string{cfg.input, target}
	if cfg.conceptOperationBinding != "" {
		paths = append(paths, cfg.conceptOperationBinding)
	}
	if cfg.conceptOperationInputDir != "" {
		paths = append(paths, cfg.conceptOperationInputDir)
	}
	optional := []string{
		cfg.promotion,
		cfg.guarded,
		cfg.useCases,
		cfg.syntax,
		cfg.diagnostic,
		cfg.packageRuntime,
		cfg.toolchainCLI,
		cfg.toolchainFormatFix,
		cfg.toolchainLSP,
		cfg.toolchainConformance,
		cfg.toolchainRelease,
	}
	for _, path := range optional {
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}
