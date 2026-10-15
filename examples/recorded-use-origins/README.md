# 만든 값 버전을 어떤 호출이 읽었는가

개발 소스용 메타프로그램입니다. [Inspect](inspect.gooo)는 저장된 지역 변수
읽기 한 건과 그 읽기가 참조한 대입을 받아, 생성 관계와 사용 관계를 함께 반환합니다.
조건과 결과의 의미는 Gooo에 있으며 Go는 범위가 정해진 기본 타입의 열을 전달합니다.

| 구분 | 같은 저장 문서 안의 좌표 |
| --- | --- |
| 값 버전 | `case_index + assignment_index` |
| 호출 | `case_index + frame` |
| 읽기 사건 | `case_index + use_index` |
| 문서 범위 | 원문 JSON의 `receipt_sha256` |

원래 번호와 문서 해시로 버전을 구분합니다. 생산자 프로그램의 선택 소스와
지역 기록의 해시도 함께 반환합니다. 메타프로그램 자신의 소스 해시는 그 실행의
runtime에 별도로 남습니다.

## 작은 기록 읽기

저장소 루트에서 실행합니다. 앞선 [값 버전 예제](../local-version-origins/README.md)의
압축된 기록을 그대로 사용합니다.

```sh
saved=$(mktemp)
gzip -dc examples/local-version-origins/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-shape local-use-columns \
  --origin-offset 15 \
  examples/recorded-use-origins/gooo.workspace.json
rm "$saved"
```

이 형태의 기본 페이지 크기는 1입니다. 위 조회는 결과가 0인 입력에서 `total`을
읽은 기록입니다. 원래 `case_index=1`, `use_index=7`, `frame=0`을 보존하고,
참조한 `assignment_index=7`이 완료됐는지 함께 읽습니다. 원래 생산자를 다시
실행하지 않으며 모델 호출도 없습니다. 전체 보관 읽기 32건 중 한 건의 조회입니다.

## Gooo가 반환하는 관계

| 관계 | 주체 → 대상 |
| --- | --- |
| `prov:wasGeneratedBy` | 값 버전 → 그 버전을 만든 호출 |
| `prov:used` | 읽은 호출 → 해당 값 버전 |

관계 IRI는 [W3C PROV-O의 생성·사용 어휘](https://www.w3.org/TR/prov-o/#used)를 씁니다.
현재 지역 읽기 계약은 같은 호출 프레임의 마지막 완료 대입을 참조합니다.
Gooo는 완료 여부와 프레임·활동 일치도 다시 확인해 두 관계를 구성합니다.
값의 이동을 보려면 이 두 방향을 함께 읽으면 됩니다.

대입이 보관되지 않았다면 `UNRESOLVED_LOCAL_USE`와 원래 이유를 반환합니다.
`NO_RECORDED_ASSIGNMENT`와 `WRITE_LIMIT`를 구분하고, 이 경우 관계 IRI는 빈 값입니다.
앞선 기록 일부가 한도에 걸려도 지금 보관된 정상 읽기 관계는 유지하면서
`local_records_dropped`를 표시합니다. 모순된 입력에는 `INVALID_INPUT`을 반환합니다.

`used`와 생성 관계는 여기서 관찰한 두 관계입니다. 반환값의 파생은 별도 연결을
필요로 합니다. [wasDerivedFrom](https://www.w3.org/TR/prov-o/#wasDerivedFrom)은 엔티티를
바탕으로 다른 엔티티를 만드는 관계이므로, 이 조회에서는 파생 관계를 덧붙이지 않습니다.

## 문서의 기원과 실행 회차

`receipt_sha256`은 **압축을 푼 원문 JSON 바이트**를 식별합니다. 줄바꿈이나 표기가
달라지면 해시가 달라집니다. 같은 문서의 복사본은 같은 해시를 가집니다. 실행 회차의
전역 UUID는 후속 범위이며, 이 필드를 새 실행의 식별자로 해석하지 않습니다.

결과의 `receipt_scope=saved_json_bytes`와 사례·대입 번호를 함께 사용하면 읽기와
참조 버전을 같은 문서 안에서 연결할 수 있습니다. `origin_input.receipt_sha256`에도
같은 해시가 남습니다. 이 뷰의 범위는 저장된 기록의 주장과 그 원문입니다.

`read_sha256`, `read_start`, `read_end`는 보관된 선택 소스의 읽기 좌표입니다.
원래 파일의 좌표로 바꾸어 붙이지 않습니다. 원본 선언은 같은 기록의 지역 변수
파일 위치 조회나 [값 버전 조회](../local-version-origins/README.md)로 이어 읽을 수 있습니다.

## 현재 범위

- 지역 변수의 보관된 읽기를 다룹니다. 함수 입력 포트, 반환값 전달, 필드·별칭의
  전체 인과 연결은 각자의 추가 연결이 필요합니다.
- 전송은 최대 128행과 기존 32 KiB 입력 한도를 사용합니다. 배열은 선택한 페이지
  크기로 할당하고, 기본 조회는 한 행입니다. 이 Gooo 활동은 한 읽기를 선택합니다.
- 값 자체를 열에 복사하지 않습니다. 읽기 식별자와 참조한 버전의 메타데이터를 옮깁니다.
- 현재 출력은 타입 있는 관계 레코드입니다. RDF 내보내기와 qualified usage의 역할·시각은 후속 범위입니다.

첫 새 코호트는 검사 함수 5개를 10.17초에 통과했습니다. 통제 입력은 **12/12**,
실제 기록 조회는 원문과 줄바꿈이 다른 문서에서 같은 읽기를 두 번 관찰했습니다.
이 두 문서는 입력 해시가 달랐고 원래 사례·읽기·버전 번호는 같았습니다.
생산자 실행과 모델 호출은 0회였습니다. 시간에는 빌드와 테스트 준비가 포함됩니다.
[원본 소스와 관찰](../../docs/observations/recorded-use-origins/observations.json)에 범위를 남깁니다.

별도의 첫 조회는 [전체 생산자 영수증](full-producer-receipt.json.gz)을 입력으로
사용했습니다. offset 14에서 `case=1, use=6 → assignment=0`을 관찰했고, 문서 전체의
해시 `a15f33ca…`를 결과에 보존했습니다. 이 명령은 3.55초였으며 생산자와 모델은
실행하지 않았습니다. 추가 정적 어댑터 검사는 같은 전체 문서의 해시가 내부 runtime
해시로 바뀌지 않는지를 확인했습니다. 앞선 프로그램 조회를 로컬에서 반복하지 않았습니다.
