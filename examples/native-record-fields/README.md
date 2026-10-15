# 같은 값이어도 기원은 다를 수 있습니다

이 프로그램은 레코드를 복사하고, 헬퍼에서 왼쪽 필드를 바꿉니다.
`flip`이 켜지면 두 필드를 서로 바꾼 뒤 오른쪽 값을 읽습니다.
입력에 따라 최종 숫자가 둘 다 `0`이어도, 그 숫자가 지나온 필드는
서로 다를 수 있습니다.

Gooo의 개발 소스는 이제 실행 중 각 레코드 필드의 쓰기와 읽기를 기록합니다.
읽기는 그 변수·필드에서 마지막으로 완료된 쓰기를 가리킵니다.
복사와 헬퍼 인자 전달, 지역 레코드 반환에는 필드별 부모 쓰기를 남깁니다.
레코드 전체 대입은 오른쪽 계산이 끝난 뒤 기원 슬롯도 함께 바꿉니다.
그래서 두 필드를 서로 바꾸는 대입도 갱신 전 값을 읽습니다.

## 실행하고 Gooo로 읽기

현재 개발 소스로 빌드한 `gooo`를 사용해 저장소 루트에서 실행합니다.

```sh
gooo package execute --json --cases examples/native-record-fields/cases.json \
  examples/native-record-fields/gooo.workspace.json > field-execution.json

gooo package execute --json --origin-receipt field-execution.json \
  --origin-kind field-uses --origin-limit 8 \
  examples/native-record-fields/reader.workspace.json
```

첫 명령은 네 가지 입력으로 프로그램을 실행하고, 두 실행의 기록을
비교합니다. 두 번째 명령은 저장된 읽기 기록 여덟 개를 `reader.gooo`의
입력으로 전달합니다. 원래 프로그램의 실행 횟수는 조회 중 `0`입니다.

Gooo 독자는 처음 들어온 입력 필드를 `input:...`, 부모 쓰기와 연결된
필드를 `linked:...`, 그 밖의 완료된 쓰기를 `observed:...`로 설명합니다.
이름과 값이 같더라도 변수 바인딩·필드 ID·프레임·쓰기 번호로 구분합니다.
이 작은 독자는 한 기록의 직접 연결을 설명합니다. 전체 기원 그래프의
탐색 범위와 미해결 경로는 별도로 확인해야 합니다.

`--origin-kind field-writes`는 쓰기 기록을 전달합니다.
`--origin-offset`과 `--origin-limit`으로 다음 페이지를 읽을 수 있습니다.
입력은 최대 32KiB이므로 긴 표현식이 있는 기록은 작은 페이지로 읽습니다.

## 무엇을 볼 수 있나요?

- `field_bindings`: Gooo 레코드·필드의 안정적 ID와 지역 변수/입력의 정체성
- `field_sites`, `field_use_sites`: 쓰기·읽기의 정적 지점
- `traces[].field_writes`, `field_uses`: 실제 경로의 쓰기와 읽기 인스턴스
- `parent_assignment`: 복사·인자 전달·지역 레코드 반환에서 연결된 부모 쓰기
- `completion`, `completion_clock`: 완료된 쓰기와 그 뒤의 읽기 순서
- `field_observation`: 기록 수, 연결 수, 미해결 부모, 중단과 누락 수

관측용 Go 코드의 바이트 범위는 `projection_*`로 표시합니다.
지역 변수의 선택된 Gooo 선언은 `field_bindings[].declaration`에서
확인할 수 있습니다. 두 좌표는 서로 다른 소스를 가리킵니다.

## 범위와 기록 상한

현재 대상은 필수 단일 정수·불리언·문자열 필드로 이뤄진 평평한 레코드입니다.
기록에는 필드 값과 문자열 본문을 복제하지 않습니다. 쓰기는 4,096개,
읽기는 65,536개, 함수 프레임은 65,536개까지 기록합니다. 상한에 도달하면
계산을 이어가고 누락량과 `PARTIAL` 상태를 남깁니다. 도중에 계산이 실패한
필드 대입은 미완료 쓰기로 남고, 이전 완료 쓰기를 덮어쓰지 않습니다.

스칼라 지역 변수를 거친 간접 의존성, 활동 간 그래프 입력 연결,
복잡한 레코드 인자 표현식, 지역 변수 이외의 레코드 반환,
선택·반복 필드는 확장이 필요합니다. 일반 레코드 식은
`shared_rhs`로 표시하고, 부모가 연결되지 않으면 `unknown_parents`에 셉니다.
`COMPLETE`는 현재 기록 프로필에서 누락이 없는 상태를 뜻합니다.

실험 내용과 처음 발견한 문제는
[관측 기록](../../docs/observations/native-record-fields/README.md)에 있습니다.
실행 입력 없이 가능한 필드 의존성을 살펴보는 방법은
[형식 기반 기원 예제](../typed-value-origins/README.md)에서 설명합니다.
