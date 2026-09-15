# Toolchain LSP readiness

`TOOLCHAIN-LSP` is satisfied only by an exact-head `gooo/toolchain-lsp-report/v1` receipt with all fixed denominators closed.

Required values are 22/22 cases, 8/8 advertised capabilities, 7/7 read features, 3/3 diagnostic paths, 3/3 navigation paths, 2/2 symbol paths, 1/1 semantic-token path, 1/1 UTF-16 replay, 1/1 transcript replay, and 5/5 bounded fail-closed paths.

All 18 guardrails must be zero. In particular, unknown, stale, cancelled, and fail-closed upstream decisions cannot be promoted to a fixed point and cannot emit navigation. Repository writes and mutation authorities are both zero.

With the unchanged `gooo/self-improving-language-obligations/v1` denominator of 24, exact promotion changes completed obligations from 22 to 23 and readiness from 9166 to 9583 basis points. The transition is accepted only when CI emits `IMPROVED`, `+1`, `+417`, zero regressions, and zero unresolved evidence.
