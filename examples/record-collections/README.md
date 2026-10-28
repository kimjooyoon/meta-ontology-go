# 레코드 목록으로 합계 계산하기

`main.gooo`는 상품 항목마다 수량과 가격을 함께 받습니다. `Basket.items`의
타입은 `Line`이고, `Total`은 항목 순서대로 두 값을 곱해 더합니다.
항목을 여러 숫자 배열로 나눠 전달할 필요가 줄어듭니다.

저장소 루트에서 실행합니다. 이 예제는 개발 중인 V7 레코드 참조 경로를 사용합니다.

```sh
go build -o /tmp/gooo-record-demo ./cmd/gooo
/tmp/gooo-record-demo check --semantic --json examples/record-collections/main.gooo
/tmp/gooo-record-demo package interface --json examples/record-collections/gooo.workspace.json
/tmp/gooo-record-demo body-codegen --json --activity Total examples/record-collections/main.gooo
/tmp/gooo-record-demo body-compose --source examples/record-collections/main.gooo \
  --cases examples/record-collections/cases.json --out /tmp/gooo-record-first
/tmp/gooo-record-demo body-compose --source examples/record-collections/main.gooo \
  --cases examples/record-collections/cases.json \
  --composition /tmp/gooo-record-first/composition.json --out /tmp/gooo-record-replay
```

출력 폴더는 새 경로여야 합니다. 기본 실행은 설치된 Go를 확인하고, `--go-bin`으로
선택한 Go 1.27.3 실행 파일을 지정할 수 있습니다. 모델 설정 없이 결정론적으로
코드를 구성하고 저장된 구성도 같은 소스·타입·코드 생성 결과와 대조합니다.

`cases.json`의 기대값은 빈 목록 `0`, 단일 항목 `14`, 같은 항목 두 개 `28`,
큰 정수 `9007199254740993`입니다. 실행 결과의 `finite_passed/finite_total`은
이 네 기대값에 대한 관측입니다. 정수는 Go의 `int64` 범위와 연산을 사용합니다.

## Go 구조와 구현 슬롯 만들기

일반 Go 프로젝트에서 쓸 구조체와 구현 슬롯의 초안은 다음 명령으로 만듭니다.

```sh
/tmp/gooo-record-demo generate --json examples/record-collections/main.gooo \
  --out /tmp/gooo-record-structures
```

V7은 선언의 ID를 따라 `Line`, `*Line`, `[]Line`, `*[]Line` 형태를 구성합니다.
필드 순서와 원래 필드·타입 ID는 생성된 소스 맵에 남습니다. Go 이름이 바뀌면
같은 ID를 참조하는 필드의 Go 타입도 함께 갱신됩니다. `generate`의 활동 구현은
기존 구현 슬롯의 초안을 사용합니다. 작성된 `computes` 계산의 실행과 기대값 확인은
위의 `body-codegen`, `body-compose` 경로를 사용합니다.

## 출처와 입력 한계

- 타입 참조는 선언의 안정된 ID로 연결됩니다. 원래 타입 이름이나 따옴표 URI는 소스에 남습니다.
- `input_fields`와 `actual_fields`의 `children`은 목록의 `index`, 레코드 ID, 내부 필드 ID를 보존합니다.
  같은 값의 항목도 각 위치를 유지합니다.
- 목록 읽기 관측은 원본 입력의 backing element와 실제 인덱스를 연결합니다.
  중첩 필드의 전체 동적 데이터 의존성은 별도 관측 범위로 남아 있습니다.
- 선택한 레코드 타입은 기존 16개, 각 레코드의 필드는 32개, 목록의 요소는 256개 이내입니다.
  참조 그래프는 순환을 거부하고 깊이를 8단계로 제한합니다.
- 새 중첩 값은 레코드·필드·배열 요소를 합쳐 8,192개 JSON 값 노드 이내입니다.
  입력 JSON 크기, 출력 크기, 텍스트 바이트와 반복문 작업량에는 기존의 별도 한계가 적용됩니다.
- 빠진 필수 필드, 알 수 없는 키, 중복 키, 잘못된 내부 타입, `null`, 초과 목록은
  실제 입력 오류 코드와 함께 코드 생성 전에 거부됩니다. 선택 필드는 생략할 수 있으며 빈 목록은 존재하는 값입니다.

## 후보 평가에서 항목 레코드 다루기

`assembly.gooo`는 같은 합계 문제에 `assembling`을 붙입니다. 소스에 합계 그대로와
합계에 1을 더하는 두 선택지를 선언하고, 빈 목록·단일 항목·반복 항목·큰 정수의
네 기대값으로 선택을 확인합니다.

```sh
/tmp/gooo-record-demo body-codegen --json --activity Total \
  examples/record-collections/assembly.gooo
```

`record_assembly`의 `passed/total`은 이 소스의 네 사례를 분모로 합니다.
선택된 본문과 `picked` 기록을 Gooo 소스에 남기며, 저장한 선택은 추가 모델 호출
없이 재구성할 수 있습니다. 기존 모델의 표현 계약에 중첩 레코드가 없으면
`NOMINAL_RECORD_LAYOUT_FEATURE_UNAVAILABLE`로 기록하고 결정론적으로 선택합니다.

후보 평가기는 필수·선택 레코드, 순서가 있는 레코드 목록, 존재하는 빈 선택 목록,
목록 순회와 인덱스 읽기, 레코드 복사와 목록 리터럴을 다룹니다. 선택 레코드는
`nil`과 비교해 존재를 확인한 뒤 내부 필드를 읽을 수 있습니다. 자식을 지역 값으로
복사해 수정하고 부모 필드에 다시 넣는 방식은 원래 값과 형제 값을 보존합니다.
필수 자식의 지역 복사는 값을 갖습니다. 선택 자식은 존재를 나타내는 불변 참조를
사용하므로, 다른 지역 이름으로 받아도 원래 저장 위치와 연결됩니다. 이 참조를 통한
필드 수정은 코드 생성과 평가에서 거부됩니다. 선택 자식을 바꾸려면 새 자식 값을
만들고 부모 필드 전체를 `&child`로 교체합니다. 실제 Go 회귀 검사도 호출 전후의
입력 JSON을 비교해 원래 입력이 보존되는지 확인합니다.

중첩 기본값은 타입별 불변 구조를 공유하고, 값 한계는 실제로 펼쳐질 각 위치를
계산합니다. 반복된 같은 값도 각 목록 위치의 비용을 가집니다.

현재 검증은 위 네 사례와 회귀 검사의 작성된 프로그램별 기대값에 한정됩니다.
중첩 필드의 직접 대입, 두 개의 존재하는 선택 포인터 간 동일성 비교, 전체 동적
데이터 의존성과 중첩 구조를 이해하는 모델 특징은 이어서 다룰 범위입니다.
[전체 언어 과제 #1834](https://github.com/kimjooyoon/meta-ontology-go/issues/1834)에서
값 생성과 출처 관측의 남은 경계를 추적합니다.

반복문이 항목의 필드를 바인딩한 원본 위치는
[반복 필드 기원](../../docs/language/record-iteration-origins.ko.md)에서 설명합니다.
목록 표현식과 실제 변수 토큰의 범위, 파생 필드 ID를 구분합니다.
반복 항목의 각 필드는 실제 원소 읽기 번호를 통해 원래 입력 목록·위치에 연결됩니다.
같은 값이 반복돼도 위치를 유지하며, 기록 한도를 넘으면 `READ_LIMIT`로 남깁니다.
