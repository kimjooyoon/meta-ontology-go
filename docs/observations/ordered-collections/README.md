# Ordered collection development observations

These are fresh programs authored during development, based on compiler source
`801e81882829bcc4dfff200125abf13506fd5993` with the exact overlays in each producer
archive. The baseline was accepted in PR #1537 as
`f59f345cfd8ca7a93e1eb30d2dae698837a28a52`, with the same tree. A local result
belongs to its recorded producer; the submitted revision needs its own CI.

## Results by question

| Program | Observation | Producer |
| --- | --- | --- |
| 7013 | Integer aggregation: native 3/3, empty list and integer above 2⁵³. Prototype uses Go-style range. | 7013 |
| 7193 | Gooo for-in: native 6/6 across Integer and Text entries; ordering, Korean condition positions and input-origin connection. | 7193 |
| 7379 | Imported helper call per element, record constructor, optional Boolean lists: native 3/3; original workspace condition location retained. | 7409 |
| 7409 | Three record inputs through two forwarding activities; ordered arrays, exact int64, absence/empty distinction, detached copies; seven invalid forms rejected. | 7409 |
| 7523 | Two declared candidates. Selected construction cases 3/3; disjoint holdout 1/1. Ordered record fields compare by values; condition cites selected source expression. | 7631 |
| 7567 | Unused element names and guarded indexing: native 3/3. | 7631 |
| 7673 | Public CLI parser through ordinary check/generate, collection Go types and binding plan; unchanged-source BX check; V4 rejects many fields. | 7673 |
| 7781 | Finite evaluator sums 256 values exactly; rejects 257 values and four unsupported source loop forms. | 7781 |
| 7901 | Collection observation labels use exact ordered JSON; separate allocations compare by values and changed order differs. No source execution. | 7901 |
| 7867 | Original call location and stable ID after for-in; rename changes one call while preserving quoted text. | 7901 | Collection observation labels use exact ordered JSON; separate allocations compare by values and changed order differs. No source execution. | 7901 |
| 7867 |

Every listed program makes zero model/provider calls. These observations do not
measure ranking improvement, independent adoption, general correctness or a
language-wide completeness percentage. Source locations and static dependencies
have their own observations. Collection dependencies join possible elements and
retain explicit partial-analysis reasons.

## First failures retained

- 7139: two unbound entries were given without an explicit entry selector. No
  program ran. 7193 uses separate explicit entries and different source cases.
- 7301 in producer 7409: the authored `value_holdout` spelling was invalid. No
  candidate ran. 7523 uses the implemented `holdout_value_case` syntax with a
  different threshold, source identity and input set.
- 7631: its test invoked the intentionally deferred generic parser seam. The
  public CLI uses `EntityFieldsCLIParser`. 7673 exercises that adapter on a new
  source identity. The V5 BX check in the first test had already succeeded.

Early record field labels in the saved 7523 result retain the prototype Go display.
The submitted source uses ordered JSON labels, checked separately by 7901. The
earlier raw result is preserved unchanged.

All original logs are retained in this directory. Producer 7409 and producer
7631 therefore have a failing process exit even though their other named fresh
tests passed. No failed process is counted as a successful batch. The repaired
7139 test in the submitted source was not locally rerun; its result belongs to
the submitted revision's CI.

## Artifacts

- `producers.tar.gz`: exact replacement files, producer manifests, process-end
  records and the common baseline. Overlay paths are repository relative.
- `*.log`: original first-run output, including failures.
- `observations/*.json`: source, cases and recorded construction/native output
  for 7013, 7193, 7379 and 7523.
- `profile-v5.txt`: exact UTF-8 profile text; its SHA-256 is
  `edac62212632b1f61390c95e624d0671c718b357cd5aa42ff0eef86358d90899`.

Later producers include later implementation changes. A passing earlier test
does not establish that all later revisions pass it. Original local programs
were not rerun. No training, historical cohort execution or new worktree was
used. Temporary native outputs were removed by the runtime's existing cleanup.

## 입력 규칙과 내부 리팩토링 후속 확인

목록 지원에 맞춘 단일 값·목록 입력 규칙과 출처 분석 메서드의 분리는
[후속 관찰](../ordered-cardinality/README.md)에 기록했다. 앞선 공개 CI 실패와
새 소스의 관찰 결과를 각각 유지한다.
