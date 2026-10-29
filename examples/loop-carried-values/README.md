# 반복해서 전달되는 값의 출처

`Shift`는 배열을 읽으면서 `left`와 `right`의 이전 값을 서로 전달한다.
입력이 비어 있으면 처음 `left`를 반환한다. 한 번 반복하면 처음 `right`를,
두 번 반복하면 처음 `left`와 첫 원소의 `value`를 더한 값을 반환한다.

```sh
go run ./cmd/gooo check --semantic examples/loop-carried-values/main.gooo
go run ./cmd/gooo body-compose --entry Shift \
  --cases examples/loop-carried-values/cases.json --source examples/loop-carried-values/main.gooo
go run ./cmd/gooo body-origins --activity Shift examples/loop-carried-values/main.gooo
```

`cases.json`에는 9개 결과를 독립적으로 적었다. 길이 0~5인 배열의 결과는
`3, 11, 10, 24, 27, 43`이다. 나머지는 읽지 않는 필드의 값 변경, `2^53`보다
큰 정수의 정확한 계산, 같은 값을 가진 원소의 반복을 다룬다.

출처 검사는 입력값을 실행하지 않고 원문을 읽는다. 일반 부모 간선은 앞선
노드만 가리키고, 다음 반복으로 돌아오는 관계는 `loop_carries`에 별도로 둔다.
처음 값인 `seed`, 반복문의 입력인 `header`, 다음 값인 `yield`, 반복 뒤의
`exit`를 구별한다. 따라서 두 변수의 상호 의존과 0회 반복을 함께 보존한다.

원소 위치의 `symbolic_any`는 원문에서 가능한 반복 위치를 뜻한다. 실제 실행에서
기록한 배열 위치와 구별한다. `ignored`는 원문에서 읽지 않는 필드다.
이 관계는 소스의 가능한 의존성이며, 실제 실행의 PROV-O 파생 사실과 분리한다.

현재 구현은 필수 배열 필드의 `for … in …`과 중첩 반복, 레코드 필드의
갱신, 조기 반환을 다룬다. [직접 만든 배열](../constructed-record-values/README.md)의
읽기 전용 순회도 지원한다. 배열 자체의 반복 전달, 카운터 반복문과
선택 필드의 정적 출처는 남은 작업이다. 기존 512노드·64바인딩 한도를 유지하며,
범위를 넘으면 부분 그래프를 성공으로 내보내지 않는다.

반복의 초기 값·본체 인자·다음 값을 분리하는 선행 설계로
[MLIR SCF의 반복 인자와 `scf.yield`](https://mlir.llvm.org/docs/Dialects/SCFDialect/)를
참고했다. 여기서는 Gooo 원문의 출처를 보존하는 별도 관측 IR로 표현한다.

인덱스식으로 직접 원소를 고르는 출처는
[인덱스 값 예제](../indexed-record-values/README.md)에서 확인한다.
