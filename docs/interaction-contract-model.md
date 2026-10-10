# 출력과 중간 조건을 함께 읽는 모델

이 문서는 SDK 0.2.39 연결을 추가한 개발 소스 기준입니다. 공개 0.6.26의
선택별 사례 모델 사용법은 [해당 버전 안내](releases/0.6.26-dev.md)에 있습니다.

Gooo 소스에 선택 가능한 분기와 기대 출력을 적고, `condition_case`로 중간 비교의
참·거짓도 적습니다. 자체 모델은 이 세 정보를 함께 읽어 처음 확인할 경로를 고릅니다.
이후에는 Gooo가 작성된 출력과 조건을 검사하며 남은 후보를 진행합니다.

## 준비된 모델로 시작하기

[SDK 0.2.39 배포](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.39-experimental)의
`model-interaction-v1.json`을 사용합니다. 가중치 배열은 FP32 19,034개, 76,136바이트입니다.
파일과 실행 중 전체 메모리의 크기는 이 배열보다 큽니다.

개발 저장소 루트에서:

```sh
mkdir -p out/interaction-model
curl -fL https://github.com/kimjooyoon/gooo-decision-runtime/releases/download/v0.2.39-experimental/model-interaction-v1.json \
  -o out/interaction-model/model-interaction-v1.json
printf '%s\n' '4b61fe8d84df77c8dfac6a880eedcddfd8f5f903fa538d1e80532dd77f74b1ac  out/interaction-model/model-interaction-v1.json' \
  | shasum -a 256 -c -
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o out/gooo ./cmd/gooo

out/gooo body-context --activity Choose \
  --model out/interaction-model/model-interaction-v1.json \
  examples/body-codegen/source-interaction-condition-cases.gooo.fixture

out/gooo body-codegen --json --activity Choose \
  --path-model out/interaction-model/model-interaction-v1.json --path-step-attempts 1 \
  examples/body-codegen/source-interaction-condition-cases.gooo.fixture
```

입력 확인은 모델 판단과 후보 검사를 하지 않습니다. 조립은 처음 한 번 판단하고,
진행 기록의 `predictions_this_advance`는 0으로 유지합니다. 모델을 생략하면 고정 순서로
같은 요구를 검사합니다. 손상됐거나 지원하지 않는 모델 파일을 지정하면 읽기 오류를 반환합니다.

## 실제 실행과 저장 재사용

```sh
out/gooo body-construct --entry Main \
  --source examples/body-codegen/source-interaction-condition-cases.gooo.fixture \
  --model out/interaction-model/model-interaction-v1.json \
  --construction-cases examples/body-codegen/source-condition-construction-cases.json \
  --cases examples/body-codegen/source-condition-evaluation-cases.json \
  --attempts 8 --out out/interaction-construction

out/gooo body-construct \
  --source out/interaction-construction/original.gooo \
  --construction out/interaction-construction/construction.json \
  --cases examples/body-codegen/source-condition-evaluation-cases.json
```

저장 재실행은 모델 파일을 읽지 않고 새로운 판단도 하지 않습니다. 이 예제의 후속
`Main` 활동은 선택한 `Choose`의 결과에 원래 입력을 더합니다. 큰 정수의 정확한 값은
소스·실행·기록에서 유지되며 모델에는 버전이 정해진 특징 배열을 전달합니다.

## 무엇을 읽었는지 확인하기

| 채널 | 명시적 입력 확인 결과 |
| --- | --- |
| 선택별 소스·피연산자 순서 | `inputs[].ordered_source_features`, 선택마다 528칸 |
| 기대 출력 | `declared_contract_cases.features`, 각 사례의 32칸 전체 |
| 중간 조건의 기대 참·거짓 | `declared_condition_cases.features`, 0~128개 조건의 32칸 전체 |

일반 조립 기록에는 입력의 식별값과 개수가 남습니다. 배열 전체는 명시적 입력 확인에서
공개합니다. 입력 확인과 실제 최초 판단은 같은 인코더를 사용하고 식별값을 대조합니다.
조건이 0개인 경우도 빈 조건 채널의 식별값으로 기록합니다.

모델이 표현하지 못하는 소스는 `DECLINED_TO_DETERMINISTIC` 사유를 남기고 모델 호출
없이 고정 순서로 진행합니다. 예를 들어 기존 `source-condition-cases.gooo.fixture`의
별도 산술식 선택 지점은 이 모델의 순서 표현 범위 밖입니다. 위 예제는 그 산술식을
고정하고 비교·분기를 선택합니다. 모델 형식을 넓히려면 입력 표현과 학습 자료를 함께 바꿔야 합니다.

## 확인 범위

연결부 회귀 검사는 합성 가중치로 입력 배열·호출 횟수·네이티브 조립·모델 없는 재실행을
확인합니다. 학습 품질은 [SDK의 원래 실험](https://github.com/kimjooyoon/gooo-decision-runtime/tree/01dc72229552260f0e2da10a0a312865a8b9ed32/studies/interaction-requirement-learning-20261011)에
성공과 실패를 함께 기록했습니다. 연결부 테스트 통과율을 새 정확도 수치로 세지 않습니다.
규칙을 바꾼 뒤 저장 결과는 [결과 비교 명령](workflow-outcome-delta.md)으로 읽을 수 있습니다.
