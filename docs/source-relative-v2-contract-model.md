# 작은 모델로 비교식과 반환 경로를 조립하기

[공개0.6.33](releases/0.6.33-dev.md)은 새 V2 결정 모델을 `body-context`,
`body-codegen`, `body-construct`에 연결합니다. 포함된 SDK는0.2.42이며,
앞선 V1 모델 연결도 유지합니다. 설치·배포 증거와 원래 모델의 측정 범위를 안내에서 구분합니다.

Gooo 소스에 선택할 비교식 순서, 반환 계산의 가지 배치, 출력과 Boolean
조건 사례를 적습니다. 모델은 먼저 검사할 조합을 정합니다. Gooo가 그 조합을
코드로 조립하고 사례를 검사합니다. 모델을 생략하면 같은 후보를 정해진 순서로 봅니다.

## 모델과 소스

[공개 파일과 원래 측정 결과](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1/tree/ece9b3a7934fa5aae5ca7faecb3bc035a05ba917/research/return-layout-learning-20261012/result)는
게시 당시 버전으로 고정합니다. 모델은 19,034개의 FP32 값, 76,136바이트의
가중치 배열을 사용하며 JSON 파일은 222,227바이트입니다.

```sh
hf download asketeddy/gooo-canonical-decision-tiny-v1 \
  --revision ece9b3a7934fa5aae5ca7faecb3bc035a05ba917 \
  --include research/return-layout-learning-20261012/result/model.json \
  --local-dir out/return-layout-model

model=out/return-layout-model/research/return-layout-learning-20261012/result/model.json
shasum -a 256 "$model"
```

SHA256은 다음 값입니다.

```text
8570dd8c6a2aef2676dfc027740d869797a8a11733bd7e07cdd501dc21e4a0bc
```

[예제](../examples/body-codegen/source-division-layout-cases.gooo.fixture)는
입력이 23 미만이면 3으로 나눈 정수 몫을, 그 외에는 7을 더한 값을 요구합니다.
초기 비교 순서와 가지 배치는 그 요구와 다르게 적혀 있어 두 선택이 필요합니다.
변수 선언과 대입, 조건식, 두 계산, 여섯 출력 사례와 세 조건 사례가 소스 안에 있습니다.
정수 나눗셈은 Go와 같이 0 방향으로 버립니다. 0으로 나누는 후보는 실패로 남습니다.

## 입력을 읽고 조립하기

```sh
gooo body-context --activity Choose --model "$model" \
  examples/body-codegen/source-division-layout-cases.gooo.fixture

gooo body-construct \
  --source examples/body-codegen/source-division-layout-cases.gooo.fixture \
  --entry Main --model "$model" --attempts 4 \
  --construction-cases examples/body-codegen/source-division-layout-construction.json \
  --cases examples/body-codegen/source-division-layout-evaluation.json \
  --out division-layout-first --format text

gooo body-construct --source division-layout-first/original.gooo \
  --construction division-layout-first/construction.json \
  --cases examples/body-codegen/source-division-layout-evaluation.json --format text
```

첫 명령은 후보 검사와 예측을 하지 않습니다. `source_relative_features`의
528개 값은 최초 판단에 쓰는 입력입니다. V2 형식과 16개 조건 관계가 따로
표시됩니다. 일반 조립에서는 후보 실행 전에 한 번 판단하며, 다음 후보로
넘어갈 때 모델을 다시 호출하지 않습니다. 저장 결과 재실행의 새 모델 호출은 0회입니다.
모델을 쓰지 않으려면 첫 조립 명령에서 `--model`을 생략합니다. 출력 폴더는 새 이름을 씁니다.

## 원래 관측 결과와 남은 문제

학습 전에 100개 소스와 계획을 공개했습니다. 25개 계약의 네 변형이며,
한 계약의 변형은 같은 분할에 둡니다. 60개로 한 번 학습하고, 검증 20개와
미사용 평가 20개는 학습 API에 넘기지 않았습니다. 미사용 평가는 다섯 계약입니다.

첫 조합이 출력과 조건 사례를 모두 맞힌 비율은 12/20(60%)였습니다.
고정 순서의 첫 조합은 5/20(25%)였습니다. Boolean 조건은 60/60이 맞았지만
여덟 소스에서 반환 계산을 놓는 가지가 틀렸습니다. 보존된 후보 관측값을
검사한 뒤에는 20/20이 완성됐고, 추가 입력 160/160도 맞았습니다.
첫 선택률과 검사 후 완전성은 각각 기록합니다.

원래 학습은 Go와 CPU에서 0.736초 걸렸고, 단일 예측 100개의 중앙값은
0.094ms였습니다. 학습 프로세스 최대 메모리는 18.4MB이며 전체 컴퓨터의
사용량을 뜻하지 않습니다. 이 예제의 별도 로컬 시제품 실행에서는 최초 판단
0.149ms, 지역 출력 6/6과 조건 3/3, 실제 실행 8/8을 관측했습니다.
저장 결과는 같은 실행값과 프로그램 식별값으로 재사용됐습니다.

입력은 구조화된 소스와 요구 사례입니다. 한글·영문 설명이 함께 있지만
임의 자연어의 이해 능력을 측정한 결과는 없습니다. 선언하지 않은 요구,
더 큰 언어 범위와 1.58비트 양자화 품질은 별도 근거가 필요합니다.
원래 학습·예측·실행 자료는 그대로 보존합니다.
