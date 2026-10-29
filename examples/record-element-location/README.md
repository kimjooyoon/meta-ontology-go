# 저장된 원소 위치를 Gooo에서 선택하기

`main.gooo`의 `Locate`는 기록에서 받은 `position`, `container`, `status`를 읽습니다.
출처가 해석되었고 `container`가 `urn:record-items:items`이면 원래 배열 위치를 반환합니다.
해석되지 않았거나 다른 배열의 기록이면 `-1`을 반환합니다.

예를 들어 `{ "position": 1, "container": "urn:record-items:items", "status": "RESOLVED" }`의 결과는 `1`입니다.
같은 값을 가진 원소도 기록된 위치 `0`과 `1`로 구별합니다.

```sh
gooo check --semantic main.gooo
gooo roundtrip --json main.gooo
```

`internal/bodyexecution/record_iteration_element_test.go`는 실제로 수집한 원소 위치 8행과
별도로 작성한 거절 조건 2개를 이 프로그램으로 실행합니다.
저장된 기록에서 행을 읽는 동안 원래 생산자나 모델을 다시 실행하지 않습니다.
여러 필드와 복사 경로까지 따라가는 예제는 `examples/local-field-bridge`의 `TraceIndexed`입니다.
