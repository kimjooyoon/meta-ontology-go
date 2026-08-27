# Claim dependency causality fixture

이 fixture는 실행 proposition 6개와 canonical IR에서 복원한 typed edge 8개를
사용해 직접 `UNKNOWN`과 dependency-blocked `OPEN`을 분리한다.

| 입력 | 관측 | 상태 결과 | observed causal / eligible |
| --- | --- | --- | ---: |
| `unknown.gooo` | raw observation 없음 | direct 1 + blocked 5 OPEN | 5 / 8 |
| `refuted.gooo` + `refuted-target.gooo` | 외부 contract, claim 4개, `CONTRADICTS` 2개, 실제 edge-specific non-zero failure 1개 | 4 DISCHARGED, 2 dependency REFUTED | 6 / 8 |
| `refuted.gooo` + failure receipt 없음 | 같은 source/target, failure edge 비활성 | 4 DISCHARGED, 1 REFUTED, 1 OPEN | 5 / 8 |
| `main.gooo` + unknown receipt | append-only recovery | 6 DISCHARGED | 3 / 8 recovery |

`validator-contract.json`은 fixed expected target/value material이며 source를
관찰 직전에 해시해 만든 순환 expected 값이 아니다. `.gooo` source는 graph와
recipe만 선언한다. observer가 source graph와 분리된 target artifact bytes를
실제로 읽고 claim별 target-specific material을 비교한다. profile은 fixture label
일 뿐 상태나 predicate를 고르지 않는다.

`CONTRADICTS`는 established upstream과 edge-specific structured opposite value가
함께 있을 때만 target을 `REFUTED`로 만든다. `FAILURE_ENTAILMENT`는 정확한 edge에
결합된 실제 non-zero process receipt와 upstream `REFUTED`가 함께 있을 때만
전파된다. 관련 없는 artifact mismatch, missing observation, reversed edge,
claim/proposition/target/edge-kind tamper는 fail closed한다.

각 claim에는 proposition digest, input/output/artifact target tuple, procedure와
output digest, stage/step/reason이 있다. transition은 append-only이며 root와
upstream transition digest chain을 보존한다. `MaximumCausePathDepth`는 edge 수다.
shortest-path union은 cardinality-minimum 증명이 아니다. 별도 judge는 producer
package와 expected graph를 import하지 않고 raw `.gooo`, target bytes, contract,
bundle, prior ledger에서 독립 재구성한다.
