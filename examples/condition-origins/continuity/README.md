# 오래된 기록의 출처 부재를 유지하기

새 `dispatch3179` 소스를 정적으로 준비해 원본 파일의 조건 위치를 연결했다.
같은 프로그램의 출처 부가 정보가 없는 표현과 현재 표현을 비교하며, 예전 기록에
없던 출처를 덧쓰지 않고 프로그램의 나머지 부분을 계속 대조한다.

- 현재 출처와 같은 기록, 출처가 없던 기록을 각각 읽는다.
- 기록된 출처의 변경·삭제·빈 배열, 다른 본문·진입점·패키지·활동·범위·추가 필드는 거절한다.
- 큰 정수 차이와 잘못된 JSON도 구별한다.
- 프로그램 실행·모델 호출·과거 fixture 재실행은 0회다.

[관찰](observation.json), [로그](observation.log), [실제 생산자](producer.json)를 남긴다.
생산자는 `4ca8181` 커밋과 임시 Go 변경 5개다. 이 관찰을 이후 공개 커밋의 실행으로
바꿔 적지 않는다. 전체 저장 기록 호환성은 새 PR의 원래 CI에서 확인한다.

[앞선 PR #1529](https://github.com/kimjooyoon/meta-ontology-go/pull/1529)는 기존 기록의
프로그램 비교에서 실패해 병합하지 않았다. 원래 실행 결과를 그대로 남겼다.

후속 `ticket3233` 소스는 receipt의 원래 JSON으로 비교한다. 타입으로 읽는 동안
출처의 부재·`null`·빈 배열이 하나로 합쳐지지 않도록 했고, 알 수 없는 필드도 유지해 비교한다.
[후속 로그](receipt-observation.log)와 [생산자](receipt-producer.json)에 정적 관찰을 남겼다.
앞 관찰에서 사용한 [비교 함수](platform_package_profile_program.go.fixture)와
[검사 소스](platform_package_program_origin_test.go.fixture)는 변경 전 모습으로 보존한다.
두 사례를 로컬에서 다시 실행하지 않았다.
