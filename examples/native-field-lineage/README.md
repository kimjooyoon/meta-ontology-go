# 필드 기록의 기원을 Gooo로 따라가기

프로그램에서 계산된 값이 어느 입력 필드에서 왔는지 찾아보는 예제입니다.
컴파일러는 실행 중 필드의 쓰기·읽기와 복사 연결을 남깁니다.
`lineage.gooo`는 이 기록을 받아 부모 연결을 따라가고, 도달한 입력 필드를 모읍니다.
주소를 적은 운송장을 받아 물건이 지나온 창고를 거슬러 올라가는 것과 비슷합니다.

## 실행

저장된 작은 실행 기록을 사용합니다. 원래 프로그램과 모델을 호출하지 않습니다.
저장소 루트에서 Go 1.27.3로 실행합니다.

```sh
saved="$(mktemp)"
gzip -dc examples/native-field-lineage/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" \
  --origin-shape field-lineage-columns --origin-offset 12 \
  examples/native-field-lineage/gooo.workspace.json
rm "$saved"
```

첫 케이스의 마지막 `Read`가 읽은 필드를 선택합니다. 결과의 `roots`에는
`urn:gooo:native-fields:build|input|urn:gooo:native-fields:input-right;`가
나옵니다. `OBSERVED`와 `finite_total=0`은 입력만 받은 실행이라는 뜻입니다.
테스트의 기대값과 비교한 결과는 별도의 [관찰 기록](../../docs/observations/native-field-lineage/README.md)에 있습니다.

## 역할

| 담당 | 하는 일 |
| --- | --- |
| Gooo 생산 프로그램 | 필드가 있는 값과 계산 경로를 선언 |
| Go 관측 코드 | 필드 ID, 부모 쓰기, 읽기 구간, 실제 완료 순서를 기록 |
| 열 전송기 | 선택된 읽기가 속한 케이스 전체를 배열로 전달 |
| `Trace` 활동 | 역순 탐색, 입력 기원 선택, 누락·불확실 상태 계산 |

전송기는 조상 목록이나 최종 기원을 미리 계산하지 않습니다.
복사와 인자·반환 연결은 직접 부모로 따라가고, `scalar_rhs`는 그 필드가
평가되는 동안 기록된 필드 읽기를 따라갑니다. 배열 예약 순서 대신 실제
완료 순서를 거꾸로 순회하므로 안쪽 헬퍼가 먼저 끝나는 경우도 다룰 수 있습니다.

## 결과를 읽는 법

- `SCOPED_FIELD_ORIGINS`: 현재 기록 프로필 안의 연결을 따라간 결과입니다.
- `PARTIAL_LINKS`: 공유 평가 창이나 알려지지 않은 연결이 남았습니다.
- `PARTIAL_RECORDING`: 원래 관측에서 누락·미완료가 있었습니다.
- `UNRESOLVED_READ`: 선택된 읽기가 가리키는 쓰기를 알 수 없습니다.
- `INVALID_INPUT`: 배열 모양·필드 정체성·완료 순서 등이 서로 맞지 않습니다.

`root_count`, `reachable_writes`, `parent_links`, `field_read_edges`,
`unresolved_links`, `shared_rhs_writes`, `scalar_rhs_writes`를 따로 남깁니다.
스칼라 지역 변수나 분기 제어를 거친 모든 인과관계까지 설명하는 지표는 아닙니다.
`source_sha256`과 `journal_sha256`은 받은 기록의 출처 참조입니다.
기록 생산자의 신원을 인증하지는 않습니다.

## 현재 크기

한 케이스에 쓰기 60개·읽기 128개까지 받습니다. 60비트 방문 표와 두 단계
배열 순회를 사용합니다. 큰 케이스를 잘라 완전한 것처럼 전달하지 않고,
상한 밖이면 전송을 거절합니다. 다음 작업은 더 큰 기록을 여러 페이지로
읽으면서 경계를 넘어가는 기원 연결을 유지하는 것입니다.

[언어 설명](../../docs/language/native-field-lineage.ko.md) ·
[실행 관찰](../../docs/observations/native-field-lineage/README.md)
