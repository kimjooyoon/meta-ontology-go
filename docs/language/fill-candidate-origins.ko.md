# 선택 전 후보의 기원

Gooo의 `source_fill`은 의도, 빈자리와 허용할 대입을 소스에 적는다. 후보가
선택되기 전에도 사람이 쓴 식과 본문의 빈자리를 함께 읽을 수 있다. 후보를 고르는
과정에서 문제가 생겼을 때, 어디에 적은 선택인지부터 돌아보는 용도다.

```sh
gooo body-origins --activity Choose --fill-candidate wrong_type \
  examples/fill-candidate-origins/source.gooo
```

보고서의 각 행에는 빈자리 ID, 후보의 식, `hole` 선언 범위와 본문 사용 범위가
있다. 파일 해시와 활동 ID, 후보 ID가 그 행의 소스를 설명한다. 바이트 범위는
원래 UTF-8 파일의 시작 포함·끝 제외 구간이다. 문자열의 이스케이프가 있으면
디코딩한 본문 좌표와 원본 파일 좌표를 각각 보관한다.

`true`를 Integer 본문의 빈자리에 넣는 후보도 선언·사용 위치를 읽을 수 있다.
이는 위치가 연결됐다는 관찰이다. 후보가 타입 검사를 통과했다거나 사례에서
몇 퍼센트 맞았다는 점수는 생성 보고서에 따로 남는다. 이 조회는 생성 보고서나
실행에서 거부된 이유를 대신 만들지 않는다.

## Gooo가 그 자료를 해석하기

```sh
gooo body-origins --activity Choose --fill-candidate wrong_type \
  --inputs-for origin/fill-candidates:Inspect \
  examples/fill-candidate-origins/source.gooo > candidate-origins.json
gooo package execute --json --inputs candidate-origins.json \
  examples/fill-candidate-origins/gooo.workspace.json
```

Go가 원시 위치와 식을 전달하면 예제의 `inspect.gooo`가 조건식으로
`DECLARED_ASSIGNMENT`와 `UNRESOLVED_USE`를 구분한다. 예제의 해석을 바꾸려면
Gooo 소스를 고친다. 입력 전용 실행에는 기대값이 없으므로 `finite_total=0`이며,
이 예제의 출력 행 비교와 프로그램 전체의 기능 정답률을 합치지 않는다.

## 범위가 분명하지 않을 때

주석이나 문자열 속 이름은 본문 토큰 사용으로 세지 않는다. 실제 토큰이 없거나
반복되면 `TEMPLATE_HOLE_NOT_UNIQUE`와 `UNRESOLVED`를 출력하고, 사용 위치는
비워 둔다. 종료 코드는 실패이며 출력된 기록은 남는다. 내보낸 행의 `has_use`가
빈 위치를 구분한다. Gooo 소비자는 그 불확실성을 그대로 받을 수 있다.

직접 선언 조회는 소스 문법과 같은 1..8개 빈자리를 다룬다. 문법에서
생성된 후보는 [문법 기원 조회](derived-candidate-origins.ko.md)로 원래 문법의 위치를 읽는다. 원본 소스는 128KiB, 입력 내보내기는
32KiB 범위이며 초과분을 생략하지 않고 오류를 낸다. 기존 타입 기원 조회의 `--field`,
`--include-control`은 이 모드와 함께 쓰지 않는다.

[실제 예제](../../examples/fill-candidate-origins/README.md) ·
[위치와 Gooo 소비자 관찰](../observations/fill-candidate-origins/README.md) ·
[선택 후 필드 기원 조회](origin-subject-queries.ko.md)
