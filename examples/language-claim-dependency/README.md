# Claim dependency causality fixture

이 fixture는 실행 proposition 6개와 canonical IR에서 복원한 typed edge 8개를
사용해 직접 `UNKNOWN`과 dependency-blocked `OPEN`을 분리한다.

| 입력 | current evidence | 상태 결과 | 실제 causal / eligible |
| --- | --- | --- | ---: |
| `unknown.gooo` | availability, no raw observation → 6 UNKNOWN | direct 1 OPEN, blocked 5 OPEN | 5 / 8 |
| `refuted.gooo` | target observation → 4 accepted + 2 unknown | 4 DISCHARGED, dependency 2 REFUTED | 6 / 8 |
| `main.gooo` + unknown receipt | acceptance → 6 current | direct local 4 + dependency 2 `DISCHARGED` | 3 / 8 recovery |

claim별 허용 shortest-path edge union은 각각 3, 5, 2이며 이는 cardinality-minimum
증명이 아니다. 최대 edge depth는 2다. `SUPPORTS 2`,
`REQUIRES 3`, `CONTRADICTS 2`, `FAILURE_ENTAILMENT 1`은 모두 artifact의 edge
metric에 표시된다. truth table은 edge kind마다 2건(positive/reversed/unknown
activation 사례), 총 8건이다.

CI observer가 실제 target bytes/output/exit/result를 raw observation으로 만들고,
provider가 이를 target path/procedure digest와 결합해 `CURRENT_EVIDENCE`를
생성한다. observation이 없으면 `UNKNOWN`이다. HISTORICAL_FIXTURE와 임의
문자열은 PASS 근거가 아니다. 별도 judge는 raw `.gooo`, evidence artifact와
raw observation을 직접 re-observe하여 producer receipt를 검증한다.

`value-intervention.gooo`는 같은 evidence에서 source value만 바꾸고,
observation-only는 같은 source에서 provider operation만 바꾸며,
`edge-intervention.gooo`는 refuting edge kind만 바꾼다. `main.gooo`의 주석은
semantic/graph digest와 decision을 보존해야 한다.
