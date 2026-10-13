# 조건을 만든 조각으로 돌아가기

Gooo가 빈칸에 표현식을 채우면 완성된 본문만으로는 그 선택의 기원을 알기 어렵다.
이 예제는 실행한 조건에서 선택한 후보, 빈칸, 원래 표현식, 작성한 의도로 돌아가는 경로를 남긴다.

`quote.gooo.fixture`의 `open` 빈칸에는 `bounded3607` 후보의 `input < 61`이 들어갔다.
[실행 기록](native-consumer3607.json)의 조건은 이 표현식의 원문 바이트 위치와 연결된다.
`source_origin.fragments`는 템플릿 조각과 채워진 조각을 본문 순서로 담는다.
후보의 출처와 실행 위치가 같은 레코드에서 서로 다른 범위로 보인다.

## Gooo가 읽는 출처

`bodyexecution.ConditionFillOriginFacts(runtime)`가 기록의 위치와 ID를 옮긴다.
[`classify.gooo.fixture`](classify.gooo.fixture)의 조건문이 그 종류를 해석한다.
Go 어댑터에 분류 정책을 추가하지 않는다.

- 실제 실행에서 얻은 후보 출처 1개 → `declared`
- 별도로 작성한 해석용 입력 3개 → `derived`, `retained`, `missing`

생산자 출력은 2/2, 소비자 출력은 실제 입력 1개와 해석용 입력 3개를 합쳐 4/4였다.
이 작은 분류 예제는 출처를 다음 언어 동작의 입력으로 쓰는 연결을 보여 준다.
프로그램 전반의 출처 완전성이나 전체 입력 영역의 정확도를 측정한 수치는 아니다.

## 원래 관찰

| 서로 다른 새 소스 | 관찰한 내용 | 자료 |
| --- | --- | --- |
| fragment3371 | 동일 표현식 후보의 선택 ID, 앞선 치환에 따른 위치 이동, 이스케이프된 의도, 변경된 출처 구별 | [결과](fragment3371.json) · [생산자](producer3371.json) · [로그](observation3371.log) |
| derived3491 | 두 표현식의 문법 기원, 직접 작성된 후보 위치의 부재 | [결과](derived3491.json) · [생산자](producer3371.json) |
| origin3607 / originconsumer3607 | 실제 조건 → 후보 표현식, Gooo의 출처 해석 | [결과](native-consumer3607.json) · [생산자](producer3607.json) · [로그](observation3607.log) |
| record3751 | 두 빈칸의 서로 다른 문법 위치, 새 기록에서 선택적 출처를 비워 둔 경우의 보존 | [결과](record3751.json) · [생산자](producer3751.json) · [로그](observation3751.log) |
| workspace3889 | 미리 채운 패키지 본문 → 최종 조건, 템플릿과 후보가 섞인 조건 조각 2개 | [결과](workspace3889.json) · [생산자](producer3889.json) · [로그](observation3889.log) |

원래 생산자는 `de8b2b67` 소스와 각 manifest의 임시 Go 변경이다. 이전 소스·모델·저장
fixture를 로컬에서 재실행하지 않았다. 모델 호출은 0회다. 문법 후보의 유한 성적은
각 결과에 따로 남기며 출처 연결 정도와 합쳐서 평가하지 않는다.

앞의 `.gooo.fixture`와 cases 파일은 API 관찰에 사용한 프로그램·입력이다.
다음 단계는 패키지 병합 전 원문까지의 연결, 변수·반환식의 기원,
출처를 따라 변경 영향을 찾는 동작이다.

## 패키지 명령행

별도의 새 `discount3937` 소스를 실제 CLI로 실행했다. 본문을 먼저 채워서
패키지를 실행하는 경로에서도 최종 조건에 `construction`이 연결된다.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo package execute --json \
  --cases examples/source-fill-origins/cli3937/cases.json \
  examples/source-fill-origins/cli3937/workspace.json
```

[원래 CLI 결과](cli3937-receipt.json)의
`result.runtime.condition_sites[0].source.construction`에서 `cap3937` 후보와
`limit` 빈칸을 찾을 수 있다. 표현식의 위치는 조립 입력 문서의 `47`을 가리킨다.
원래 로컬 관찰은 먼저 빌드한 바이너리를 한 번 호출했고, 위 명령은 개발 소스 사용법이다.

출력 2/2, 조건 기원 1/1, 모델 호출 0회, 입력 파일 3개가 유지됐다.
전체 CLI 호출의 한 번 관찰값은 약 1.23초이며 프로세스 시작, 코드 생성과 두 네이티브
실행을 포함한다. 빌드한 25.9 MB 바이너리는 관찰 후 제거했다.
[관찰값](cli3937-observation.json)과 [실제 생산자](producer3937.json)를 함께 남긴다.
