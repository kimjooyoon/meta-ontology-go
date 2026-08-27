# Claim dependency causality

이 실험은 일반 dependency graph 복제가 아니다. raw `.gooo`를
`syntax.ParseFile → bidir.Lower`로 canonical IR에 내린 뒤 실행 activity의
정규화된 proposition과 input/output/artifact target address를 claim으로 복원하고,
typed edge가 상태 전파와 실패 책임에 실제로 기여하는지를 검사한다. case name,
source substring, profile label은 결론을 선택하지 않는다.

## Source graph and external material

각 activity claim은 `execute(activity, inputs, output, artifact, value-program)`
proposition, proposition digest, IR에서 복원한 `wasGeneratedBy` output과 `used`
input, producer/consumer/meta-operation/proof choice, stage/step/reason provenance를
가진다. 이 fixture는 서로 다른 observed proposition 6개와 typed edge 8개
(`SUPPORTS 2`, `REQUIRES 3`, `CONTRADICTS 2`, `FAILURE_ENTAILMENT 1`)를 가진다.
graph는 source marker가 아니라 syntax/IR에서 재구성된다.

`.gooo`에는 claim topology와 observation recipe만 둔다.
`validator-contract.json`은 source를 관찰 직전에 해시해 만든 값이 아니라 고정된
외부 validator material이다. `accepted-target.gooo`와 `refuted-target.gooo`는 source
graph와 별도인 관찰 대상이며, 여섯 activity의 expected target tuple과 value program,
반대 비교가 가능한 alternate value만 제공하고 status/outcome을 선언하지 않는다.
따라서 contract의 expected bytes digest는 CI 관찰 대상에서 계산한 값이 아니다.

## Current evidence and state algebra

CI observer는 source graph와 분리된 raw target artifact를 실제로 읽은 뒤
`syntax.ParseFile → bidir.Lower`로 정확히 하나의 activity occurrence를 재구성한다.
occurrence address는 `activity:<claim_id>`인 위치 독립 semantic 주소다. raw span
start/end와 raw row digest는 별도 provenance 필드이며, lowered occurrence projection
digest와 whole-graph context digest도 서로 분리한다. 하나의 whole-file digest
match로 여섯 claim을 discharge하지 않는다. process exit가 필요한
`FAILURE_ENTAILMENT`는 CI가 실제 non-zero process를 실행해 stdout/stderr와 exit
code를 보존한 `FailureReceipt`가 추가로 있을 때만 관측된다. zero/success exit에
실패 label만 붙인 receipt는 거부된다.

profile은 fixture 목록 label일 뿐 predicate, claim state, edge activation을
선택하지 않는다. `-operation`도 `CLAIMED_INPUT/REQUEST`다. source marker,
호출자 문자열, 임의 accepted/contradiction 문자열은 current evidence가 아니다.
provider와 judge는 source bytes, target bytes, contract bytes, raw observation
bundle을 각각 재관측한다. bundle이 없거나 관련 없는 target mismatch이면
`UNKNOWN`이다. `HISTORICAL_FIXTURE`는 PASS 근거가 아니다.

| 조건 | 결과 |
| --- | --- |
| 직접 observation `UNKNOWN` | root `OPEN/DIRECT_UNKNOWN` |
| upstream `UNKNOWN` | dependent `OPEN/DEPENDENCY_BLOCKED` |
| upstream `REFUTED` on `SUPPORTS` or `REQUIRES` | dependent `OPEN/BLOCKED` |
| established upstream `DISCHARGED` + edge-specific opposite target value on `CONTRADICTS` | target `REFUTED` |
| upstream `REFUTED` + exact non-zero failure antecedent on `FAILURE_ENTAILMENT` | target `REFUTED` |
| local `EVIDENCE_ACCEPTED` and all incoming `REQUIRES` discharged | `DISCHARGED` |
| `SUPPORTS` alone | standalone discharge/refutation 금지 |

`CONTRADICTS(A,B)`는 A proposition이 성립하고 edge comparator가 B의 구조화된
expected/observed value가 반대임을 계산할 때만 B를 반증한다. A가 REFUTED라는
사실은 B를 반증하지 않는다. `FAILURE_ENTAILMENT(A,B)`는 edge에 결합된 실제
failure antecedent와 upstream A의 REFUTED 상태가 모두 있어야 B를 반증한다.
각 관계는 producer와 independent judge의 실제 classifier와 8개 truth-table
case에서 같은 대수로 실행된다(네 edge kind별 2 case, positive/reversed 또는
unknown activation 포함).

모든 transition은 before/after, stage/step/reason, local evidence digest,
provenance, upstream edge IDs와 upstream transition digest chain, 이전 head를
보존한다. resolution cause는 root digest 하나가 아니라 경로의 transition
digest 목록이다. `MaximumCausePathDepth`는 `CauseEdgeIDs`의 edge depth다.

