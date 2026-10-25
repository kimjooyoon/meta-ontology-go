# Meta Mesh Ontology를 Gooo에 적용할 지점

검토 대상은 Blitz 등, *Meta Mesh Ontology for Cross-Domain Mental Health Knowledge
Integration: Development and Evaluation Study*, JMIR Mental Health 2026;13:e96311,
[doi:10.2196/96311](https://mental.jmir.org/2026/1/e96311)이다.
첨부된 출판본과 공개 저장소의 고정 소스를 함께 확인했다.

Gooo에 유용한 부분은 원래 개념을 보존하면서 서로 다른 체계를 연결하는 방법이다.
여러 지도를 포개 보되, 지도의 원점과 축을 지우지 않는 방식에 가깝다. 이를 출처,
버전, 문맥을 가진 연결로 표현하고, 연결이 허용하는 판단의 범위를 구분한다.

## 논문에서 실제로 평가한 것

MMO core와 CDF가 공통 구조를 제공하고, 별도로 버전 관리되는 열 개 도메인 모듈이
외부 체계의 선택된 개념을 연결한다. 모든 외부 온톨로지를 통째로 가져오는 방식이
아니며, 원래 식별자와 매핑의 출처를 보존한다. CDF의 여덟 상위 도메인은 그 연구의
앵커 설계다. Gooo의 도메인 구성을 그 여덟 범주로 고정할 이유는 없다.

평가에는 합성 시나리오의 50개 변수, 사전에 정한 11개 competency question과
네 개의 음성 대조군이 사용됐다. OWL 2 DL 적합성, 논리적 일관성, 지정된 조회와
추론의 결과를 확인했다. 54/54 라벨 보유율은 해당 모듈 자원의 문서화 지표다.
매핑의 독립적 타당성, 실사용자의 데이터 상호운용, 운영 성능과 학습 성능은 별도
평가 대상으로 남아 있다. 이를 Gooo의 정확도나 저사양 성능 근거로 옮기지 않는다.

가중치는 문맥별 부분 함수로 정의돼 있다. 정의되지 않은 가중치가 허용되며,
도메인 소속 정도·관측 빈도·작업 중요도 같은 서로 다른 구성을 같은 실수 축만으로
합치지 않도록 한다. 공개 v1.0.1의 구체적 학습 가중치나 집계 규칙을 제시한 것은
아니다. 유한 그래프들을 계속 확장할 수 있다는 수학적 설명도 실행 시간과 메모리의
상한을 보장하지 않는다.

## Gooo에서 가져올 설계

| 설계 | Gooo에 적용할 방식 | 현재 범위 |
| --- | --- | --- |
| 원래 식별자 보존 | 외부 개념과 Gooo 개념을 별도 ID로 두고 매핑을 기록 | 실행 가능한 예제 구현 |
| 관계 강도 구분 | exact/close/related/association/refinement를 구분 | 입력 주장 분류, 추론은 미구현 |
| 매핑 자체의 출처 | 출처 참조, 양쪽 버전, 적용 문맥을 함께 보존 | 예제 입력·런타임 기록에 보존 |
| core와 모듈 분리 | 공통 출처·의미 ID 계약을 유지하며 도메인별 선언을 추가 | 기존 package/manifest 경로 재사용 |
| 사전 질문과 반례 | 조회 가능성과 잘못된 강한 결론을 별도 검사 | 20입력·40개 예상 활동 출력 |
| 문맥별 점수 | 값·단위·종류·문맥·분모·출처와 미정 상태를 구분 | 후속 설계, 점수 집계 미구현 |

중요한 경계는 ‘연결했다’와 ‘같다고 판정했다’ 사이에 있다. W3C SKOS는
`exactMatch`를 추이적 관계로 정하고, `closeMatch`와 `relatedMatch`는 추이적으로
정하지 않는다. 또한 이 매핑 관계들은 OWL의 개체 동일성·클래스 동등성과 다른
의미를 가진다. [SKOS Reference §10](https://www.w3.org/TR/skos-reference/#mapping).

현재 Gooo 예제는 관계 IRI를 입력 데이터로 읽고 직접 주어진 주장만 분류한다.
RDF/OWL 추론기를 구현한 것이 아니다. 앞으로 문맥이 붙은 매핑을 RDF로 내보낼
때에는 매핑 주장에 별도 식별자를 주고 문맥·출처를 함께 표현해야 한다. 문맥을
생략한 전역 동등성 triple로 바꾸면 원래 계약보다 강한 결론을 만들 수 있다.

고정 [MMO v1.0.1 소스](https://github.com/MetaMeshOntology/meta-mesh-ontology/blob/83d768469c8b5b0a6ca81e7e9aa86f84f2f3af89/ontology/mmo.ttl)에는
매핑 속성과 별도로 명시적인 `rdfs:subClassOf`, `owl:sameAs`도 존재한다.
따라서 모든 연결을 약한 관계 하나로 취급해서도 안 된다. 여기서는 URI와 속성 선언을
읽었으며, 외부 온톨로지 전체의 reasoner 검증을 다시 수행하지 않았다.

## 이번에 만든 작은 적용

[contextual-mappings](../../examples/contextual-mappings/README.md)는 기존 Gooo
record·조건식·지역 변수·binding만 사용한다. 추가 모델이나 문법이 필요하지 않다.

입력은 source ID, target ID, 관계, 매핑 문맥, 출처 참조, 양쪽 버전과 현재 조회의
문맥·버전이다. 출처가 없거나 버전이 바뀌거나 다른 문맥이면 이를 별도 상태로
보인다. 현재 값은 호출자가 공급하는 문자열이므로, 실제 어댑터는 원본 자료에서
관찰한 값과 연결해야 한다. 예제는 URI를 조회하거나 매핑의 현실적 타당성을 판단하지 않는다.

결과는 다음 판단을 분리한다.

- 참조할 매핑 주장이 현재 조회 문맥에 있는가.
- 그 매핑의 의미 타당성이 평가됐는가: 항상 `UNASSESSED`.
- 이 조회가 ID 병합·행동 검증·권한 부여를 승인하는가: 모두 `false`.

`UNSUPPORTED_RELATION`은 이 예제의 지원 범위 밖이라는 뜻이다. OWL 동일성이라는
표준 자체의 오류나 소스 모순을 뜻하지 않는다. 동일 문자열 ID나 exact-match 주장도
이 조회에 새로운 실행 권한을 부여하지 않는다. 이 프로그램은 명시적 조회로 실행한다.
컴파일러의 필수 경로에 온톨로지 로딩이나 reasoner를 추가하지 않는다.

평가 파일은 여섯 개의 참조 가능한 주장 사례와 열네 개의 누락·문맥·버전·관계 범위
사례를 포함한다. 예상값은 프로그램과 별도 JSON에 선언했다. 입력 하나당 두 활동의
전체 출력이 검사되므로 분모는 40개 출력이며, 세계의 모든 매핑이나 개별 필드 수를
분모로 삼지 않는다. 저장한 실행의 재생도 같은 기대값으로 별도 확인한다.

## 앞으로 사용할 지표와 순서

1. **출처 충족률:** 필요한 식별자·버전·문맥·출처가 있는 주장 수 / 검사한 주장 수.
   이는 구조적 추적 가능성이며 의미 정답률은 별도다.
2. **문맥 적합 조회율:** 요청한 문맥·버전에서 참조 가능한 주장 수 / 요청한 주장 수.
   누락과 오래된 연결은 각각 남긴다.
3. **금지된 결론의 오수락:** 매핑만으로 ID 병합·행동 동일성·권한을 승인한 출력 수 /
   해당 반례의 출력 수. 테스트 범위에서의 값으로 보고한다.
4. **독립 평가와 비용:** 사전에 정한 질문·반례의 결과, 실제 사용자 적용, 조회 시간,
   메모리를 각각 보존한다. 라벨 보유율이나 그래프 크기로 이 값을 대신하지 않는다.

다음 연결은 도구 정의나 Gooo 소스의 실제 버전·출처를 예제의 입력 계약에 공급하는
어댑터다. 기존 PROV-O entity/activity/used/wasDerivedFrom 경로에 매핑 생성과 원본
자료를 연결하고, 실제 행동과 현재 권한은 각 소유자가 검증한다. 모듈마다 새로운
커널 속성이나 독립 실행 체계를 먼저 늘리지 않고 기존 package와 조회 경로에 맞춘다.

소형 decision 모델을 연결한다면, 허용된 후보 중 어느 연결을 먼저 조회할지 또는
어떤 추가 문맥을 요청할지를 고르게 한다. 출처가 없다는 사실을 모델 점수로 덮거나,
추상 가중치를 검증된 의미 호환성으로 읽지 않는다. 논문은 이 모델의 학습 자료나
성능을 제공하지 않으므로 별도 데이터와 평가가 필요하다.

## 참고와 출처

- Blitz R, Richter M, Kerber A, Leenings R, Thun S, Opel N. *Meta Mesh Ontology for
  Cross-Domain Mental Health Knowledge Integration: Development and Evaluation Study*.
  JMIR Mental Health. 2026;13:e96311. [논문](https://mental.jmir.org/2026/1/e96311),
  [DOI](https://doi.org/10.2196/96311). 2026-10-23 발행.
- [MMO 고정 구현 v1.0.1](https://github.com/MetaMeshOntology/meta-mesh-ontology/tree/83d768469c8b5b0a6ca81e7e9aa86f84f2f3af89),
  [Modules Registry v1.0.0](https://github.com/MetaMeshOntology/modules-registry/tree/v1.0.0).
- [W3C SKOS Reference](https://www.w3.org/TR/skos-reference/), 특히 §10.6.3, §10.6.8.

원 논문 저작권은 © 2026 Rogério Blitz, Maike Richter, André Kerber, Ramona Leenings,
Sylvia Thun, Nils Opel이며 JMIR Mental Health에 처음 게재됐다.
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)으로 공개됐다.
이 문서는 논문의 설계 아이디어를 Gooo에 맞춰 검토·변형한 설명이다.
새 예제와 코드는 별도로 작성했으며 외부 임상 데이터나 온톨로지 본문을 복제하지 않았다.
