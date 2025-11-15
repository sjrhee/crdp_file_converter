# 기여 가이드

이 프로젝트에 기여해주셔서 감사합니다! 다음 가이드라인을 따라주세요.

## 시작하기

### 1. 저장소 포크 및 클론

```bash
git clone https://github.com/YOUR_USERNAME/crdp_file_converter.git
cd crdp_file_converter
```

### 2. 의존성 설치

```bash
go mod download
```

### 3. 브랜치 생성

```bash
git checkout -b feature/your-feature-name
```

## 코드 스타일

### Go 코드 포맷팅

```bash
# 자동 포맷팅
make fmt

# 린트 검사
make lint
```

### 명명 규칙

- **함수/메서드**: camelCase (예: `processFile`, `validateInput`)
- **상수**: UPPER_SNAKE_CASE (예: `MAX_BATCH_SIZE`)
- **패키지**: 소문자, 단어조합 (예: `converter`, `crdp`)

### 코드 주석

- 공개 함수/타입에는 Godoc 주석 추가
- 복잡한 로직에는 인라인 주석 추가

```go
// ProcessFile converts CSV/TSV files by encoding/decoding specific columns.
// It handles header detection, batch processing, and error management.
func (dc *DumpConverter) ProcessFile(...) error {
    // Implementation
}
```

## 테스트

### 테스트 작성

```bash
# 테스트 실행
make test

# 커버리지 포함 테스트
make test-cov
```

### 테스트 요구사항

- 새로운 기능에 대한 유닛 테스트 필수
- 커버리지 80% 이상 목표
- 테스트 파일은 `*_test.go` 형식

### 테스트 예시

```go
func TestProcessFile(t *testing.T) {
    // Setup
    converter := converter.NewDumpConverter(host, port, policy, timeout)
    
    // Test
    err := converter.ProcessFile(inputFile, outputFile, delimiter, columnIndex, operation, skipHeader, batchSize)
    
    // Assert
    if err != nil {
        t.Fatalf("ProcessFile failed: %v", err)
    }
}
```

## 커밋 메시지

### 형식

```
<type>: <subject>

<body>

<footer>
```

### Type

- **feat**: 새로운 기능
- **fix**: 버그 수정
- **docs**: 문서 변경
- **style**: 코드 스타일 (포매팅, 세미콜론 등)
- **refactor**: 코드 리팩토링
- **perf**: 성능 개선
- **test**: 테스트 추가/수정
- **chore**: 빌드, 의존성 등

### 예시

```
feat: add parallel file processing capability

- Implement SplitInputFile() to divide input file into chunks
- Add ProcessFileParallel() for concurrent processing
- Support configurable number of parallel workers
- Include --parallel flag in CLI

Closes #42
```

## Pull Request

### PR 전 체크리스트

- [ ] 코드가 `make fmt`를 통과하는가?
- [ ] 모든 테스트가 통과하는가? (`make test`)
- [ ] 새 테스트를 추가했는가?
- [ ] README를 업데이트했는가? (필요시)
- [ ] CHANGELOG.md를 업데이트했는가?
- [ ] 커밋 메시지가 명확한가?

### PR 설명 템플릿

```markdown
## 설명
이 PR이 하는 일에 대한 설명

## 연관 이슈
Closes #123

## 변경 유형
- [ ] 새로운 기능
- [ ] 버그 수정
- [ ] 문서 변경
- [ ] 성능 개선

## 테스트
테스트 방법 설명

## 스크린샷 (필요시)
```

## 문제 제출 (Issues)

### 버그 보고

```markdown
## 버그 설명
버그가 무엇인지 명확하게 설명

## 재현 방법
1. 이 명령을 실행...
2. 이 옵션을 사용...
3. 결과 보기...

## 예상 동작
무엇이 일어나야 하는가

## 실제 동작
실제로 일어난 일

## 환경
- OS: Ubuntu 20.04
- Go 버전: 1.21
- CRDP 서버: X.X.X
```

### 기능 제안

```markdown
## 제안 설명
새로운 기능에 대한 설명

## 사용 사례
이 기능이 필요한 이유

## 가능한 해결책
어떻게 구현할 수 있을까?

## 추가 정보
```

## 개발 워크플로우

### 로컬 환경 설정

```bash
# 저장소 클론
git clone https://github.com/sjrhee/crdp_file_converter.git

# 의존성 설치
make install

# 빌드
make build

# 테스트
make test
```

### 개발 중

```bash
# 코드 포맷팅
make fmt

# 린트 검사
make lint

# 크로스플랫폼 빌드 (옵션)
make build-cross
```

### PR 제출 전

```bash
# 모든 테스트 통과 확인
make test

# 커버리지 확인
make test-cov

# 최종 빌드 확인
make build
```

## 프로젝트 구조 이해

```
crdp-file-converter/
├── cmd/               # CLI 엔트리 포인트
├── pkg/
│   ├── crdp/         # CRDP API 클라이언트
│   └── converter/    # 파일 변환 로직
├── testdata/         # 테스트 데이터
├── docs/             # 문서
├── .github/workflows # CI/CD 파이프라인
└── Makefile          # 빌드 스크립트
```

## 질문?

GitHub Issues를 통해 질문해주세요. 더 빠른 응답이 필요하면 이메일로 연락해주세요.

감사합니다! 🎉
