# 저장된 기원을 Gooo로 읽기

Gooo 프로그램이 자신과 다른 Gooo 프로그램의 계산 기록을 입력받아 설명하는 예제다.
기록을 만든 프로그램의 출처와, 이를 읽는 프로그램의 출처가 함께 남는다.

이 기능은 이 변경을 포함한 개발 소스에서 사용할 수 있다. 저장소 루트에서 실행한다.

```sh
go run ./cmd/gooo package execute \
  --origin-receipt examples/native-origin-reader/construction9533-runtime.json \
  --origin-kind elements --origin-offset 6 --origin-limit 5 \
  examples/native-origin-reader/elements/elements.workspace.json
```

[원소 설명](elements/elements.gooo)은 `let`과 조건문으로 계산 중 읽기가 없었던
경우, 한 번 읽은 경우, 여러 번 읽은 경우, 계산이 중단된 경우를 나눈다.
판단 규칙을 바꾸려면 이 `.gooo` 파일을 수정한다.

## 기원 입력

| 종류 | 읽는 내용 | Gooo 예제 |
| --- | --- | --- |
| `local-fill-fragments` | 실행된 대입을 조립한 원래 후보·문법·템플릿 조각 | [조립 출처 설명](local-fill/README.md) |
| `elements` | 생성한 원소의 계산 범위, 완료 여부, 원본 표현 | [원소 설명](elements/elements.gooo) |
| `constructions` | 생성한 목록, 빈 목록, 시도한 원소 수 | [목록 설명](constructions/constructions.gooo) |
| `reads` | 읽은 위치, 외부 입력 또는 생성한 원소의 기원 | [읽기 설명](reads/reads.gooo) |
| `links` | 같은 사례 안의 읽기 → 생성 인스턴스 → 원소 연결 | [연결 설명](links/links.gooo) |
| `local-uses` | 지역 변수의 실제 사용 → 마지막 대입 연결 | [지역 변수 설명](locals/locals.gooo) |

각 하위 폴더에는 독립적인 워크스페이스가 있다. `--origin-kind`와 해당 폴더의
`*.workspace.json`을 함께 선택한다. `--json`을 넣으면 원래 기록의 해시, 선택한
페이지와 원래 누락 수를 포함한 실행 결과를 얻는다.

빈 목록은 `construction9547-runtime.json`과 `constructions` 예제로 살펴볼 수 있다.
계산 중단은 `construction9551-runtime.json`과 `elements` 예제로 볼 수 있다.

지역 변수의 사용과 재할당을 보려면 다음 저장 기록을 읽는다.

```sh
go run ./cmd/gooo package execute --json \
  --origin-receipt examples/native-origin-reader/local10433-runtime.json \
  --origin-kind local-uses --origin-offset 0 --origin-limit 2 \
  examples/native-origin-reader/locals/locals.workspace.json
```

이 두 행은 각각 `first`, `second`의 사용을 해당 대입에 연결한다.
이 예전 기록에는 사용 표현의 원래 위치가 없으므로 두 행의 `classification`은
`declaration-only`다. 선언과 대입의 연결은 유지한다. 원래 생산 프로그램과 관찰 출처는
[지역 변수 관찰](../../docs/observations/native-local-origins/README.md)에 있다.

기본 페이지 크기는 32행이다. `next page`가 출력되면 그 offset으로 이어 읽는다.
최대 128행이며, 인코딩한 입력이 32KiB를 넘으면 limit를 줄이라는 진단이 나온다.
저장 기록은 최대 32MiB까지 읽는다.

## 기록의 출처

세 JSON은 [앞선 목록 생성 관찰](../../docs/observations/construction-origins/README.md)의
`runtime` 필드를 꺼낸 것이다. [추출 경로와 해시](record-origins.json)를 보관했다.
이 예제 실행에서는 새 설명 프로그램만 실행한다. 원래 프로그램의 실행 횟수는 0이다.

`NO_READ_IN_WINDOW`에는 상수뿐 아니라 앞서 계산된 지역 변수를 사용한 경우도
포함된다. 지역 변수 관찰이 포함된 새 기록은 `local-writes`, `local-uses`, `local-elements`로
이 관계를 더 살펴볼 수 있다. [지역 변수 기원 안내](../../docs/language/native-local-origins.ko.md)를 참고한다.
원래 기록의 `PARTIAL`,
중단과 누락 수는 새 결과에도 남는다. 정답을 입력하지 않은 실행 결과의 상태는
`OBSERVED`이며, 별도 테스트의 기대 결과 일치 수는 [관찰 기록](../../docs/observations/native-origin-inputs/README.md)에 적었다.

자세한 입력 계약은 [기원 입력 안내](../../docs/language/native-origin-inputs.ko.md)를 참고한다.

## 원래 사용 표현이 포함된 새 기록

```sh
go run ./cmd/gooo package execute --json \
  --origin-receipt examples/native-origin-reader/local10773-runtime.json \
  --origin-kind local-uses --origin-offset 0 --origin-limit 2 \
  examples/native-origin-reader/locals/locals.workspace.json
```

이 기록은 반복문과 조건식을 가진 프로그램에서 얻었다. 읽는 두 행은 `assignment`로
분류된다. 지역 변수 입력의 `source_sha256`, `source_start/end`, `source_mapping`으로
사용 표현의 Gooo 위치를 다룰 수 있다. [관찰과 출처](../../docs/observations/local-use-source/README.md)를 참고한다.
