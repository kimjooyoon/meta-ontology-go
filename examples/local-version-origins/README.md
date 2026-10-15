# 선언 하나, 호출 여러 번, 값의 여러 버전

개발 소스용 예제입니다. `Add`를 두 번 호출해 같은 숫자를 얻어도 호출의 기원은
각각 있습니다. `total = value`로 같은 숫자를 다시 넣으면 새 갱신 기록이 생깁니다.
이 예제는 그 차이를 저장된 실행 기록에서 읽고 Gooo로 해석합니다.

| 구분 | 이 예제에서 사용하는 좌표 |
| --- | --- |
| 원래 선언 | 원본 패키지·파일·활동 ID·파일 SHA-256·선언 범위 |
| 기록된 호출 | 저장된 기록의 범위 안에서 `case_index + frame` |
| 완료된 값 버전 | 같은 범위에서 `case_index + assignment_index` |

범위에는 선택된 소스와 지역 기록의 SHA-256을 함께 보존합니다. `frame`과
`assignment_index`는 원래 번호를 유지합니다. 번호는 할당 시작 순서를 나타내므로
중첩 호출의 완료 순서를 알려면 별도 실행 관계가 필요합니다.

## 무엇을 Gooo가 하는가

[inspect.gooo](inspect.gooo)의 `Summarize`는 열의 정합성, 선언과 호출의 묶음,
완료된 갱신 수, 누락과 페이지 범위를 판단합니다. Go 어댑터는 저장된 기록에서
선택한 행을 기본 타입의 열로 전달합니다. 묶는 규칙과 결과 상태는 Gooo가 소유합니다.

`source_sha256`은 선택된 실행 소스, `file_sources_sha256`은 원본 파일 주석 묶음의
출처입니다. 선언 비교에는 각 행의 **원본** 패키지·파일·활동과 SHA-256·범위를 씁니다.
원본 주석이 없는 과거 기록은 `unmapped`로 남습니다.

## 저장한 기록 읽기

저장된 [작은 기록](recorded-runtime.json.gz)은 [producer](producer/main.gooo)를
입력 `5, -36, 5, 31`에 한 번 실행해 얻었습니다. 결과는 `41, 0, 41, 67`이고,
지역 갱신 32개와 호출 프레임 12개를 기록했습니다. 다음 명령은 저장된 기록을
읽고 새 메타프로그램을 실행합니다. 생산자 실행과 모델 호출은 각각 0회입니다.

저장소 루트에서 현재 개발 CLI를 사용합니다.

```sh
saved=$(mktemp)
gzip -dc examples/local-version-origins/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-shape local-version-columns \
  --origin-offset 8 --origin-limit 8 \
  examples/local-version-origins/gooo.workspace.json
rm "$saved"
```

이 페이지는 결과가 모두 0인 두 번째 입력의 기록입니다. 원래 선언 **4개**, 갱신이
있는 호출 **3개**, 완료된 값 버전 **8개**를 구분합니다. 전체 32행 중 8행을
읽었으므로 `page_complete=false`입니다. 마지막 페이지도 전체 집합은 아닙니다.

## 완료된 갱신과 PROV-O

`Generation`은 한 행을 받아 완료된 값 버전과 그 버전을 만든 호출을 연결하는
타입 있는 레코드를 만듭니다. [PROV-O의 wasGeneratedBy](https://www.w3.org/TR/prov-o/#wasGeneratedBy)는
엔티티가 활동으로 생성되는 관계를 나타냅니다. 여기서는 완료된 지역 값 버전과
그 갱신이 일어난 호출에 대응시킵니다.

위와 같은 방식으로 압축을 푼 기록에 다음 선택을 적용할 수 있습니다.

```sh
saved=$(mktemp)
gzip -dc examples/local-version-origins/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-shape local-version-columns \
  --origin-offset 31 --origin-limit 1 \
  examples/local-version-origins/generation.workspace.json
rm "$saved"
```

`case_index=3`, `assignment_index=7`, `frame=0`인 완료 기록에서 관계 IRI
`http://www.w3.org/ns/prov#wasGeneratedBy`를 반환합니다. 도중에 끝난 갱신은
`INCOMPLETE_WRITE`와 빈 관계를 반환합니다. 원본 위치 매핑과 전체 기록의
완전성은 별도 필드로 유지합니다.

이 레코드는 저장된 실행 주장에 대한 뷰입니다. RDF 내보내기, 전역 실행 UUID,
기록의 인증, `used`·`wasDerivedFrom`·에이전트 귀속은 후속 범위입니다.

## 결과를 해석할 때의 범위

- 호출 수는 갱신 행에 나타난 호출을 셉니다. 갱신이 없는 모든 함수 호출까지
  센 수치로 사용할 수 없습니다.
- 선언·호출이 여러 페이지에 걸칠 수 있어 페이지별 수를 단순 합산하면 중복됩니다.
- `page_complete`, `identity_recording_complete`, `source_mapping_complete`를
  각각 읽습니다. 이 세 값으로 모든 인과 의존성의 완전성을 주장할 수 없습니다.
- 전송 어댑터는 최대 128행과 32 KiB 입력 한도를 적용합니다. 이 Gooo 예제는
  최대 32행을 받아 서로 비교합니다. 비교 비용은 행 수의 제곱에 비례합니다.
- 값의 실제 내용은 열에 복사하지 않습니다. 동일한 숫자의 동등성으로 기원을
  묶는 대신 기록의 식별 좌표를 사용합니다.

## 개발 중 관찰

생산자 최초 실행은 **4/4**, 원본 파일 정책의 새 통제 입력은 **7/7**,
생성 관계의 새 완료/미완료 입력은 **2/2**를 통과했습니다. 각각의 전체 명령 시간은
6.45초, 10.00초, 6.30초였습니다. 컴파일과 테스트 준비를 포함한 로컬 관찰이므로
추론 지연이나 일반적인 성능 지표로 사용할 수 없습니다.

이전 선택 소스 기반 시제품과 현재 원본 파일 기반 구현의 관찰은 별도로
[소스·시간 기록](../../docs/observations/local-version-origins/observations.json)에
보존했습니다. 본문 줄바꿈 이후의 정확한 소스에 대한 수락은 PR의 원래 CI에서
판정합니다. 기존 실험을 성공 수치로 다시 합산하지 않습니다.
