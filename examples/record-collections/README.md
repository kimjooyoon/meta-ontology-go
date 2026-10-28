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

현재 실행 범위는 작성된 `computes` 본문과 기존 `body-codegen`, `body-compose`,
패키지 선언 조회 경로입니다. 일반 구조 생성기는 V7 레코드 타입과 구현 슬롯 투영을
지원합니다. 레코드를 받는 `assembling` 후보 평가는 별도 지원 범위를 유지합니다.
[전체 언어 과제 #1834](https://github.com/kimjooyoon/meta-ontology-go/issues/1834)에서
값 생성과 출처 관측의 남은 경계를 추적합니다.
