# V3에서 소스 상수와 반환값 예시의 관계 읽기

현재 소스 체크아웃은 공개 SDK0.2.44의 V3 모델을 `body-context`,
`body-codegen`, `body-construct` 경로에 연결합니다. 설치된 공개 Gooo0.6.33은
SDK0.2.42와 V2 연결을 사용합니다. V3 공개 컴파일러 배포는 별도 단계입니다.

Gooo는 원본 조건식·반환 계산·변수와 허용된 대안을 유지합니다. 모델은 먼저
검사할 조합을 정합니다. V3는 선언된 입력·기대값과 소스 정수 상수의 관계를
비교식 피연산자의 작음·같음·큼 구간별로 정리합니다. 사칙연산 관계와 정수
overflow·zero 관계를 읽으며, 이 입력 준비 중에는 후보 본문을 실행하지 않습니다.
V2에서도 모든 반환값 예시는 별도 입력 스트림으로 모델에 전달되고 있었습니다.

입력은528개 float32 값, 가중치는19,034개 FP32 값/76,136바이트를 유지합니다.
V3는 기존 빈 영역 중48개 값과 명시적인 새 모델 형식을 사용합니다.
원래 V1·V2 파일은 각각의 기존 해석을 유지합니다.

## 공개된 새 학습 결과

[원본 소스와 분할](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1/tree/d63359863c7a685eca81239e00626aebef3a7b05/research/output-relations-v3-learning-20261012/prefit),
[Go 수집·학습 코드](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1/tree/af76ace87ed5fdbbb10adbdb9a86ac1584fb1694/research/output-relations-v3-learning-20261012/code),
[두 모델과 실패를 포함한 전체 결과](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1/tree/089c7b3200ee80c55a721f16580985054ce997e1/research/output-relations-v3-learning-20261012/result)를
각 단계에서 먼저 계획하고 공개했습니다.

같은 새 학습128개 문서로 V2·V3를 각각 한 번 학습했습니다. 평가64개 문서는
학습에 없었던 네 반환 연산 조합의 여덟 계약을 한영 설명·소스 배치로 반복한 것입니다.

| 평가 지표 | 새 V2 | 새 V3 |
| --- | ---: | ---: |
| 첫 후보로6출력+3Boolean 요구 충족 | 21/64 | 18/64 |
| 한영 첫 조합 일치 | 11/32쌍 | 5/32쌍 |
| 기록된 후보 검사 수 | 140/64 | 152/64 |
| 유한 후보 검사 후 요구 충족 | 64/64 | 64/64 |

이 묶음에서 V3 선택 품질은 더 낮았습니다. V3는 명시적으로 선택하는 실험
경로로 유지합니다. 추가 입력의512개 관찰은64개의 독립 계약/입력 쌍을 반복한
값입니다. 전반적인 자연어 이해, 임의 프로그램, 속도·CPU·전체 RAM 개선은
이 자료에서 측정하지 않았습니다.

## 파일 선택과 실행

```sh
hf download asketeddy/gooo-canonical-decision-tiny-v1 \
  research/output-relations-v3-learning-20261012/result/v3/model.json \
  --revision 089c7b3200ee80c55a721f16580985054ce997e1 --local-dir model-v3
```

모델 파일의 SHA256은
`683a19f86fb3a7b5b4e436d1b692884b96b613652caa693e321818bde8645816`입니다.
현재 소스에서 빌드한 Gooo에 이 파일을 `--model` 또는 `--path-model`로
명시해 사용합니다. 구체적인 원본 소스·호출·평가 파일은
[V2 실행 안내](source-relative-v2-contract-model.md#모델과-소스)의 명령과 같습니다.

입력 검사는 모든6출력·3Boolean 요구와48개 관계, 소스·예시 해시를 보여 주고
모델 판단·후보 검사·학습은0회입니다. 조립 중에는 후보 검사 전에 한 번 판단합니다.
모델 옵션을 생략하면 결정론적으로 진행합니다. 지원하지 않는 모델 입력은
그 이유와 결정론적 경로를 유지하고, 잘못된 모델 파일은 오류를 반환합니다.
저장 결과 재실행은 모델 파일 없이 수행하며 새 판단은0회입니다.

새 영점 가중치 회귀 검사는 소스 검사, 원본 바디 조립, 별도 native8/8,
모델 삭제 후 저장 결과8/8과 큰 Integer 값의 유지를 확인합니다.
학습된 모델의 위 평가 결과와 회귀 검사 결과는 각각의 분모로 읽습니다.
