# 조건의 기준값 관계를 조립에 전달하기

이번 개발 소스는 SDK0.2.41의 관계 모델을 `body-context`, `body-codegen`,
`body-construct`에 연결합니다. 공개 컴파일러0.6.28에는 앞선 정규화 분기 모델 연결이
들어 있습니다. 아래 명령은 이번 기능이 포함된 소스 체크아웃과 Go1.27.2를 사용합니다.

`backlog < 8`에서 입력이 기준보다 작은지·같은지·큰지를 요구한 참·거짓과 함께
모델에 넘깁니다. 모델은 소스에 표시한 비교 순서와 분기 배치 중 먼저 볼 조합을
추천합니다. 실제 코드 조립과 요구 검사는 Gooo가 수행합니다.

## 모델 파일과 작은 예제

[공개 모델](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1)의
관계 가중치 하나를 사용합니다. 학습이나 별도 모델 서버를 시작할 필요가 없습니다.

```sh
hf download asketeddy/gooo-canonical-decision-tiny-v1 \
  --revision be6d95747d3a498620d0a5a6c42b4284c964cba2 \
  --include research/source-relative-learning-20261011/result/model-relative.json \
  --local-dir out/source-relative-model

model=out/source-relative-model/research/source-relative-learning-20261011/result/model-relative.json
shasum -a 256 "$model"
```

파일의 SHA256은 다음 값입니다. FP32 가중치 배열은76,136바이트이고,
메타데이터와 숫자의 문자열 표현을 포함한 JSON 파일은222,730바이트입니다.

```text
cae32bdad17ab818498c6099066918cd9c9c591eb118afe1bc6a7ef82dc83cc4
```

[배치 처리량 예제](../examples/body-codegen/source-relative-batch.gooo.fixture)는
대기 수가8보다 작으면3을 더하고,8이상이면두배를 반환하는 작은 시연용 규칙입니다.
소스의 초기 분기 배치와 기대 출력이 다르게 적혀 있어 조립 과정에서 선택이 필요합니다.
변수 대입·비교식·두 반환 계산·검사 사례가 모두 Gooo 소스에 있습니다.

## 실제 입력 먼저 읽기

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-context \
  --activity Choose --model "$model" \
  examples/body-codegen/source-relative-batch.gooo.fixture
```

`source_relative_features`는 모델이 실제로 읽는528개 값입니다. 그중 `[192:208]`의
16개 값이 조건 관계이며, `source_relative_conditions`에 그 값의 개수·구간·가용
이유가 표시됩니다. `canonical_branch_context`는 해석한 비교식과 반환식을 보여줍니다.
원래 소스·계획의 식별값과 출력·Boolean 조건 배열도 따로 남습니다.
이 명령의 모델 예측과 후보 검사는0회입니다.

## 조립하고 다시 사용하기

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-construct \
  --source examples/body-codegen/source-relative-batch.gooo.fixture \
  --entry Main --model "$model" --attempts 4 \
  --construction-cases examples/body-codegen/source-relative-batch-construction.json \
  --cases examples/body-codegen/source-relative-batch-evaluation.json \
  --out relative-batch-first --format text

GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-construct \
  --source relative-batch-first/original.gooo \
  --construction relative-batch-first/construction.json \
  --cases examples/body-codegen/source-relative-batch-evaluation.json --format text
```

출력 폴더는 새 이름을 씁니다. 첫 조립에서 `--model "$model"`을 생략하면 같은
소스와 요구를 정해진 순서로 검사합니다. 모델을 쓰는 경우 각 지역 조립 세션은 후보
검사 전에 한 번 판단하고, 같은 세션의 다음 후보로 넘어갈 때는 다시 판단하지 않습니다.
저장 재실행은 모델 파일을 읽지 않고 현재 입력을 실제 Go 값으로 실행합니다.

조건이 서로 맞지 않거나 예산 안에 답을 찾지 못하면 부분 결과와 실패가 남습니다.
명령의 성공 여부와 조립 결과의 완전성은 [결과 보고서](construction-results.md)에서
함께 읽습니다. 사례가 없는 항목은 미측정으로 남깁니다.

## 범위와 근거

입력은 한 루트 분기, 정수 비교 `less_than`·`less_equal`·`equal`, 최대 두 선택과
128개 조건을 지원합니다. 직접 반환과 지원하는 변수 대입은 같은 계산 입력을 사용합니다.
지원 밖 소스는 이유와 함께 고정 순서로 이어가며 모델 호출은0회입니다. 모델 파일 형식이
틀리면 오류를 반환합니다. 모델 점수는 후보를 정렬하는 값입니다.

[원래288개 연구](https://github.com/kimjooyoon/gooo-decision-runtime/tree/v0.2.41-experimental/studies/source-relative-learning-20261011)에서
관련 첫 선택은64/128이었고 간단한 규칙도64/128이었습니다. 새 연결은 일반 SDK 후보 큐를
사용하므로 원래 네 조합 정렬 실험의 시간과 후보 수를 그대로 적용할 수는 없습니다.
이번 네이티브 연결의 별도 관찰은 [고정 계획](research/source-relative-native-20261011/protocol.md)을 따릅니다.
[원래 실행 결과](research/source-relative-native-20261011/results.md)에서 모델과 고정 순서는
모두 후보2개를 확인했습니다. 첫 출력 검사는1/4, 최종4/4였으며 저장 재실행은 새 모델 호출0회로
같은 네이티브 사례4개를 통과했습니다. 판단 시간·가중치 준비·전체 실행 비용은 따로 기록했습니다.
