# ID 표기와 기원을 함께 유지하기

같은 URI가 대소문자가 다른 표기로 들어왔을 때, 그 필드의 소유자는 URI의 의미로
비교한다. 소스에 쓴 ID와 선언의 바이트 위치는 남긴다.

이 사례는 URI를 비교하는 기존 함수로 필드의 소유자 ID를 해석한다.
모델의 출처 표현과 의미상 URI 표현을 각각 읽는다. 새 source3001에서 원본 소스를
유지하면서 Get-Put·Put-Get, 다른 소유자의 거절, 필드 출처2/2와 기대값2/2를 관찰했다.
모델 호출은0회다. 실제 생산자는 공개4ca8181 소스와 Go 변경3개다.

[기록](observation.json), [생산자](producer.json), 첫 사례의
[실패](first-failed-observation.log)와 [생산자](first-failed-producer.json)를 함께 남긴다.
첫2867 사례는 URI 소유자 해석과 native2/2를 통과한 뒤, activity ID가 원본 표기와
같아야 한다는 테스트의 잘못된 가정에서 실패했다. 그 원래 관찰을 재실행하지 않고
다른3001 소스에서 URI 의미와 원본 선언을 나누어 확인했다.

CLI 파일 묶음은 API로 관찰한 본문을 제공한다. 이 CLI 묶음의 로컬 재실행은 포함하지 않는다.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo package execute --json \
  --cases examples/source-owner-identity/cases.json \
  examples/source-owner-identity/workspace.json
```

[발견한 문제 #1530](https://github.com/kimjooyoon/meta-ontology-go/issues/1530)
