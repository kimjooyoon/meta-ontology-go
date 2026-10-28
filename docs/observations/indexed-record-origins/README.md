# 인덱스 원문의 기원 관측

이 자료는 `36c55e72d0699ad6397d5a1c8b0cac1d7e980227` 위에 인덱스 출처 지원을
추가한 로컬 후보의 관측이다. 전체 저장소의 exact-head CI와 dev 반영은 별도 단계다.

- 새 예제의 값 기대 4개가 실제 Go 실행과 저장 재실행에서 4/4를 만족했다. 두 실행이 분모를 늘리지 않는다.
- 원문 출처는 16노드이며, Gooo 조회가 5/5단계에 `items`와 `offset`의 원본 입력을 찾았다.
- 한글 변수와 인덱스식의 위치를 보존하고, 읽지 않는 필드를 제외했다.
- 상수·계산·스칼라 목록, 실제 실행의 잘못된 범위, 512노드 초과, 기존 평면 입력 거절을 검사했다.
- Gooo 조회의 생성된 Go 실행은 원본 그래프를 읽는 독립 스택 탐색의 방문 집합과 대조했다.
- 기존 bodycodegen·탐색기 전체 테스트, 새 인덱스 경로 race와 등록·렌더링 한도 검사를 통과했다.

`native-values.json.gz`와 `source-origins.json.gz`는 각각 원래 JSON 바이트의 비손실
보관본이다. 원래 스키마와 소스·실행·투영 해시를 유지하며 `source-query.json`이
그 원본 보고서를 해시로 참조한다. 모델 호출은 0회다. 시간·RAM 개선을 측정한 자료는 아니다.

로컬 로그의 출처:

- `full-codegen-reader.log`: SHA-256 `8de348dd99df686aa13c66ccc3a9e1ff726b42c419a0b28771cfa4ef41a0b243`.
- `registration-capacity.log`: SHA-256 `29eff5e7f1234aaf220165d85e88949e9ee798f76e203bb869f7015e68ef7e2a`.
- `final-index-race.log`: SHA-256 `846a4087ce5ec3a1c2fac7671bfff414849a58dd6527342301651c4f177c1eb5`.
