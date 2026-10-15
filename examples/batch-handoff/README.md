# 선택한 목록을 다음 활동으로 넘기기

작업 시간을 앞에서부터 읽고, 예산 안에 들어오는 부분을 `Select`가 반환한다.
`Summarize`는 그 목록을 받아 개수와 합계를 다시 계산한다. 계산에 쓰는 목록이
활동 사이를 실제로 이동하므로, 선택 개수만 보고 다음 작업을 재구성할 필요가 없다.

```gooo
bind Select.result -> Summarize.input
```

`main.gooo`에는 두 활동의 본문과 연결, 경계값 비교식의 후보가 함께 있다.
`input.durations[:accepted]`의 끝 위치는 반복문에서 계산한다. 예산을 초과하거나
음수 시간이 나오면 그 앞까지 반환한다. 예산과 합계가 같으면 포함하며, 빈 목록도
유효한 결과다. 남은 예산과 비교하므로 큰 양의 정수 합계를 먼저 만들지 않는다.

개발 소스로 설치한 Gooo에서:

```sh
gooo body-compose --source examples/batch-handoff/main.gooo \
  --cases examples/batch-handoff/cases.json --out out/batch-handoff
```

예산 19, 작업 시간 `[6, 13, 4]`를 주면 `Select`는 `[6, 13]`, `Summarize`는
`{"count":2,"total":19}`를 반환한다. 입력 순서를 `[13, 4, 6]`으로 바꾸면
선택 결과는 `[13, 4]`, 합계는 17이다. 각 작업을 쪼개거나 순서를 바꾸지는 않는다.

## 조립과 기원

비교식은 소스의 `source_fill`에서 `<=`와 `<` 후보로 선언한다. 선택용 사례 세 개와
별도 보류 사례 하나를 사용하며, 모델 없이 결정론적으로 구성한다. 실행 입력 일곱
개는 이 네 개의 조립 입력과 구분한다.

워크스페이스 실행 기록의 출처 조회는 `Summarize`에서 bind를 따라 `Select`의
선택된 비교식까지 연결하도록 구성했다. 반복문의 정적 분석은 원소와 경로를 합쳐
보므로 `PARTIAL`을 유지한다. 각 원소의 실행 경로를 개별적으로 증명하는 지표는
아직 제공하지 않는다.

[첫 실행과 테스트 수정 기록](../../docs/observations/collection-handoff/README.md)에
관찰 범위와 남은 확인을 적었다. 이 예제는 개발 소스에 포함되며 새 릴리스 배포는
별도 단계다.
