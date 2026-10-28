# 지역 변수에 담아 둔 값의 기원 따라가기

이 예제는 필드를 지역 변수에 저장한 뒤 원래 필드를 바꿉니다.
마지막 출력은 저장해 둔 값에서 조립합니다.

```gooo
let kept_score=original.score
original.score=input.right
let moved_score=kept_score+7
let assembled=State{score:moved_score,live:moved_live,tag:moved_tag}
```

`assembled.score`의 기원을 따라가면 `original.score`의 이전 쓰기에 도달해야
합니다. 이번 읽기에서 사용한 쓰기 번호를 두 종류의 기록 사이에 남겨
이 경로를 연결합니다. 정수·불리언·문자열을 함께 다룹니다.

## 새 입력으로 실행

저장소 루트에서 Go 1.27.3를 사용합니다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/local-field-bridge/cases.json \
  examples/local-field-bridge/gooo.workspace.json
```

네 케이스에는 양수와 같은 값 `0`에 도달하는 서로 다른 분기가 있습니다.
이번 예제는 선언한 계산을 사용하며 모델을 호출하지 않습니다.

## 저장된 기록을 Gooo로 읽기

`trace.gooo`가 두 종류의 쓰기와 읽기를 직접 순회합니다. 다음 명령은
저장된 첫 케이스에서 `forwarded.score`를 선택합니다.
원래 생산 프로그램의 실행 횟수는 0입니다.

```sh
saved="$(mktemp)"
gzip -dc examples/local-field-bridge/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-shape field-local-columns \
  --origin-offset 15 examples/local-field-bridge/trace.workspace.json
rm "$saved"
```

결과의 `result.runtime.traces[0].deliveries[0].actual`에서 다음을 볼 수 있습니다.

- `roots`: `urn:gooo:local-bridge:build|input|urn:gooo:local-bridge:input-left;`
- `status`: `SCOPED_FIELD_LOCAL_ORIGINS`
- `carrier_writes`: 거쳐 간 지역 변수 쓰기의 수
- `field_read_edges`, `local_read_edges`: 따라간 두 종류의 읽기 수
- `unresolved_links`, `collection_rhs_writes`: 아직 연결하지 못한 부분

입력만 받은 소비자의 결과는 `OBSERVED`, `finite_total=0`입니다.
별도 테스트에서 예상 기원과 비교한 결과는 [관찰 기록](../../docs/observations/local-field-bridge/README.md)에 있습니다.

## 레코드 목록의 원래 위치까지 조회

`TraceIndexed`는 기존 쓰기·읽기 열과 원소 읽기 표를 함께 받습니다.
`Trace`와 `TraceIndexed`는 같은 Gooo 활동 `TraceCore`에서 경로를 따라갑니다.
Go 전송기는 기록 번호와 형식·필드 ID를 옮깁니다.

다음은 같은 값의 두 항목 중 두 번째 항목의 수량·가격 읽기를 조회합니다.
출력 폴더는 아직 없는 경로를 지정합니다.

```sh
go run ./cmd/gooo body-compose --source examples/record-collections/main.gooo \
  --entry Total --cases examples/record-collections/cases.json \
  --out /tmp/gooo-indexed-items > /tmp/gooo-indexed-items.json
go run ./cmd/gooo package execute --json \
  --origin-receipt /tmp/gooo-indexed-items.json \
  --origin-shape field-local-columns-v2 --origin-offset 4 --origin-limit 2 \
  examples/local-field-bridge/indexed.workspace.json
```

두 결과의 `indexed_low`는 각각 `4`, `8`입니다. 원래 쓰기 번호 `2`, `3`을
선택한 값입니다. 선택된 번호는 입력의 `elements.refs`를 통해 원래 읽기 번호로
이어집니다. `read_ids`에서 같은 번호를 찾으면 `indices`가 원래 목록 위치 `1`,
`field_ids`가 목록의 필드 ID, `owner_ids`가 입력 레코드 ID를 가리킵니다.
`record_ids`는 목록 항목의 명목 형식 ID입니다.

`indexed_low`는 쓰기 `0..31`, `indexed_high`는 `32..59`의 선택 결과를 담는
양의 32비트 정수입니다. 두 값은 일반 JSON 숫자로 정확히 표현됩니다.
이 선택과 원래 기록의 연결은 화면에 표시할 문장이나 구분자 해석에 의존하지
않습니다. `summary.roots`는 기존 필드 입력의 표시 경로를 유지합니다.

원소 읽기의 출처가 미해결이거나 생성된 목록까지 연결되면 `PARTIAL_LINKS`,
읽기가 유실되면 `PARTIAL_RECORDING`을 반환합니다. 실제 원소 위치가 확인된
쓰기만 선택합니다. 중복 읽기 번호, 명목 형식 불일치, 빠진 참조는
`INVALID_INPUT`으로 처리합니다.

실제 기본 예제의 8개 필드 읽기와 중첩 반복의 12개 읽기는 해석 실행과 별도
네이티브 실행에서 같은 선택을 냅니다. 중첩 사례의 복사 읽기 4개는 원래 필드
쓰기를 거쳐 같은 값이 든 두 목록을 구별합니다. 쓰기 번호 32, 미해결·생성된 원소, 읽기 유실,
빠진 참조·형식 불일치·중복 번호는 별도 작성한 8개 입력으로 확인합니다.
CLI의 4개 조회는 기존 읽기의 재조회이며 새로운 생산 사례로 세지 않습니다.

## 연결과 상태

컴파일러는 기존 지역 변수·필드 기록을 유지하면서 교차 읽기 구간과 공통
완료 시계를 남깁니다. 두 기록의 함수 프레임 번호는 각자의 체계 안에서
사용합니다. 교차 연결은 실제 케이스별 쓰기 참조로 이어집니다.
Go 전송기는 두 읽기 구간을 32개 원시 열로 옮기고, Gooo가 역순 탐색과
입력 기원 선택을 수행합니다.

지역 변수는 `scalar_local`과 이름·선언 순번·원시 타입으로 표시합니다.
그 열의 `record_ids`와 `field_ids`는 비워 둡니다. 원래 선언 좌표는
생산 기록의 `local_sites`와 `local_use_sites`에 남습니다.

기록이 끊기면 `PARTIAL_RECORDING`, 연결되지 않은 출처가 있으면
`PARTIAL_LINKS`가 됩니다. 컬렉션 부분을 아직 연결하지 못해도 함께 읽힌
필드·지역 변수의 알려진 기원은 결과에 남깁니다. 단순 상수 표현식은
`literal_writes`로 따로 셉니다.

각 조회 입력은 선택된 케이스 전체를 받습니다. `--origin-limit 3`은 연속한
세 필드 읽기를 세 입력으로 보내 한 번에 실행합니다. 입력별 합계 쓰기
60개·읽기 128개, 전송 페이지 전체 JSON 32KiB가 현재 소비자 상한입니다. 큰 기록을 조용히 잘라 전달하지
않습니다. 원래 두 기록은 각각 쓰기 4,096개·읽기 65,536개의 저장 상한을
유지하며, 상한 이후에도 실제 프로그램의 계산은 계속됩니다.

분기 제어의 의존성, 헬퍼 인자·반환의 정밀 선택, 활동 그래프 사이의 레코드
전달, 큰 기록의 페이지 탐색과 생성된 목록의 전체 입력 기원은 남은 작업입니다.
평가 중 관측한 읽기 구간을 따라가는 범위를 결과와 함께 봅니다.

[언어 설명](../../docs/language/local-field-bridge.ko.md) ·
[관찰 기록](../../docs/observations/local-field-bridge/README.md)
