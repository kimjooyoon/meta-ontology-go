# 후보를 선택하기 전에 출처 읽기

`source.gooo`는 두 빈자리와 두 후보를 선언한다. `wrong_type`의 `true`는 Integer
본문에 맞지 않지만, 그 식을 쓴 위치와 들어갈 빈자리의 위치는 먼저 읽을 수 있다.

```sh
gooo body-origins --activity Choose --fill-candidate wrong_type \
  examples/fill-candidate-origins/source.gooo
gooo body-origins --activity Choose --fill-candidate wrong_type \
  --inputs-for origin/fill-candidates:Inspect \
  examples/fill-candidate-origins/source.gooo > candidate-origins.json
gooo package execute --json --inputs candidate-origins.json \
  examples/fill-candidate-origins/gooo.workspace.json
```

Go는 파일 해시·활동 ID·후보 ID와 원본 바이트 범위를 옮긴다. `inspect.gooo`는
그 자료를 받아 `DECLARED_ASSIGNMENT` 또는 `UNRESOLVED_USE`로 분류한다.
선언 위치와 본문 사용 위치가 모두 있다고 해서 후보의 동작이 올바른 것은 아니다.
정적 조회에서는 후보 선택·타입 검사·사례 평가·모델 호출을 수행하지 않는다.
입력 전용 실행의 기능 평가 분모는 `finite_total=0`이다.

주석과 문자열 안의 빈자리 이름은 실제 사용으로 세지 않는다. 실제 토큰이 없거나
여러 곳에 나타나면 `UNRESOLVED`와 `TEMPLATE_HOLE_NOT_UNIQUE`를 남기고,
본문 위치는 비워 둔다. 보고서는 출력하지만 명령 종료 코드는 실패다.
내보낸 원시 행의 `has_use=false`가 없는 위치를 구분하므로 Gooo도 이를 해석할 수 있다.

현재 조회는 소스에 직접 적은 후보와 1..8개 빈자리를 다룬다. 문법에서 생성된
후보는 직접 선언된 대입 위치가 없으므로 이 모드가 거부한다. 내보낼 전체 자료가
32KiB를 넘으면 생략하지 않고 오류를 낸다. 실행에서 거부된 이유·점수는 원래의
생성 보고서에서 별도로 읽어야 한다.

[사용 범위](../../docs/language/fill-candidate-origins.ko.md) ·
[실험 기록](../../docs/observations/fill-candidate-origins/README.md)
