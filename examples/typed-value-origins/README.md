# 필드별 기원을 Gooo로 읽기

레코드를 복사한 뒤 왼쪽 필드를 바꾸고 오른쪽 필드를 반환하면,
오른쪽 결과는 어디에서 왔을까요? 이 예제의 `Compute`는 원래 입력의
`right`를 반환합니다. `left`에 대입한 다른 입력은 그 결과의 데이터
기원에 포함되지 않습니다. 레코드 전체를 한 묶음으로 추적하면 이 둘이
섞이기 쉽습니다.

`body-origins`는 일반 `computes` 본문의 형식을 확인하고 필드별 그래프를
만듭니다. 복사, 필드 대입, 고정된 같은 파일의 헬퍼 호출, 분기의 합류를
표현합니다. `ancestry.gooo`는 이 그래프를 받아 반환값까지 이어지는
입력 포트와 필드 ID를 직접 고릅니다.

## 실행

저장소 루트에서 실행합니다. 아래 `gooo`는 현재 소스로 빌드한 CLI입니다.

```sh
gooo package execute --json --cases examples/typed-value-origins/cases.json \
  examples/typed-value-origins/gooo.workspace.json

gooo body-origins --activity Compute examples/typed-value-origins/main.gooo

gooo body-origins --activity Compute --inputs-for typed/ancestry:Inputs \
  examples/typed-value-origins/main.gooo > ancestry-inputs.json

gooo package execute --json --inputs ancestry-inputs.json \
  examples/typed-value-origins/ancestry.workspace.json
```

마지막 프로그램의 반환 문자열은 다음과 같습니다.

```text
input0:urn:gooo:typed-origins:right;
```

`Select`는 조건에 따라 `right`를 바꿉니다. 기본 출력의 데이터 기원은
`input0.right`와 `input1`입니다. `--include-control`을 추가하면 그 선택에
사용되는 `input2`도 탐색합니다. 실행 입력을 주지 않은 분석이므로
소스에서 가능한 양쪽 경로를 유지합니다.

레코드 반환값은 `--field <안정적 필드 ID>`로 한 필드를 골라 입력으로
내보냅니다. 전체 조회 결과의 `outputs`에 사용 가능한 ID와 루트가 있습니다.

## 역할과 범위

Go는 형식과 필드 ID, 복사·대입·호출·분기의 연결을 읽습니다. 입력으로
내보낼 때는 노드와 간선, 번호에 대응하는 비트를 배열에 담습니다.
Gooo는 배열 길이·순서·비트·간선 범위를 확인하고, 그래프를 역순으로
따라가 기원 목록을 만듭니다. 반환값은 입력 포트와 필드의 조합이므로
서로 다른 두 필드에 같은 숫자를 넣어도 기원이 합쳐지지 않습니다.

직접 조회는 512노드, 64개 동시 바인딩, 분기·헬퍼 깊이 16까지입니다.
직접 조회는 필수 단일 스칼라 필드와 중첩 레코드를 다룹니다.
중첩 레코드는 별도의 형식·값·변경 이력 사전을 가진 v2 보고서로 표현합니다.
분석이 지원하지 않는 반복문 등의 본문은 `UNRESOLVED`와 이유를 반환하고
성공한 접두 그래프를 내보내지 않습니다.

이 Gooo 탐색기는 정수 한 개에 집합을 담아 60노드를 처리합니다.
내보내기는 이를 넘거나 32KiB 입력 제한을 넘으면 오류를 반환합니다.
큰 그래프를 작게 잘라 완전한 기원처럼 표시하지 않습니다.

노드 위치에는 원래 `.gooo` 리터럴의 바이트 범위와 파일 내용 해시가
있습니다. 입력 포트는 별도 ID로 표시하며 표현식 위치가 없습니다.
이 조회는 모델 선택이나 프로그램 실행을 하지 않습니다. 저장된 실행의
지역 변수 기원은 [workspace 예제](../workspace-local-origins/README.md)에서
별도로 다룹니다. 필드별 실제 실행 경로 추적은 남은 작업입니다.

관측 결과와 처음 실패한 시도는
[실험 기록](../../docs/observations/typed-value-origins/README.md)에 있습니다.

중첩 v2 보고서를 Gooo로 읽는 과정은 [레코드 출처 탐색기](../record-origin-reader/README.md)에서
실행합니다. 원래 전체 그래프를 페이지로 전달하고 Gooo가 작업 단계를 계획합니다.
이전 60노드 입력 형식은 v2 사전을 생략해서 내보내지 않고 명시적으로 거부합니다.
