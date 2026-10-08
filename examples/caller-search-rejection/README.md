# 잘못된 계산식을 기록하고 다음 후보 고르기

개발 소스의 예제입니다. 공개 0.6.12 실행 파일에는 아직 포함되지 않았습니다.

함수의 `0 → 0`만 보면 여러 계산식이 맞습니다. 전체 프로그램이 `3 → -1`을
기대할 때, 첫 후보는 `1`을 반환합니다. 다음 `0` 후보는 분모가 0이라 컴파일할 수
없습니다. 그 이유를 기록하고 세 번째 `-input`을 고르면 호출 결과가 맞습니다.

```sh
go run ./cmd/gooo body-construct \
  --source examples/caller-search-rejection/main.gooo.fixture --entry Main \
  --construction-cases examples/caller-search-rejection/construction-cases.json \
  --cases examples/caller-search-rejection/evaluation-cases.json \
  --attempts 5 --out /tmp/my-caller-search-rejection
```

세 번을 시도하고 두 프로그램을 실제 실행합니다. 마지막 평가는 네 입력입니다.
`--attempts 2`로 줄이면 첫 프로그램과 미해결 결과를 남깁니다. 실패한 후보도
시도 한도를 사용하며, 실행하지 않은 결과에는 정확도 0%를 붙이지 않습니다.

```sh
go run ./cmd/gooo body-construct \
  --source /tmp/my-caller-search-rejection/original.gooo \
  --construction /tmp/my-caller-search-rejection/construction.json \
  --cases examples/caller-search-rejection/evaluation-cases.json
```

저장 재실행은 거절된 계산식과 실패 이유까지 다시 확인합니다. 새 모델 호출은
없습니다. 거절을 포함한 기록은 `gooo/joint-construction/v3`이고, `rejection`의
`slot`은 `candidate_kinds`의 위치입니다. 해당 행의 지역 검사 수는 거절 지점 앞에서
실제로 검사한 부품만 셉니다. 거절된 조합은 최종 프로그램으로 선택하지 않습니다.

`mixed-model.gooo.fixture`와 `mixed-` 사례 파일은 이 계산식에 레코드 선택 세 개를
붙입니다. `--attempts 40`을 사용하고, `--model`을 지정하면 기존 자체 그래프 모델이
레코드를 살펴볼 순서를 제안합니다. 계산식의 순서는 결정론적으로 유지합니다.

이 변경은 지역 계산식의 타입·순수 평가 실패를 다룹니다. 취소, 실행 도구 오류,
컴파일한 프로그램의 실행 오류와 소스 재구성 실패는 계속 중단 사유입니다.
초기 지역 조립에서 유효한 후보를 하나도 만들 수 없는 경우도 중단합니다.
[원래 실패·수정 후 실제 결과·모델 사용](../../docs/research/caller-search-rejection-20261009).
