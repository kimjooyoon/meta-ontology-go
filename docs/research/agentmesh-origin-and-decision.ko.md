# AgentMesh에서 Gooo에 가져올 연결 원리

검토일: 2026-10-13. 대상은 Google 저장소의
[`9fd1a41a15174df0aaef601c942410c2116bff91`](https://github.com/google/agentmesh/tree/9fd1a41a15174df0aaef601c942410c2116bff91)이다.
문서와 Go 코드를 읽었으며 설치, 외부 노드 연결, 모델 실행은 하지 않았다.
이 문서의 연동 인터페이스는 후속 설계다.

## 무엇을 하는 프로젝트인가

AgentMesh는 에이전트 옆의 노드가 도구와 모델을 공개하고, 다른 제공자를 발견하고
호출하게 하는 네트워크다. MCP, A2A, OpenAI 호환 추론 연결을 제공한다.
README는 현재를 pre-1.0으로 설명한다.
[프로젝트 소개](https://github.com/google/agentmesh/tree/9fd1a41a15174df0aaef601c942410c2116bff91#readme)

Gooo에는 모델 학습법보다 **이름·제공자·연결 방식의 분리**가 먼저 유용하다.
한 컴퓨터에서 작은 결정 모델을 쓰다가 실행 위치를 옮겨도, 그 결정을 요청한
Gooo 선언과 후보의 정체성을 유지하는 데 적용할 수 있다. 이는 우리의 적용 판단이다.

## 가져올 원리

| 확인한 AgentMesh 구조 | Gooo에 적용할 방향 | 순서 |
| --- | --- | --- |
| 서비스 종류와 이름으로 발견하고 실제 노드는 peer ID로 구별 | 결정 기능 ID, 제공자 ID, 모델 파일 해시를 각각 유지 | 먼저 |
| backend adapter가 모델 목록을 제공하고 서비스가 연결을 담당 | Gooo의 후보 선택 요청을 로컬 함수·HTTP 어댑터가 같은 형태로 처리 | 먼저 |
| 발견 정보에 시각, 제공 키, 부하가 있음 | 모델이 지원하는 특징 스키마와 선택지 종류, 관찰 시각을 남김 | 다음 |
| 로컬 제공자를 우선하고 원격은 조건과 부하로 선택 | 로컬 결정 모델 우선, 같은 조건에서는 안정된 ID 순서, 연결 부재 시 소스의 기본 선택 | 다음 |
| A2A 카드의 주소·전송 방식을 mesh 주소에 맞춰 변환 | 원본 설명과 변환된 설명을 각각 보존하고 변환 활동으로 연결 | 기원 설계에 반영 |

이 표의 근거는 [구조 문서](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/site/content/docs/concepts/architecture.md),
[추론 연결부](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/internal/node/inference_engine.go),
[노드 API](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/site/content/docs/reference/node-api.md),
[서비스 공개 문서](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/site/content/docs/guides/exposing-services.md)다.
안정된 ID에 따른 동점 처리와 결정론적 기본 선택은 Gooo 쪽 제안이다.

## 기원을 두 층으로 읽기

AgentMesh의 `ServiceAnnounce`는 어느 peer가 어떤 서비스·모델 키를 알렸는지,
언제 알렸는지와 부하 힌트를 담는다. 수신 코드는 발신 peer와 선언된 peer를
대응시키고 오래된 정보를 발견 목록에서 제외한다.
[프로토콜](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/api/agentmesh.proto#L175-L197),
[발견 기록 구현](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/internal/node/discovery/view.go)

여기서 얻는 기원은 **제공자와 발표의 기원**이다. 검토한 발견·추론 인터페이스에서
모델의 학습 자료, 가중치 해시, Gooo 후보의 선언 위치까지 이어지는 필드는 확인하지 못했다.
Gooo에는 **판단 내용의 기원**을 이어 붙일 필요가 있다.

- **선언:** 원본 패키지·파일 해시, activity·의도·빈칸·후보 ID와 위치
- **제안:** 모델 메타데이터·가중치 해시, 특징 스키마, 입력 문맥, 후보 순서
- **연결:** 기능 ID, 제공자 ID, 실행 위치, 요청 ID와 관찰 시각
- **구성:** 선택한 후보, 적용한 변환, 만들어진 본문과 원본 조각의 연결

예를 들어 같은 `gooo-ranker`라는 서비스 이름을 두 PC에서 제공해도, 어느 모델
파일이 어느 입력으로 후보를 제안했는지 읽을 수 있어야 한다. 서비스 이름은 찾는
이름이고, 파일 해시는 사용한 모델을 가리키며, 선언 ID는 판단 대상의 의미를 가리킨다.

기원은 프로그램의 여행 기록처럼 이어진다. 검사 결과는 그 여행 중 관찰한 동작으로
연결한다. 원본이 사라지거나 제공자가 바뀌어도 이미 기록한 선언과 변환은 보존한다.

## 작은 결정 모델에 맞춘 연결안

현재 Gooo에는 이미 `decisionroute.Request`의 후보·기본 선택과 요청 해시,
`Receipt`의 TinyGo 메타데이터·가중치 해시가 있다. `RetainedModelInfo`는 특징 스키마도
기록하고, 로컬 제공자는 불변 모델과 호출별 작업 배열을 분리한다.
[기존 요청과 기록](../../internal/decisionroute/request.go),
[로컬 제공자](../../internal/decisionroute/tiny_go.go),
[모델 정보](../../internal/bodycodegen/body_paths_retained.go)를 재사용한다.

첫 실험은 이 계약에 실제 제공자 ID와 원본 선언 참조를 연결하는 Go 어댑터다.
기존 Laya용 `provider_model` 열거값은 외부 제공자 전체의 식별자로 쓰기에는 범위가
좁으므로, 일반 제공자 기능과 Laya 선택자를 분리하는 변경이 필요하다.
소스가 소유한 유한 후보 목록을 유지하고 선택 결과를 기존 기록에 연결한다.
아래는 제안 필드이며 새 공개 API는 아직 없다.

| 요청 | 응답 |
| --- | --- |
| 요청 ID, 원문 해시, activity·hole ID | 요청 ID, 선택한 후보 ID 또는 후보 순서 |
| 후보 ID 배열, 특징 스키마, 특징 값 | 모델 메타데이터·가중치 해시, 사용한 특징 스키마 |
| 소스가 정한 기본 선택, 처리 기한 | 실제 제공자, 관찰 시각, 기본 선택으로 전환한 이유 |

MCP 연결이 필요해지면 이 선택 작업을 하나의 도구로 공개할 수 있다.
AgentMesh의 추론 경로는 OpenAI 호환 모델 목록과 completion 요청을 사용한다.
우리 모델의 입출력은 후보 배열과 점수이므로, 도구 호출이나 작은 HTTP 요청이
먼저 검토할 형태다. `gooo.decision.select` 같은 도구명은 설계 예시다.

취소·시간 제한은 Go context를 전달하고, 상태 조회 실패와 실제 판단 실패를 구별한다.
AgentMesh의 모델 조회에는 짧은 캐시와 조회 기한이 있다.
[발견용 조회](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/internal/node/discovery_source.go),
[모델 목록 캐시](https://github.com/google/agentmesh/blob/9fd1a41a15174df0aaef601c942410c2116bff91/internal/node/inference_service.go)
단일 PC의 로컬 모델 호출은 직접 처리하고, 여러 실행 위치가 필요해질 때 mesh 어댑터를 추가한다.
P2P·제어 평면·대형 모델 서버를 컴파일러 필수 구성에 포함하는 작업은 현재 범위 밖이다.

## 현재 구현과 후속 작업

이번 [원본 파일 연결](../../examples/workspace-fill-origins/README.md)은 후보·빈칸·의도와
조건 조각을 최초 파일까지 잇는다. Gooo 프로그램이 그 출처를 typed record로 읽는
경로도 있다. 이 작업은 AgentMesh 조사 전 시작했으며, 원본과 변환 결과를 각각
보존한다는 원리를 비교하는 기반이 된다.

다음 작은 범위는 결정 요청 한 건에 모델 파일과 제공자 정보를 함께 연결하는 것이다.
그 뒤 같은 후보 계약을 로컬 호출과 어댑터 호출에 전달한다. 재학습은 이 연결을
만들기 위한 선행 조건이 아니다.

관찰 지표는 출처를 찾을 수 있는 판단 수/전체 판단 수, 변환 뒤 유지된 연결 수,
끊긴 연결의 이유를 먼저 둔다. 이어 실제 호출 시간·메모리를 같은 입력 범위에서
기록한다. AgentMesh나 기존 구현보다 빠르다는 측정은 이번 조사에서 수행하지 않았다.

AgentMesh는 Apache-2.0이다. 이번 변경은 구조를 분석하고 Gooo의 설계 방향을
정리한 것이며, 외부 구현 코드를 복사하거나 새 의존성으로 추가하지 않았다.
