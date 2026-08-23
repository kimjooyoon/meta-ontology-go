package main

func completePaths(cfg config, target string) []string {
	paths := []string{cfg.input, target}
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
