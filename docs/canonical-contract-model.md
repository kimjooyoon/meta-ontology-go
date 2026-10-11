# 변수 대입과 직접 반환을 같은 계산으로 읽기

개발 소스에서 [SDK 0.2.40의 연구 모델](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.40-experimental)을
일반 분기 조립에 연결합니다. 공개 컴파일러 0.6.27은 이 모델 형식을 지원하지 않습니다.
아래 명령은 이 기능이 포함된 소스 체크아웃과 Go 1.27.2를 기준으로 합니다.

`return input - limit`과 `let result = input - limit; return result`처럼 같은 계산을
다르게 적은 경우, 지원하는 순수 분기에서 모델이 같은 입력을 받게 합니다. 비교식과 각
분기에서 반환하는 식을 소스로부터 읽습니다. 실제 조립·컴파일·실행은 원래 소스를 사용합니다.

## 모델 입력을 먼저 보기

배포물의 `MODEL-SHA256SUMS`로 파일을 확인하고 `model-canonical-research-v1.json`을 사용합니다.
파일명은 아래 명령에서 현재 위치에 맞게 바꿀 수 있습니다.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-context \
  --model model-canonical-research-v1.json --activity Choose \
  examples/body-codegen/source-interaction-assignment-cases.gooo.fixture
```

`canonical_source_features`는 실제 모델이 읽는 528개 FP32 값입니다.
`canonical_branch_context`는 그 값을 만든 소스 비교식·반환식·구조를 보여줍니다.
원래 소스와 계획의 식별값도 남습니다. 조회는 추론과 후보 실행을 하지 않습니다.
출력 사례와 중간 Boolean 조건은 각각 별도의 입력 배열로 전달합니다.

## 조립하고 저장하기

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-construct \
  --source examples/body-codegen/source-interaction-assignment-cases.gooo.fixture \
  --entry Main --model model-canonical-research-v1.json \
  --construction-cases examples/body-codegen/source-condition-construction-cases.json \
  --cases examples/body-codegen/source-condition-evaluation-cases.json \
  --attempts 8 --out canonical-first --format text

GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-construct \
  --source canonical-first/original.gooo --construction canonical-first/construction.json \
  --cases examples/body-codegen/source-condition-evaluation-cases.json --format text
```

출력 폴더는 새 이름을 사용합니다. 모델은 각 지역 조립 세션의 첫 후보를 실행하기 전에
한 번 호출됩니다. 이후 후보를 계속 확인할 때는 다시 호출하지 않습니다. 저장 재실행은
모델 파일 없이 기록을 재검사하고 현재 평가 입력을 실행합니다. 처음 조립할 때 `--model`을
생략하면 고정 순서로 진행합니다.

## 지원 범위와 성적

현재 모델 입력은 한 순수 분기의 비교 순서·분기 배치 선택을 다룹니다. 변수의 마지막 대입을
따라갈 수 있는 평평한 정수·Boolean 식이 대상입니다. 지원 범위를 벗어나면
`DECLINED_TO_DETERMINISTIC`과 이유를 남기고 모델 호출 없이 일반 탐색으로 이어갑니다.
모델 파일 자체의 형식 오류는 명령 오류로 반환합니다.

기존 모델은 원래 입력 형식을 계속 사용합니다. 새 모델은 19,034개 FP32 가중치를 가지며,
[원래 연구](https://github.com/kimjooyoon/gooo-decision-runtime/tree/v0.2.40-experimental/studies/canonical-contract-learning-20261011)의
관련 변형에서 첫 선택 성공은 36/128, 비교 모델은 33/128, 고정 순서는 32/128이었습니다.
학습 자료에서도 성적이 낮아 연구 모델로 제공합니다. 이번 연결은 새 학습이나 품질 향상 측정이 아닙니다.
조립에 쓴 사례와 이후 평가의 범위는 [결과 보고서](construction-results.md)에서 나눠 읽을 수 있습니다.