구조적 inventory는 graph+external contract의 불일치를 claim별로 전부 기록한다.
accepted target은 structural contradiction `0/0`, refuted target은
`2/2`(ContradictionCheck, FailureEntailmentCheck)이며 이는 re-derived expected
inventory를 분모로 삼는 runtime edge observation이나 claim state가 아니다.
unknown target은 현재 관측 inventory가 없으므로 분모 `0`이며
`OBSERVE/structural-inventory/NO_CURRENT_TARGET_OBSERVATION_EXPECTED_INVENTORY_ZERO`
좌표를 남긴다. inventory row는 claim/proposition, procedure ID, target
occurrence/address, expected/declared value program, raw row digest와 semantic
digest를 모두 결속하고 producer와 judge가 각자 다시 만든다. semantic
address/digest는 comment/whitespace 삽입에도 보존되고 raw span/row/artifact
provenance는 변할 수 있다.
missing/duplicate/additional/replacement는 각각 다른 fail-closed reason이다.

UNKNOWN은 5/8 eligible blocking edge, REFUTED는 6/8 observed causal edge와
8 eligible edge, failure receipt 부재 REFUTED는 5/8 observed edge와 8 eligible
edge, recovery는 실제 discharge에 사용한 3/8 edge와 shortest-path union 2/8을
기록한다. shortest-path union은 결정론적 허용 경로의 합집합이며 전역
cardinality-minimum 증명이 아니다. DISCHARGED resolution의 실패 책임은
`N/A`/empty, direct unknown은 자기 claim, dependency blocked/refuted는 실제
최소 원인 경로의 첫 upstream claim이다.

## Append-only recovery

recovery는 이전 UNKNOWN receipt의 raw source, contract, target artifact,
observation bundle/evidence를 다시 관측하고 graph, transitions, resolutions,
metrics, decision 전체를 재계산한다. prior receipt digest, previous transition
head, claim states, source/target/evidence digests가 모두 맞을 때만 기존 12개
transition을 보존하고 6개를 append한다. 새 ledger를 만들지 않는다. recovery
chain은 `12 preserved + 6 appended`, append-only chain `1`이다.

## Independent judge and interventions

별도 judge는 producer package와 expectedClaims/expectedEdges를 import하지 않고
raw `.gooo`, source-derived graph, raw target bytes, contract, observation bundle,
prior ledger에서 결과를 독립 재구성한다. CI 보고 분모는 source reconstruction
`1/1`, producer package import `0/1`이다.

네 개의 intervention은 원인을 분리한다.

* source-only: 같은 raw evidence를 고정하고 `VALUE_PROGRAM`만 바꾼다.
* observation-only: source semantics를 고정하고 external observation 유무만 바꾼다.
* edge-only: source value가 선언한 typed edge만 바꾼다.
* comment-only: 유효 target의 raw bytes만 바꾼다. raw artifact/span/provenance는
  변하지만 semantic address/occurrence digest, claim states, transition digest
  vector, decision을 보존해야 한다. 이 보존 회귀는 `1/1`이다.

동일 source/target/contract에서 profile label만 바꾼 case도 decision과 claim
transition 의미를 바꾸지 않는다. 반면 관련 없는 artifact mismatch, claim target
교체, edge kind 교체, zero-exit failure label은 fail closed한다. repository
효과는 tracked+untracked pre/post snapshot과 실제 output path를 receipt에 결합한다.
무변경은 write authority 부재를 증명하지 않으므로 별도 capability/permission
evidence가 없으면 authority는 `TRANSIENT_WRITE_AUTHORITY_UNKNOWN`으로 낮춘다.
현재 observation 범위는 `NET_REPOSITORY_STATE_UNCHANGED`이고 semantic promotion
authorization은 false다.

고정 CI 회귀 분모는 truth algebra 8, evidence provenance 3, authority 3, prior
tamper 3, path metric 2, owner applicability 3, observation binding negative 7,
structural inventory negative 4(누락/중복/추가/치환), semantic occurrence 6,
raw provenance binding 6, comment-only semantic preservation 1이다. invalid/
duplicate/comment-like `.gooo` target은 canonical parse/lower 실패로 evidence가
되지 않는다. 구조적 모순 지표는 accepted `0/0`, refuted `2/2`로 고정한다.

## Principles and limits

provenance 경계는 [W3C PROV-O](https://www.w3.org/TR/prov-o/)의 entity, activity,
agent, influence 어휘를 참고했다. declared graph와 observed run event를 분리하는
방식은 [OpenLineage object model](https://openlineage.io/docs/spec/object-model/),
[run cycle](https://openlineage.io/docs/spec/run-cycle/),
[producer identity](https://openlineage.io/docs/spec/producers/)를 참고했다.
방향성 있는 shortest path와 intervention 보고의 한계는
[Stanford causal models](https://plato.stanford.edu/entries/causal-models/)의
구조적 관점과만 연결되며, 이 실험은 통계적 인과나 외부 세계의 truth를 주장하지
않는다.

반증 조건은 명확하다. judge가 source-derived proposition/target, typed edge,
raw observation, prior head, transition provenance, state algebra, cause path,
effect snapshot 중 하나라도 producer receipt-only 정보로 수용하면 실패한다.
SUPPORTS/REQUIRES가 upstream refutation을 dependent refutation으로 바꾸거나,
UNKNOWN이 evidence 없이 discharge되거나, profile/source marker가 predicate를
선택하거나, comment-only가 semantic digest를 바꾸면 이 실험은 반증된다. graph는
고정된 acyclic six-node fixture의 closed-world reconstruction일 뿐 외부 데이터의
진실성·완전성·실행 정확성·네 edge type의 보편성을 보장하지 않는다.
