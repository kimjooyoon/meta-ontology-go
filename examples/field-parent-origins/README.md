# 복사된 필드 값의 바로 앞 기원

레코드를 복사하거나 헬퍼에 넘겼다가 돌려받으면, 같은 필드에도 여러 값 버전이
남습니다. 이 예제의 Gooo 프로그램은 저장된 읽기가 가리킨 쓰기에서
**실제로 기록된 부모 쓰기 한 건**을 선택합니다.

예를 들어 쓰기14가 쓰기17의 값을 받았을 수 있습니다. 쓰기 번호는 예약 순서이고,
완료 순서는 별도로 보관하므로 번호의 크기로 앞뒤를 판단하지 않습니다.

## 개발 소스에서 사용

이 조회는 `field-parent-columns`를 구현한 개발 소스에서 Go1.27.2로 실행합니다.
공개0.6.35 바이너리는 이 새 전송 옵션을 제공하지 않습니다. 저장소 루트에서:

```sh
saved="$(mktemp)"
gzip -dc examples/local-field-bridge/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-shape field-parent-columns \
  --origin-offset 15 examples/field-parent-origins.workspace.json
rm "$saved"
```

결과의 `result.runtime.traces[0].deliveries[0].actual`에는 `child=14`,
`parent=17`과 RDF문장4개가 있습니다. 입력만 받은 조회의 결과는 `OBSERVED`,
`finite_total=0`입니다. 원래 생산 프로그램은 실행하지 않으며 모델도 호출하지 않습니다.

## Gooo가 하는 일

`parent.gooo`는 기존 `fieldlineage.Trace` 검사와
`recordeduse.Decimal`·`IsSHA256`·`RDFLiteral`을 import해 재사용합니다.
같은 레코드·필드·타입에서 먼저 완료된 부모를 선택하고 URI와 문장을 조립합니다.
Go 전송기는 원래 기록과 파일 해시·프레임·타입을29개 원시 필드에 담습니다.
기존26개 필드 조회와32개 필드·지역변수 조회의 입력 계약은 그대로 유지합니다.

URI의 범위는 원본 JSON 파일 바이트의 SHA-256, 원본 케이스 번호,
필드 쓰기 번호입니다. 공백만 달라져도 파일 해시와 URI는 달라집니다.
소스 해시나 읽기 구간 해시가 같다는 이유로 두 파일의 값 버전을 합치지 않습니다.
Gooo는 전달된 식별자의 형태를 검사합니다. 원본 파일의 해시는 CLI가 계산하며,
파일 해시가 생산자의 진실성이나 외부 인증을 보장하지는 않습니다.

부모가 기록되지 않았거나 한도로 손실되면 관계를 만들지 않습니다.
부분 기록에서도 확인한 직접 관계는 `recording_complete`와 함께 남깁니다.
출력 관계는 `urn:gooo:field-parent:v1:retainedParent`입니다.
실행 기록의 직접 복사 관계를 나타내며, 전체 인과 경로나 PROV 추론을 주장하지 않습니다.

## 작성한 예상 결과와 비교

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/field-parent-origins/controls.json \
  examples/field-parent-origins.workspace.json
```

21개 입력에는 직접 복사, 부분 기록, 부모 부재·손실, 잘못된 해시 형태,
자기 자신·범위 밖·미래 완료·미완료 부모, 다른 레코드·필드·타입,
입력 배열 길이, 필드·지역변수 종류 혼동, 지원하지 않는 모드와 텍스트 상한이 포함됩니다.
예상 출력은 Gooo의 판정 함수와 별도로 작성했습니다.
이 숫자는 이 유한 사례들의 비교 개수입니다. 언어 전체의 완전성이나 모델 정답률로
사용하지 않습니다. `controls.json`은 도구가 읽는 입력·예상 출력 자료입니다.

에이전트는 JSON의 `decision`, `finite_passed`, `finite_total`을 함께 확인합니다.
부분 충족은 `PROGRESS`로 표현되며 명령 종료 코드는0일 수 있습니다.
