# Claim dependency causality fixture

이 fixture는 실행 proposition 6개와 canonical IR에서 복원한 typed edge 8개를
사용해 직접 `UNKNOWN`과 dependency-blocked `OPEN`을 분리한다.

| 입력 | current evidence | 상태 결과 | 실제 causal / eligible |
| --- | --- | --- | ---: |
| `unknown.gooo` | availability → 6 UNKNOWN | direct 1 OPEN, blocked 5 OPEN | 5 / 8 |
| `refuted.gooo` | contradiction → root explicit, 5 UNKNOWN | direct 1 REFUTED, dependency 2 REFUTED, 3 OPEN | 7 / 8 |
| `main.gooo` + unknown receipt | acceptance → 6 current | direct local 4 + dependency 2 `DISCHARGED` | 3 / 8 recovery |

claim별 허용 shortest-path edge union은 각각 3, 5, 2이며 이는 cardinality-minimum
증명이 아니다. 최대 edge depth는 2다. `SUPPORTS 2`,
`REQUIRES 3`, `CONTRADICTS 2`, `FAILURE_ENTAILMENT 1`은 모두 artifact의 edge
metric에 표시된다. truth table은 edge kind마다 positive/negative 2건, 총 8건이다.

CI provider가 raw artifact path/bytes/digest, operation request, 관측 절차, per-claim proposition digest,
repository pre/post snapshot, output path, `contents:read` capability를 담은
`CURRENT_EVIDENCE` receipt를 생성한다. HISTORICAL_FIXTURE와 임의 문자열은
PASS 근거가 아니다. 별도 judge는 raw `.gooo`와 evidence artifact를 직접
re-observe하여 producer receipt를 검증한다.

`value-intervention.gooo`는 같은 evidence에서 source value만 바꾸고,
observation-only는 같은 source에서 provider operation만 바꾸며,
`edge-intervention.gooo`는 refuting edge kind만 바꾼다. `main.gooo`의 주석은
semantic/graph digest와 decision을 보존해야 한다.
