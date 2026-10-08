# 0.6.15 native candidate observations

Plan and original raw fixtures were frozen at `9eb4513d` before implementation.
The new tests first failed because the new witness was absent. The clean candidate
`848f85de6a9088b7396a3b4aeb3dce32ff670dd5` then passed the retained release tests,
31 changed-evidence cases, release race checks, focused CLI regressions and vet.

The macOS arm64 platform witness built twice, compared the resulting bytes,
executed the existing native profiles plus the new rejected-fill profile, and
compared archive bytes. Its exact receipt and three new raw outputs are retained
in `observations.tar.gz`. The executable used below came from that candidate
archive. `build.txt` binds it to clean source, Go 1.27.1 and the unchanged public
decision runtime. This local candidate record does not imply a published release.

| Observation | Budgeted attempts | Rejected | Native combinations | Later evaluation |
| --- | ---: | ---: | ---: | ---: |
| budget 3 | 3 | 2 | 1 | 1/4 |
| budget 5 | 5 | 2 | 3 | 4/4 |
| budget saved replay | 5 historical | 2 historical | 3 historical | 4/4 |
| mixed fixed | 40 | 16 | 24 | 4/4 |
| mixed own model | 5 | 2 | 3 | 4/4 |
| mixed saved replay | 5 historical | 2 historical | 3 historical | 4/4 |

The workbench's `experiments/fill-rejection-observe` reader was reused with its
input directory parameterized. It independently recounted original evaluation
expectations using exact JSON numbers, checked rejection accounting and compared
selected source between fixed/model ordering and saved replays. `summary.json`
contains the result. All six observations retain the original source and cases.
Replay history counts describe saved construction; they are not new searches.

The unchanged graph QAT model made one prediction in the mixed fresh run
(39,291 ns in this observation). Its metadata SHA-256 is
`3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202` and weights
`76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
It orders record choices; source-fill assignments remain deterministic.
Saved replay made zero new predictions. This uses the all-data demonstration
model and related finite programs. No new training or independent accuracy study.

Raw process timing files are retained as observations, not controlled benchmarks
or host-wide CPU utilization measurements. Native combinations can contain two
fresh executions and are distinct from child-process counts. The original
feature and earlier release records remain available unchanged.
