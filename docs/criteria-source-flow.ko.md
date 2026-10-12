# 후보와 요구 조건을 함께 읽기

Gooo가 `11 - input`과 `input - 11` 중 하나를 고를 때, 후보의 모양만 읽으면
어느 방향이 요구에 맞는지 알기 어렵습니다. 이 관계를 `.gooo`에 쓰면 조건을
검사하는 데 사용할 수 있습니다. 작은 결정 모델을 만들 때도 같은 조건의 구조가
필요합니다.

[전체 예제](../examples/criteria-source-flow/main.gooo.fixture)는 `Fee`의 선택지와
다음 조건을 한 소스에 둡니다.

```gooo
activity Difference(Integer,Integer) -> Boolean id "criteria-view://difference" computes "let expected = Delta(input0); return input1 == expected"
activity Delta(Integer) -> Integer id "criteria-view://delta" computes "return input - 11"
```

`Difference`의 첫 인자는 원래 입력, 마지막 인자는 후보가 돌려주는 결과입니다.
여기서 `Delta`는 요구 조건을 표현하는 고정 함수입니다. 선택할 수 있는 구현 범위는
`Fee`의 `assembling` 선언에 있습니다. 설계도에 부품 후보와 연결 조건을 함께
적는 것에 가깝습니다.

## 소스에서 내보내는 내용

이 변경을 포함한 개발 체크아웃에서 다음 명령을 실행합니다.

```sh
go run ./cmd/gooo body-context --value-flow --activity Fee \
  examples/criteria-source-flow/main.gooo.fixture
```

기존 `body-context --value-flow` 결과에 선택적으로 `criteria_flow`가 붙습니다.
스키마는 `gooo/criteria-source-flow/v1`입니다.

| 항목 | 담는 정보 |
| --- | --- |
| `candidate_plan` | Integer 경로의 타입이 붙은 본문, 선택지, 자연어 의도 |
| `candidate_plan_sha256` | 후보 계획의 식별값 |
| `predicates[].inputs` | 인자 번호, 타입, `original_input` / `candidate_result` 역할 |
| `predicates[].value_flow` | 조건식의 정의·읽기·연산·분기·반환과 원문 위치 |
| `predicates[].nodes` | 인자 타입, 정규화한 상수, 연산자, 순서가 있는 연결 |
| `predicates[].graph_sha256` | 정규화한 노드 배열의 식별값 |
| `activity_id`, `generated_sha256` | 조건 함수의 의미 ID와 생성된 고정 프로그램의 식별값 |

레코드 결과의 조건도 지원합니다. 레코드 구현의 후보 구조는 기존 `value_flow`와
`choices`에, 조건 함수의 구조는 `criteria_flow`에 담깁니다. 레코드에는 Integer용
`candidate_plan`이 추가되지 않습니다.

고정 보조 함수는 인자 연결과 반환 경로를 따라 펼칩니다. 원문의 활동 ID,
호출 대상, 보조 함수 목록과 위치는 `value_flow`에 남습니다. 구조 노드 배열에서는
지역 변수의 표시 이름과 원문 위치를 제외합니다. 변수 이름만 고치면 구조가
유지되고, 빼기의 방향이나 분기 조건을 고치면 구조가 바뀌는지를 회귀 검사합니다.
이 식별값은 정규화한 표현을 비교하는 값이며, 수학적 동치 판정의 범위는 다루지 않습니다.

## 평가 입력과 구분하기

조건의 본문에는 `11` 같은 프로그램 상수가 들어갈 수 있습니다. 유한 검사에 쓰는
`input`, `holdout_input`, 관찰 질문과 그 답은 이 구조에서 제외합니다.
`--include-plan`을 명시하면 별도 `expanded_plan`에 유한 입력 선언도 담깁니다.
모델 입력을 구성하는 소비자는 이 두 영역의 목적을 구분해야 합니다.

명령은 소스를 해석하고 타입과 표현 범위를 확인합니다. 조건을 실제 입력에 대해
평가하거나 후보를 실행하지 않습니다. 반환되는 `model_predictions`와
`candidate_tests`는 모두 0입니다. [저장한 출력](../examples/criteria-source-flow/context.json)과
[출력 생성 기록](../examples/criteria-source-flow/observation.json)에 개발 바이너리,
소스 및 출력의 식별값을 남겼습니다.

조건마다 기존 분석기의 512개 노드, 64개 동시 바인딩, 중첩 조건 16단계 한도를
적용합니다. 보조 함수 확장도 이 노드·바인딩 예산을 공유합니다. 지원 범위 밖이면
해당 조건과 전체 `criteria_flow`가 `UNRESOLVED`가 되고 이유를 남깁니다.
그 조건에는 사용 가능한 것처럼 보이는 부분 구조를 내보내지 않습니다.
다른 조건의 완성된 관찰은 각각의 상태와 함께 유지됩니다.

## 자체 소형 모델에 연결할 다음 단계

현재 구현은 모델에 전달할 소스 표현을 준비합니다. 기존 모델 특징 계약과 가중치는
이 새 표현을 소비하도록 바뀌지 않았습니다. Integer criteria 선택은 계속
`SCALAR_CRITERIA_FEATURE_CONTRACT_UNAVAILABLE`을 기록하고 결정론적으로 진행합니다.
이번 변경으로 모델 정확도나 추론 속도가 개선됐다고 측정한 결과는 없습니다.

다음 모델 계약에서는 후보와 조건의 그래프를 서로 다른 입력 채널에 두고,
Gooo가 허용한 선택지별 점수만 출력하는 구성을 다룹니다. 한국어·영어 의도는
해당 선택지에 붙은 별도 채널로 유지할 수 있습니다. 코드 토큰 생성이나 새로운
조건 작성을 출력으로 요구할 필요가 없습니다.

유한 입력의 답을 보지 않은 상태에서 경로를 고른 뒤, Gooo의 실제 조건 검사와
별도 입력 평가로 결과를 측정해야 합니다. 표현 가능 비율, 선택한 후보가 조건을
만족한 비율, 추가 시도가 필요한 횟수는 각자의 분모로 기록할 예정입니다.
새 모델의 학습·연결·평가는 후속 작업입니다.

관련: [Integer 조건](source-scalar-criteria.md),
[레코드 값의 흐름](record-value-flow.md),
[조건을 Gooo에 쓰기](source-behavior-criteria.md).
