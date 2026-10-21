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

## 보관된 읽기 비율

같은 Gooo 소스의 `MeasureRetention`은 저장 기록에서 **보관된 지역 읽기 수**와
**한도 때문에 누락된 읽기 수**를 따로 받습니다. 분모는 두 수의 합이며 결과는
만분율(`retention_basis_points`)입니다. 예를 들어 보관 32건·누락 8건이면
`8000`입니다. 이 값은 관찰된 읽기의 **기록 보관 비율**이며 프로그램의 정답률,
전체 의존성의 발견률 또는 CI 시간 지표가 아닙니다. 저장 기록이 주장한 수치를
계산하며, 기록 생산자의 실행을 새로 증명하지 않습니다. 비율 계산 규칙은
`ComputeRetention`에 있고, `MeasureRetention`이 기록의 열을 읽어 그 규칙에 전달합니다.

```sh
go run ./cmd/gooo package execute --json \
  --origin-receipt examples/recorded-use-origins/full-producer-receipt.json.gz \
  --origin-shape local-use-columns --origin-offset 15 \
  examples/recorded-use-origins/retention.workspace.json
```

이 저장 기록에서 보관 32건·누락 0건, `10000`을 관찰했습니다. 선택한 페이지는
한 행이므로 `page_complete=false`입니다. `other_origins_dropped`는 프레임이나
대입 기록의 누락을 별도로 나타냅니다. 기록의 원문 JSON 해시, 선택 소스 해시,
지역 관찰 해시는 결과에 각각 남습니다. 유효하지 않은 입력이나 정수 범위 밖의
분모는 `INVALID_INPUT`과 `-1`로 표시합니다. 독립적인 [6개 통제 입력](retention-cases.json)은
보관·누락·다른 종류의 누락·잘못된 페이지·오버플로와 큰 누락 수를 구분합니다. 이 실행은
원래 생산자를 다시 실행하거나 모델을 호출하지 않습니다. 현재 활동은 읽기 한 건이
선택된 페이지를 요구하므로 보관된 읽기가 0건인 기록은 별도로 다뤄야 합니다.

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
- `Inspect`는 타입 있는 관계 레코드를 반환합니다. 아래 `ExportNTriples`는 같은
  관찰을 RDF로 옮깁니다. 역할과 시각은 해당 기록을 연결하는 후속 범위입니다.

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

## 짧은 설명이 필요할 때

같은 소스의 `Explain`은 `Inspect`를 한 번 호출하고 그 판정을 한국어로 설명합니다.
생성과 사용이 연결된 경우, 대입이 보관되지 않은 경우, 기록 한도와 모순된 입력을
구분합니다. 소스 위치가 미확정이거나 다른 기록이 한도에 걸린 경우도 설명에 남깁니다.

```sh
saved=$(mktemp)
gzip -dc examples/recorded-use-origins/full-producer-receipt.json.gz > "$saved"
go run ./cmd/gooo package execute \
  --origin-receipt "$saved" --origin-shape local-use-columns \
  --origin-offset 0 \
  examples/recorded-use-origins/explanation.workspace.json
rm "$saved"
```

기본 출력은 입력 기원과 선택한 결과를 보여 줍니다. `--json`을 더하면 전체 실행 기록을
받습니다. 첫 저장 기록 조회에서는 다음 설명을 관찰했습니다.

> 호출이 'value'의 기록된 값 버전을 읽었습니다. 같은 기록에서 이 버전의 대입 완료를 확인했습니다.

이 조회의 원래 좌표는 `case=0, use=0, frame=1, assignment=1`입니다. 원문 JSON의
해시는 `a15f33ca…`로 보존됐습니다. 변수 이름은 128바이트까지 설명에 표시하며,
더 긴 이름에는 일반적인 문장을 사용합니다. 구조화된 값 버전과 관계 조회는 `Inspect`로
계속 읽을 수 있습니다.

첫 설명 프로그램은 통제 사례 **10/10**을 1.75초에 통과했고, 저장된 읽기 한 건의
별도 조회는 1.55초에 `OBSERVED`로 끝났습니다. 생산자 실행과 모델 호출은 모두 0회입니다.
이 값은 각 원본 명령의 준비·빌드를 포함합니다. 소스 바이트는 공개 예제와 같으며,
첫 실행의 원래 경로·입력·컴파일러와 현재 소스의 CI 검증은 각자의 범위를 가집니다.
[첫 실행과 소스 일관성 기록](../../docs/observations/recorded-use-explanation/observations.json)에
소스 해시와 분모를 보존했습니다.

