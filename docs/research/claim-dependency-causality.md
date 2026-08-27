# Claim dependency causality

이 실험은 일반 dependency graph 복제가 아니다. raw `.gooo`를
`syntax.ParseFile → bidir.Lower`로 canonical IR에 내린 뒤, 실행 activity의
정규화된 proposition과 input/output/artifact target address를 claim으로
복원하고, typed edge가 state propagation과 failure responsibility에 실제로
기여하는지를 검사한다. case name이나 source substring은 결론을 선택하지
않는다.

## Source-derived contract

각 activity claim은 다음을 포함한다.

`execute(activity, inputs, output, artifact, value-program)` proposition,
그 proposition의 SHA-256 digest, IR에서 복원한 `wasGeneratedBy` output과
`used` input, producer/consumer/meta-operation/proof choice, 그리고
stage/step/reason provenance이다. fixture에는 여섯 개의 서로 다른 observed
predicate가 있으므로 claim 분모는 정직하게 `6/6`이다. IR에서 digest가
중복되면 고정 분모를 유지하지 않고 검사를 실패시킨다.

현재 graph는 `SUPPORTS 2`, `REQUIRES 3`, `CONTRADICTS 2`,
`FAILURE_ENTAILMENT 1`, 총 `8/8` eligible edge다. `CONTRADICTS`는
“from proposition established/refuted, explicitly contradicts to proposition”
방향이고, `FAILURE_ENTAILMENT`는 “from established failure, entails target
failure” 방향이다. edge 이름만으로 `REFUTED`를 만들지 않는다.

## Current evidence and algebra

`.gooo`에는 observation recipe/capability와 topology만 선언된다. CI provider가
실제 artifact path/bytes/digest와 관측 절차를 실행해 `CURRENT_EVIDENCE` receipt를
만든다. `-operation`은 `CLAIMED_INPUT/REQUEST`일 뿐이며, 그 문자열 자체는
predicate가 아니다. 외부 target 관측이 없는 `availability` 요청은 predicate를
`UNKNOWN`으로 낮춘다.
receipt는 provider, artifact path, bytes digest, observed predicate/value,
status, stage/step/reason, per-claim proposition digest, tracked/untracked
repository snapshot, output path, capability evidence를 가진다. provider와
judge는 artifact bytes를 각각 다시 읽고 parse/lower한다. `HISTORICAL_FIXTURE`나
임의 문자열은 PASS 근거가 아니다.

state 대수는 다음과 같다.

| 조건 | 결과 |
| --- | --- |
| 직접 observation `UNKNOWN` | root `OPEN/DIRECT_UNKNOWN` |
| upstream `UNKNOWN` | dependent `OPEN/DEPENDENCY_BLOCKED` |
| upstream `REFUTED` on `SUPPORTS` or `REQUIRES` | dependent `OPEN/BLOCKED` |
| upstream `REFUTED` on direction-matching `CONTRADICTS` | target `REFUTED` (local child may be `UNKNOWN`) |
| upstream `REFUTED` on direction-matching `FAILURE_ENTAILMENT` | target failure `REFUTED` (local child may be `UNKNOWN`) |
| matching local `EVIDENCE_ACCEPTED` and all incoming `REQUIRES` discharged | `DISCHARGED` |
| `SUPPORTS` only | standalone discharge 금지 |

`OPEN → DISCHARGED`는 실제 current evidence predicate가 맞을 때만 허용한다.
모든 transition은 before/after, stage/step/reason, local evidence digest,
provenance, upstream edge IDs와 upstream transition digest 목록, 이전 head를
보존한다. resolution의 cause도 root digest 하나가 아니라 경로의 transition
digest 목록이다.

UNKNOWN fixture의 관측 인과 edge는 `5/8`(claim별 허용 shortest-path edge union
`3/8`), REFUTED fixture는 `7/8`(union `5/8`), recovery는 실제 discharge에
사용된 `3/8`(union `2/8`)이다. 이 union은 결정론적 경로들의 합집합이지 전역
cardinality-minimum 증명이 아니다. local evidence만으로 discharge된 downstream은 direct local
evidence로 분류하고, `REQUIRES` upstream이 실제 사용된 경우에만 dependency
discharge로 분류한다. recovery edge를 destination이 DISCHARGED라는 이유만으로 세지
않는다. 최대 cause path는 node 수가 아니라 edge depth이며 이 fixture에서는
`2`다.

edge algebra의 고정 truth-table 분모는 edge kind마다 positive/negative 두 건,
총 `8/8`이다. SUPPORTS positive/negative는 모두 target을 discharge/refute하지
않고, REQUIRES positive만 upstream+local 충족 시 discharge한다. CONTRADICTS와
FAILURE_ENTAILMENT는 명시된 방향의 positive case만 refute한다.

