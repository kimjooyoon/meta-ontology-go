# 호출자의 실패로 계산식 다시 고르기

개발 소스에서 실행하는 예제입니다. 공개 0.6.12는 `body-construct`의 조립 대상으로
레코드 선택만 받습니다. 이 변경은 소스의 `search hole` 계산식도 함께 탐색합니다.

```sh
go run ./cmd/gooo body-construct \
  --source examples/caller-ir-search/main.gooo.fixture --entry Main \
  --construction-cases examples/caller-ir-search/construction-cases.json \
  --cases examples/caller-ir-search/evaluation-cases.json \
  --attempts 8 --out /tmp/gooo-caller-hole
```

`Choose`의 예제 `0 → 0`은 `input`과 `0`을 모두 허용합니다. `Main`의 조건을
실행하면 차이가 드러납니다. 첫 프로그램은 `3 → 6`이라 실패하고, 계산식을 `0`으로
바꾸면 `3 → 3`을 충족합니다. 마지막 세 입력은 조립 조건과 겹치지 않습니다.

`mixed-model.gooo.fixture`는 이 계산식과 세 개의 레코드 선택을 함께 조립합니다.
같은 디렉터리의 `mixed-construction-cases.json`, `mixed-evaluation-cases.json`과
`--attempts 40`을 사용합니다. `--model`에 자체 그래프 모델을 지정하면 레코드의
초기 순서만 모델이 제안합니다. 계산식 순서는 소스의 문법과 한도에서 결정됩니다.

## 어떤 범위를 확인하나

- 소스에 선언한 문법이 계산식을 만들고 `max_candidates`로 남길 후보 수를 제한합니다.
- 각 함수의 `attempts`는 그 목록에서 전체 조립에 사용할 접두 범위를 제한합니다.
- 명령의 `--attempts`는 전체 프로그램 조합의 시도 횟수입니다. 실행 전에 거절된 조합도 포함하며 함수의 한도를 넘기지 않습니다.
- 계산식의 실제 값과 지역 예제를 별도로 보존합니다. 호출부가 맞아도 지역 조건이 남으면 부분 결과입니다.
- 계산식의 생성·누락 수는 각 `search_candidates[].candidate_generation`에 남습니다.
- 정수 IR 검색이 포함되면 기본 기록 버전은 `gooo/joint-construction/v2`입니다.
  `candidate_kinds`가 `masks`의 각 숫자를 레코드 비트마스크 또는 계산식 목록 인덱스로 구분합니다.
  `search_candidates`는 계산식 ID·표현식·지역 실제 값·소스와 계획 해시를 따로 저장합니다.
- 지역 계산식 검사에서 거절된 후보가 있으면 v3으로 기록하며, 이유와 후보를 보존하고 다음 조합을 시도합니다.
- 기존 레코드 전용 기록은 v1을 유지합니다. 세 버전 모두 저장 기록의 재실행에서 새 추론은 0회입니다.

[거절된 계산식 예제](../caller-search-rejection/README.md)는 `0으로 나누기` 후보를 지나 다음 계산식을 고릅니다.
소스 재구성 실패·취소·실행 도구 오류·컴파일된 프로그램의 실행 오류는 실패 기록과 함께 중단합니다.
여러 홀을 한 번에 채우는 `source_fill`과 정수 구조 선택은 별도 연결이 필요합니다.
이 예제의 유한한 통과 수는 모든 입력이나 자연어 요구에 대한 완전성을 뜻하지 않습니다.