## 읽기 사건을 RDF로 내보내기

개발 소스의 `ExportNTriples`는 Gooo 조건식으로 기록을 확인하고, 호출·값 버전·읽기의
이름과 관계를 구성합니다. 목록에 담긴 각 줄은 [N-Triples](https://www.w3.org/TR/n-triples/)
한 문장입니다. 다음 명령은 저장된 기록 한 건을 받아 `use.nt`를 만듭니다.
`jq`는 반환된 문자열 목록을 줄로 펼칩니다.

```sh
saved=$(mktemp)
gzip -dc examples/recorded-use-origins/full-producer-receipt.json.gz > "$saved"
go run ./cmd/gooo package execute \
  --origin-receipt "$saved" --origin-shape local-use-columns --origin-offset 0 \
  examples/recorded-use-origins/rdf.workspace.json | jq -r '.lines[]' > use.nt
rm "$saved"
```

파일을 만들기 전에 상태를 확인하려면 마지막 줄의 `jq`와 리다이렉션을 생략합니다.
연결 가능한 기록은 `status=RDF_READY`와 26개 문장을 반환합니다. 대입이 없거나
입력에 모순이 있으면 판정과 이유를 유지하고 빈 `lines`를 반환합니다.

호출 하나가 같은 값을 두 번 읽어도 각 읽기 사건에는 자기 이름이 있습니다.
[PROV-O의 qualified usage](https://www.w3.org/TR/prov-o/#qualifiedUsage)를 사용해
`호출 → 읽기 사건 → 값 버전`을 표현합니다. `used`와 `wasGeneratedBy`도 함께 담아
간단한 조회와 사건별 조회를 모두 할 수 있게 합니다.

| 이름 | 같은 저장 문서 안의 구성 |
| --- | --- |
| 공통 앞부분 | `urn:gooo:saved:{receipt_sha256}:case:{case_index}` |
| 호출 | 앞부분 + `:frame:{frame}` |
| 값 버전 | 앞부분 + `:version:{assignment_index}` |
| 읽기 사건 | 앞부분 + `:use:{use_index}` |

같은 원문을 복사하면 이름도 같습니다. 원문 바이트가 바뀌면 해시와 이름이 바뀝니다.
이름의 범위는 저장 문서이며 실행 회차의 전역 식별자는 후속 과제입니다.
원래 활동의 정의 ID와 호출 ID를 구분하고, 문서·생산자 소스·지역 기록·선택 소스의
해시를 각 필드로 남깁니다. 역할·시각·반환값의 파생은 추가 관찰이 연결될 때 다룹니다.
컴파일러 핵심 어휘의 범위는 유지하고 이 예제의 Gooo 소스가 내보내기 뷰를 소유합니다.

내보낼 메타데이터는 필드당 UTF-8 128바이트까지, 각 출력 줄은 기존 Text 한도 안에서
구성합니다. 큰따옴표·역슬래시·제어 문자를 이스케이프하며 128바이트 초과에는
`EXPORT_TEXT_BOUND`, 잘못된 해시 형식에는 `UNRESOLVED_EXPORT_IDENTITY`를 반환합니다.
해시 형식 검사는 문자열 모양을 확인합니다. 저장 문서의 실제 해시는 입력 어댑터가 계산합니다.

첫 새 내보내기 사례 **12/12**, 숫자·해시·문자열 경계 사례 **32/32**를 통과했습니다.
실제 전체 저장 기록의 첫 조회는 **0.67초**였고, `case=0, frame=1, version=1, use=0`을
유지했습니다. 별도 [Go RDF 파서](https://github.com/knakk/rdf)는 저장된 출력 중 그래프
7개를 각각 26문장으로 읽고 관계 방향과 타입을 확인했습니다. 원문에 줄바꿈을 더한
새 조회에서는 원래 좌표가 유지되고 문서 해시와 세 이름이 모두 바뀌었습니다.
별도 파서로 그 그래프 한 개도 읽었습니다. 이 값들은 준비·빌드를 포함한 원본 명령의 관찰이며,
외부 작업 전체의 완전성이나 성능 분포를 나타내지는 않습니다.
[소스·첫 실패와 관찰 원본](../../docs/observations/qualified-usage/observations.json)을 함께 보관합니다.
