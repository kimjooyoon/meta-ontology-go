# 대입과 사용을 원래 파일에서 찾기

[main.gooo](main.gooo)는 입력을 보정하고, 분기마다 다른 `result` 변수를 만든다.
헬퍼 호출과 레코드 생성도 포함한다. Gooo는 실행할 때 바뀐 함수·레코드 이름과
원래 워크스페이스 파일의 위치를 함께 남긴다. 이 기능을 포함한 개발 소스에서
저장소 루트를 기준으로 실행한다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/workspace-local-origins/cases.json \
  examples/workspace-local-origins/gooo.workspace.json > local-execution.json

go run ./cmd/gooo package execute --json \
  --origin-receipt local-execution.json --origin-kind local-input-locations \
  --origin-limit 32 examples/native-origin-reader/local-files/reader.workspace.json
```

첫 프로그램은 953에서 `amount: 1935`, -967에서 `amount: -69`를 반환한다.
둘째 프로그램은 저장된 선언·대입·사용 위치 28행을 읽어
`main.gooo:write:declaration` 같은 설명을 반환한다. 판단은
[reader.gooo](../native-origin-reader/local-files/reader.gooo)의 조건문에 있다.
이 읽기 단계에서 원래 프로그램과 모델은 실행하지 않는다.

두 `result`는 이름이 같아도 다른 선언이다. `assignment_index`는 실행 중의 대입을,
원래 파일의 바이트 범위는 소스에 적힌 위치를 나타낸다. 반복 실행은 같은 코드
위치를 여러 번 참조할 수 있다. 함수 이름이 길게 바뀌어도 원본 범위를 단순 복사하지
않고 그 이름 변환을 되돌려 계산한다.

조립으로 채운 식은 별도의 후보·문법·템플릿 관계로 설명한다. 원래 파일에 없는 파생
식의 토큰 위치를 만들어 붙이지 않는다. [입력 종류와 위치 의미](../../docs/language/native-local-file-origins.ko.md),
[새 관찰과 초기 실패](../../docs/observations/workspace-local-file-origins/README.md)를 함께 보관한다.
