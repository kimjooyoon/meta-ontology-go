# 이름으로 읽고, 위치로 연결하는 입력

개발 소스에서 다음처럼 계산의 방향을 이름으로 적을 수 있습니다.

```gooo
activity Difference(minuend Integer, subtrahend Integer) -> Integer
  id "urn:gooo:example:named-inputs:difference" computes `
return minuend - subtrahend
`
```

실제 실행 파일은 [main.gooo](main.gooo)입니다. `Credit`이 원금에 23을
더하고, `Net`이 그 결과와 외부 공제액을 `Difference`에 전달합니다.

공개 [0.6.35-dev](../../docs/releases/0.6.35-dev.md) 컴파일러를 설치하고,
같은 버전의 소스 체크아웃 루트에서 실행합니다.

```sh
gooo package execute --json \
  --cases examples/named-inputs/cases.json examples/named-inputs/workspace.json
gooo body-origins --activity Difference --format text \
  examples/named-inputs/main.gooo
gooo package interface --json examples/named-inputs/workspace.json
```

0.6.35 게시 전 원본 네 플랫폼 검사에서 실행과 저장 재실행을 확인했습니다.
입력 사례는 2개, 각 실행의 출력 조건은 4개이며 모델 호출은 0회입니다.
[공개 설치와 관찰 범위](../../docs/releases/0.6.35-dev.md#public-installation-observations)에 연결했습니다.

이전 개발 소스의 첫 실행에서 별도 입력 사례 2개, 출력 조건 4개가 모두 맞았고 모델 호출은
0회였습니다. 로컬 명령 벽시계 시간은 1.109초였습니다. 이 사례에 관한
관찰이며 컴파일러 전체의 성능이나 정답률을 나타내지는 않습니다.

출처 보기에는 `input0`의 타입이 원문 7행 29열의 `Integer`, 본문 이름이
7행 21열의 `minuend`로 나타납니다. 타입이 같은 두 입력도 원래 위치를
각각 보관합니다.

`bind Credit.result -> Net.input0`과 외부 입력 키 `named/inputs:Net.input1`은
선언 순서로 연결됩니다. 본문 이름은 `원금`, `공제`입니다. 입력 이름을
바꾸는 작업이 활동의 안정된 ID나 연결 위치를 새로 만들지는 않습니다.

[문법과 지원 범위](../../docs/language/named-inputs.ko.md)를 참고하세요.

## 의미와 원문을 보존하며 파일 나누기

개발 소스에서 빌드한 CLI는 `Net`의 안정된 ID를 경계로 파일 분할을 제안합니다.
위의 공개 0.6.35 릴리스에는 이 명령이 포함되어 있지 않습니다.

```sh
go run ./cmd/gooo package split-source --json \
  --package named/inputs --source main.gooo \
  --before urn:gooo:example:named-inputs:net --into net.gooo \
  examples/named-inputs/workspace.json > /tmp/named-inputs-split.json
```

결과의 `result.files`에는 두 파일의 내용과 원본 바이트 구간이,
`updated_workspace`에는 새 파일까지 등록한 목록이 담깁니다. 제안 명령은
파일을 그대로 둡니다. 변경하지 않는 소스의 본문은 JSON에 중복 저장하지 않으며,
패키지 입력 이미지 식별자로 실제 workspace와의 관계를 보관합니다.

에이전트가 저장된 제안을 적용하고 기존 사례를 다시 실행할 수 있습니다.

```sh
go run ./cmd/gooo package split-source --apply /tmp/named-inputs-split.json \
  --json examples/named-inputs/workspace.json
go run ./cmd/gooo package execute --json \
  --cases examples/named-inputs/cases.json examples/named-inputs/workspace.json
```

적용은 현재 소스와 목록을 다시 읽어 제안을 재계산합니다. 새 경로가 이미
존재하거나 원본·목록·다른 패키지 입력이 바뀌었으면 쓰기 전에 거부합니다.
임시 파일 준비 후에도 패키지 전체를 다시 확인하고, 두 소스와 목록을 함께
게시합니다. 원본 권한은 유지하고 새 파일은 원본 소스 권한을 따릅니다.
게시·디렉터리 동기화 실패 시 기존 쓰기 기능으로 대상 세 파일을 복구합니다.
`application.state`는 이 대상들의 이전 상태 유지 여부를 보고하며,
현재 출력 전체를 다시 읽고 패키지 이미지까지 맞았을 때 `OUTPUT_VERIFIED`가 됩니다.

이 쓰기 경로는 workspace를 단독으로 수정하는 작업을 전제로 합니다.
여러 파일의 교체 사이를 읽는 외부 독자, 마지막 확인 이후의 외부 수정,
프로세스 종료·전원 중단까지 원자적으로 처리하는 보장은 아직 없습니다.
`STATE_DIFFERS`나 `VERIFICATION_FAILED`는 현재 파일을 조사해야 하는 결과이며,
적용 성공으로 집계하지 않습니다. 실행 결과를 전송하지 못한 경우에도 같은
제안을 다시 적용해 덮어쓰지 않으므로, 에이전트가 현재 출력과 저장 제안을
비교한 뒤 후속 작업을 결정합니다.

이전 형식의 전체 runtime manifest가 들어 있는 제안도 읽습니다. 그 복사본을
쓰기 입력으로 삼지 않고 현재 패키지에서 다시 만든 두 결과를 사용합니다.
기존 패키지 receipt 수신기의 plain JSON 32 MiB, gzip 2 MiB·복원 JSON 32 MiB
제한을 사용하며, 입력 제안의 해시는 압축 전 JSON 바이트를 가리킵니다.

본문·주석·줄바꿈은 그대로 옮기고 패키지·네임스페이스·import 헤더만 복사합니다.
안정된 ID, 입력 이름·타입, 필드와 진입점은 분할 전후에 비교합니다. 같은 파일의
`bind`가 파일 간 연결로 바뀌므로 파일 해시와 패키지 이미지 해시는 달라집니다.
새 파일의 위치에서 원본 위치로 돌아가는 바이트 대응을 별도로 보관합니다.

이 제안은 workspace 목록으로 읽는 V4 패키지를 대상으로 합니다. Go의
`go:embed`나 단일 파일만 읽는 생성기를 자동 변경하지는 않습니다. 이름 충돌,
잘못된 경계, 선언 뒤에 놓인 import는 적용 가능한 제안으로 내보내지 않습니다.
별도 프로세스의 프로그램 실행 결과는 적용 명령 안에서 검사하지 않으며, 기존 사례와
독립된 예상 출력으로 확인하는 과정이 필요합니다.
