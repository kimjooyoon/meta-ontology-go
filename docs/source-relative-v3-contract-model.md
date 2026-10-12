# V3에서 소스 상수와 반환값 예시의 관계 읽기

[공개 Gooo0.6.34](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.34-dev)는
SDK0.2.44의 V3 모델을 `body-context`, `body-codegen`, `body-construct` 경로에
연결합니다. 공개 소스는 `b266c0e2cf6878385b59788945861a03e33f766b`, Go는1.27.2입니다.
V2 연결도 유지합니다. [설치와 결과 읽기](https://github.com/kimjooyoon/meta-ontology-go/wiki/V3-Output-Relations-Release-0.6.34).

Gooo는 원본 조건식·반환 계산·변수와 허용된 대안을 유지합니다. 모델은 먼저
검사할 조합을 정합니다. V3는 선언된 입력·기대값과 소스 정수 상수의 관계를
비교식 피연산자의 작음·같음·큼 구간별로 정리합니다. 사칙연산 관계와 정수
overflow·zero 관계를 읽으며, 이 입력 준비 중에는 후보 본문을 실행하지 않습니다.
V2에서도 모든 반환값 예시는 별도 입력 스트림으로 모델에 전달되고 있었습니다.

입력은528개 float32 값, 가중치는19,034개 FP32 값/76,136바이트를 유지합니다.
V3는 기존 빈 영역 중48개 값과 명시적인 새 모델 형식을 사용합니다.
원래 V1·V2 파일은 각각의 기존 해석을 유지합니다.

[설명 표현의 영향을 줄이는 선택 경로](prose-neutral-contract-model.md)는
SDK0.2.45의 별도 파일 형식을 사용합니다. 이 개발 연결은 원본 설명을
기록하면서 모델 입력에서 설명문128개 값만 제외합니다. 공개0.6.34의
일반 V3 경로와 새 개발 경로의 지원 범위를 안내에서 구분합니다.

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
공개 Gooo0.6.34에 이 파일을 `--model` 또는 `--path-model`로
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

## 공개 릴리스의 확인 범위

[원본 핵심 CI](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38227814300),
[원본 native 준비](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38227814162),
[원본 발행](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38231010673)이
같은 공개 소스로 완료됐습니다. 네 플랫폼의 V3 회귀는 같은8입력을 반복하며
독립 행동 입력의 분모를 늘리지 않습니다.

공개9파일 식별값·8체크섬·7출처 목록·소스 태그와 최종 발행 기록을 대조했습니다.
필요한6파일의 바이트를 확인하고 실제 macOS arm64 배포 파일을 설치했습니다.
다른 세 플랫폼 압축 파일은 API·체크섬 식별값을 대조했습니다.
원래 학습·예측·native 실험은 문서 갱신을 위해 다시 실행하지 않았습니다.

[보존된64평가 행의 언어·연산별 세부 결과](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1/tree/028ad98923f3d05a5404604bf33d89adde454c64/research/output-relations-v3-learning-20261012/audit)는
한국어 첫 성공11/32→8/32, 영어10/32→10/32를 보여 줍니다. 같은64행의 사후
분류이며 표현의 인과 효과나 일반적인 언어 이해율로 읽지 않습니다. 다음 실험은
한영 표현·소스 배치·매개변수를 분리한 새 계약과 반대 요구를 사전에 계획합니다.
