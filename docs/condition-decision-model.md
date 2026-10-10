# 자체 조건 모델을 Gooo 조립에 쓰기

개발 소스는 [SDK0.2.32](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.32-experimental)의
조건·출력 피드백 모델을 `body-codegen`, `body-context`, `body-construct`에 연결합니다. 공개0.6.25는
SDK0.2.28을 사용하므로 아래 명령은 이 변경이 포함된 개발 소스로 실행합니다.

모델은 소스에 선언된 선택지에 점수를 줍니다. Gooo가 그 조합으로 본문을 만들고 출력과
중간 조건 사례를 확인합니다. 조건이 틀리면 그 관측을 모델에 전달해 남은 경로 순서를
바꿀 수 있습니다. 모델 파일을 생략하면 선언된 기본 경로에서 결정론적으로 탐색합니다.

| 모델 입력 | 배열 | 가중치 값 | 읽는 정보 |
| --- | ---: | ---: | --- |
| 조건 v1 | 256칸 | 24,872바이트 | 소스·원래 의도·관측한 조건 실패 |
| 분기 역할 v2 | 256칸 | 24,872바이트 | v1과 각 분기가 반환하는 값의 종류 |
| 출력 피드백 v3 | 320칸 | 31,016바이트 | v2와 실제로 틀린 입력·기대값·출력값 |

모델 파일에 적힌 버전으로 배열을 구성합니다. 각 버전은 그 입력 형식으로 학습한
가중치를 사용하며, 구버전 모델의 입력을 자동으로 바꾸지 않습니다. 표의 크기는
가중치 값만 센 것으로 실행 프로세스 전체 메모리는 아닙니다.

## 출력 실패를 읽는 최신 모델