## Append-only recovery

recovery는 이전 UNKNOWN receipt가 가리키는 raw source와 raw artifact/evidence를
다시 관측하고 graph, transitions, resolutions, metrics, decision 전체를 재생한
뒤 digest, transition head, 여섯 claim state, graph/proposition digest를 검증한다.
새 ledger를 만들지 않고 기존
12 transition을 byte-for-byte 보존한 뒤 6개 transition을 append한다.
각 recovery transition은 해당 claim의 local evidence digest와 필요한 upstream
transition digest chain을 결합한다. 따라서 회복 결과는 `12 preserved + 6
appended`, append-only chain `1`, 실제 causal recovery edge `3/8`이다.

## Independent consumer and interventions

별도 judge는 producer package와 expectedClaims/expectedEdges를 import하지
않고 raw source, raw evidence artifact, prior ledger에서 graph, truth table,
state, transitions, metrics, decision을 재구성한다. CI에서 source
reconstruction은 `1/1`, producer import은 `0/1`이다.

네 개의 semantic 개입은 원인을 분리한다.

* source-only: 같은 accepted raw evidence를 고정하고 `VALUE_PROGRAM`만 바꾼다.
  semantic digest, claim transitions, decision이 바뀐다.
* observation-only: source semantic digest를 고정하고 provider operation만
  바꾼다. evidence digest, states, transitions, decision이 바뀐다.
* edge-only: source value가 선언한 typed edge만 바꿔 refuting path를 제거한다.
* comment-only: raw source bytes만 바꾸고 semantic/graph/evidence digest,
  states, transitions, decision을 보존한다.

각 report는 baseline/intervention의 source, semantic, graph, evidence digest와
정확한 claim state/transition digest vector, cause path, snapshot effects,
authority resolution을 함께 기록한다. tracked+untracked pre/post snapshot이
같고 output path가 repository 밖이어야 `RepositoryWrites=0`이다. 무변경만으로
권한 부재를 증명할 수 없으므로 독립 capability/permission evidence가 없으면
authority는 `TRANSIENT_WRITE_AUTHORITY_UNKNOWN`이며 stage/step/reason을 남기고
fail closed한다. 현재 관측 범위는 `NET_REPOSITORY_STATE_UNCHANGED`이고, 별도
회귀 분모에서 `NET_REPOSITORY_STATE_CHANGED`와 transient unknown도 실행한다.
`contents:read`는 GitHub API capability일 뿐 로컬 write 불가 증명이 아니다.
semantic promotion authorization은 항상 false다.

증거 provenance 회귀는 `3/3`(caller flag only, observation artifact changed,
observation absent), authority 회귀는 `3/3`(net same, net changed, transient
unknown), prior tamper 거부는 `3/3`, path metric은 `2/2`(UNKNOWN/REFUTED), owner
applicability는 `3/3`(direct, dependency, discharged N/A)로 고정한다. DISCHARGED
resolution의 failure responsibility/owner는 `N/A`/empty이고, direct unknown은
자기 claim, dependency blocked/refuted는 실제 shortest path의 첫 claim을 owner로
기록한다.

## Principles and limits

provenance 경계는 [W3C PROV-O](https://www.w3.org/TR/prov-o/)의 entity,
activity, agent, influence 어휘를 참고했다. declared graph와 observed run
event를 분리하는 방식은 [OpenLineage object model](https://openlineage.io/docs/spec/object-model/),
[run cycle](https://openlineage.io/docs/spec/run-cycle/),
[producer identity](https://openlineage.io/docs/spec/producers/)를 참고했다.
방향성 있는 shortest path와 개입 보고의 한계는
[Stanford causal models](https://plato.stanford.edu/entries/causal-models/)의
구조적 관점과만 연결되며, 이 실험은 통계적 인과나 외부 세계의 truth를
주장하지 않는다.

반증 조건은 명확하다. judge가 source-derived proposition/target, edge kind,
raw evidence digest, prior head, transition provenance, state algebra, cause
path, effect snapshot 중 하나라도 producer receipt-only 정보로 수용하면
실패한다. SUPPORTS/REQUIRES가 upstream refutation을 dependent refutation으로
바꾸거나, UNKNOWN이 evidence 없이 discharge되거나, comment-only가 semantic
digest를 바꾸면 이 실험은 반증된다. 또한 graph는 이 고정된 acyclic six-node
fixture의 closed-world reconstruction일 뿐, 외부 데이터의 진실성·완전성·실행
정확성·네 edge type의 보편성을 보장하지 않는다.
