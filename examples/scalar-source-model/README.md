# 소스 조건과 작은 결정 모델 연결

이 변경을 포함한 개발 소스의 예제다. `main.gooo.fixture`는 입력과 결과의 합이
37이어야 한다는 조건과, 뺄셈 방향을 바꿀 수 있는 선택지 하나를 선언한다.
`helper.gooo.fixture`는 조건식에서 보조 활동을 호출한다.

## 결정론적 구성

```sh
gooo body-codegen --json --activity Offset examples/scalar-source-model/main.gooo.fixture
```

이 소스의 시도 예산은 1이다. 모델 없이 먼저 시도하는 원래 방향은 조건을 만족하지
못하므로 부분 결과가 남는다. 소스에 선언한 예산을 늘리면 다음 허용 경로도 확인할 수 있다.
불완전한 결과에서도 시도한 후보와 실제 조건 결과를 읽을 수 있다.

## 합성 가중치로 연결 살펴보기

```sh
go run ./examples/scalar-source-model/make-model --out out/scalar-source-model-synthetic
gooo body-context --activity Offset --value-flow \
  --model out/scalar-source-model-synthetic/metadata.json examples/scalar-source-model/main.gooo.fixture
gooo body-codegen --json --activity Offset \
  --path-model out/scalar-source-model-synthetic/metadata.json examples/scalar-source-model/main.gooo.fixture
```

도구는 두 번째 선택지에 점수를 더 주는 합성 파라미터를 만든다. 학습은 0회이며,
만든 파일은 학습 모델로 배포하지 않는다. 기존 디렉터리는 덮어쓰지 않는다.
가중치 파일은 74,504바이트이며 모델 로딩·소스·Go 실행 환경의 메모리는 추가된다.

새 API 검사에서는 이 파라미터로 모델을 한 번 호출하고, 선택한 한 후보가 소스 입력
2개와 별도 입력 1개의 조건을 만족했다. 파일을 삭제한 뒤 JSON으로 저장한 새 결과를
재구성했고, 수치 특징 해시가 다른 결과는 거절했다. 별도 연결 해제 소스에서는 두 후보를
결정론적으로 시도해 두 입력의 조건을 만족했다. 보조 활동이 있는 세 번째 소스는
모델 예측과 후보 평가 없이 입력 준비를 확인했다. 제작 당시 소스와 검사 기록은
`observation.json`에 연결돼 있다.

학습된 파라미터는 같은 계약으로 연결할 수 있다. 그 모델의 첫 선택 결과와 탐색 후의
조건 충족 결과는 각자의 입력 수로 평가한다. 위 합성 파라미터의 결과는 연결 검사의 범위다.

자세한 계약과 저장 결과의 재사용은 [가이드](../../docs/scalar-source-model.ko.md)를 참고한다.

## Concurrent source observation

A fourth fresh API test sent four distinct source requests through one immutable
synthetic model. Each made one prediction and satisfied selection predicates2/2
and held predicates1/1. Request receipts and prediction values remained separate;
an already-canceled request returned its cancellation. Recorded prediction times
were16,417–20,791ns and preparation/construction times2,569,208–7,015,709ns on
this PC. These four small observations do not establish trained quality, sustained
parallel capacity, process RAM or CPU utilization. Exact producer and timings are
retained in observation.json without rebinding earlier observations.

## 잘못된 모델 제안과 시도 예산

`wrong-proposal.gooo.fixture`와 `limited-proposal.gooo.fixture`는 입력에서 43을
빼라는 조건을 가진다. 위 합성 모델은 반대 방향을 먼저 제안한다. 예산 2의
첫 소스는 다음 후보에서 조건 2/2, 별도 조건 1/1을 만족했고, 예산 1의
두 번째 소스는 조건 0/2, 별도 조건 0/1과 PARTIAL_FINITE를 남겼다.
두 첫 제안의 점수는 약 95.3%였고 실제 조건은 0/2였다.

일반 출력의 상태 표시는 별도 새 `cli-partial.gooo.fixture`에서 관찰했다.

```sh
gooo body-codegen --activity Deduct \
  --path-model out/scalar-source-model-synthetic/metadata.json \
  examples/scalar-source-model/cli-partial.gooo.fixture
```

이 명령은 생성 Go 코드를 stdout에, 실제 조건 범위를 stderr에 표시한다.

```text
gooo: Deduct: PARTIAL_FINITE; selection conditions 0/2 passed, 2 failed, 0 unobserved; separate conditions 0/1 passed, 1 failed, 0 unobserved; attempts 1/1; model calls 1
```

새 관찰의 원본, CLI·모델 해시와 정확한 제작 소스는 `observation.json`에
붙였다. 생성한 프로그램의 네이티브 실행은 0회다.