[128개 Gooo 소스 변형 실험](https://github.com/kimjooyoon/gooo-decision-runtime/tree/5a2b2d55fa5e05e1cf4f06e855f4e27ec25a3da4/studies/execution-feedback-learning-20261010)의
`result/model-v3_output_on.json`을 사용합니다. 파일은90,501바이트,
SHA256은 `7827d36a64ce72ee9d880f7e6bdb9b29f3cfba34d02d435781df4bf034dd789c`입니다.

```sh
go run ./cmd/gooo body-context --activity Choose \
  --model /path/to/model-v3_output_on.json \
  examples/body-codegen/source-output-feedback.gooo.fixture

go run ./cmd/gooo body-codegen --json --activity Choose \
  --path-model /path/to/model-v3_output_on.json \
  --path-step-attempts 1 --path-feedback-rounds 3 \
  examples/body-codegen/source-output-feedback.gooo.fixture
```

`body-context`의 `inputs[].execution_features`는 실제 판단에 쓰는320개 FP32 값입니다.
`input_sha256`은1,280바이트 배열의 해시이며 처음에는 출력 관측 구간이 비어 있습니다.
본문 한 개를 조립하고 검사한 직후, 다음 판단에는 그 본문에서 처음 틀린 출력이 들어갑니다.
실패값은 정확한 정수 바이트로 전달합니다. 어떤 선택 하나가 실패 원인이라고 단정하지 않고
본문 전체의 선택 조합도 함께 기록합니다.

`condition_feedback_judgments[].output_failure`에서 그 관측을 확인할 수 있습니다.
조건이 모두 맞아도 출력이 다르면 피드백을 받을 수 있습니다. 후보는 한 번씩 검사하며,
정해진 시도 횟수·피드백 횟수·시간 안에서 진행합니다. 점수는 남은 후보를 고르는 데 쓰고,
Gooo에 선언된 사례와 형식 검사가 실제 채택을 결정합니다.

공개 SDK 실험에서 출력 피드백은128개 변형 중4개에서 각각 두 번의 시도를 줄였습니다.
첫 선택과 변수 대입 과제는 개선되지 않았고, 전체 탐색 시간은16.138→20.855ms로 늘었습니다.
관련된 소스 변형들을 비교한 결과이며 독립적인 문제의 정답률로 일반화하기는 어렵습니다.
이 수치는 SDK 실험의 기록입니다. 아래 기존 v1 CLI 측정과는 생산자와 모델이 다릅니다.

[새 v3의 실제 CLI 사용 기록](research/execution-model-cli-20261010/README.md)에서는
기본3회에서 모델1회로 줄고, 선언 출력6개·조건3개와 별도 네이티브8개 사례를 통과했습니다.
첫 선택이 맞아 추가 피드백 호출은0회였습니다. 모델 읽기를 포함한 시간은
기본0.706ms, 모델2.821ms여서 이 작은 예제에서는 기본 탐색의 전체 시간이 더 짧았습니다.

## 기존 가중치 받기

[SDK0.2.29 배포](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.29-experimental)의
`gooo-condition-candidate-fp32.json`과 `SHA256SUMS`를 받습니다. 모델 파일은72,482바이트이고,
SHA256은 `a16696ed44c668f38cc2e7ee1dc6ff3f4716649d57e59df7c9490d1c670edc48`입니다.
SDK0.2.30은 같은 가중치를 사용합니다. 이번 연결을 위해 다시 학습할 필요가 없습니다.

## 판단할 입력부터 살펴보기

```sh
go run ./cmd/gooo body-context \
  --activity Choose --model /path/to/gooo-condition-candidate-fp32.json \
  examples/body-codegen/source-condition-model.gooo.fixture
```

`model_compatibility.status`가 `READY_FOR_RANKING`이면 현재 소스를 모델 입력으로 표현할 수
있다는 뜻입니다. 정답률을 의미하지는 않습니다. `inputs`에는 원래 한영 의도와256개 특징
값이 함께 나옵니다. `input_sha256`은 실제 판단에 쓰는1,024바이트 FP32 배열의 해시입니다.
소스·의도·조건 관측이 독립 구간을 사용하며 첫 판단의 조건 관측은 비어 있습니다.
이 명령의 모델 판단과 후보 사례 실행은0회입니다.

## 조립하고 조건 실패를 전달하기

```sh
go run ./cmd/gooo body-codegen --json --activity Choose \
  --path-model /path/to/gooo-condition-candidate-fp32.json \
  --path-step-attempts 1 --path-feedback-rounds 3 \
  examples/body-codegen/source-condition-model.gooo.fixture
```

소스에는 절댓값을 내는 사례와 **입력이 양수인지 확인하는 중간 조건**이 함께 있습니다.
음수 비교와 양수 비교는 분기까지 뒤집으면 같은 최종 답을 만들 수 있어서 두 지표를
따로 확인합니다. `condition_session_progress`는 시도한 후보를,
`condition_feedback_judgments`는 다음 판단의 관측·점수·실제 호출 수를 기록합니다.

같은 입력을 다시 판단할 필요가 없거나 남은 경로가 하나이면 호출을 생략합니다.
이미 검사한 후보는 반복하지 않습니다. 판단 도중 취소되면 기존 탐색 순서를 유지하고,
이미 수행한 모델 호출 수는 기록에 남깁니다. 한 번의 판단은1~16개 이진 선택을 처리하며,
본문 조립은 소스에 적은 예산 안에서 최대64개 후보를 확인합니다.

`body-construct --model ...`에도 같은 파일을 사용할 수 있습니다. 여기서 모델은 첫 로컬
조립에 참여하고, 이후 호출자 사례로 경로를 고르는 기존 동작을 따릅니다. 저장한
`construction.json`을 다시 실행할 때는 모델 파일을 읽거나 다시 판단하지 않습니다.
기존 출력 사례와 선언된 중간 조건은 저장 재실행에서도 확인합니다.

## 범위와 남은 일

이 모델은6,218개 FP32 매개변수를 가지며 상주 가중치 값은24,872바이트입니다.
Go API의 기존 연구에서는 같은 절댓값 변형24개를 완성하기까지 기본60개, 첫 판단35개,
조건 피드백28개 경로를 시도했습니다. [원본 기록과 비용](https://github.com/kimjooyoon/gooo-decision-runtime/tree/v0.2.30-experimental/studies/condition-search-20261010)을
함께 공개했습니다. 그 수치는 새 CLI 실행이나 독립적인 문제 정확도를 대신하지 않습니다.

이 연결의 [실제 CLI 실행 기록](research/condition-model-cli-20261010/README.md)도 공개합니다.
한 소스에서 기본 탐색4회, 모델과 피드백2회로 선언한 출력9개와 중간 조건3개를 통과했습니다.
모델 사본을 지운 뒤 저장된 본문을 실제 실행 파일로 실행해 별도10개 사례를 확인했습니다.
첫 판단10.708µs, 피드백10.625µs였고, 로딩 등을 포함한 조립은 기본1.153ms,
모델3.344ms였습니다. 이 작은 문제에서는 기본 탐색의 전체 시간이 더 짧았습니다.

v1·v2는 관측 조건을, v3는 출력 실패까지 입력으로 담습니다. CI 결과와 unfixed 피드백은
이 세 모델의 입력 형식에 없어서 `--path-feedback-ci`와 `--path-feedback-unfixed` 조합은
오류를 반환합니다. 출력 사례 자체는 모든 모델의 후보 채택에 사용됩니다. 분기별로 같은 이름의
지역 변수를 선언하는 등 현재 표현이 지원하지 않는 소스는 이유를 남기고 기본 탐색으로
이어갑니다. 레코드 모델과 이 모델은 지원하는 입력 형식이 다릅니다.

파일 해시는 내려받은 JSON을 식별하고, `model_fingerprint`는 같은 배열과 모델 형식을
식별합니다. 공백만 바뀐 JSON은 파일 해시가 달라지고 모델 지문은 같습니다. 기록에서
이 두 값을 구분하므로 별도 메타정보나 가중치 파일이 있다고 가정하지 않습니다.
