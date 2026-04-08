# PDF Splitter

실제 페이지 썸네일을 보면서 페이지 사이 구분선을 직접 눌러 PDF를 분할하는 Windows/macOS용 GUI 앱입니다.

## 주요 기능

- PDF 파일 선택 및 경로 직접 입력
- Windows 파일 드래그앤드롭 지원
- 실제 PDF 페이지 썸네일 미리보기
- 리스트형 / 격자형 페이지 보기 전환
- 페이지 클릭 확대 보기
- 페이지 사이 구분선 클릭 기반 수동 분할
- 빠른 분할 단위 토글 + 숫자 입력 + 추천 단위 드롭다운
- 원본 폴더 차번 파일명 저장
- 분할 완료 후 저장 위치 열기
- 시스템 / 라이트 / 다크 테마 전환
- 커스텀 타이틀바 및 앱 아이콘

## 기술 스택

- Go 1.26+
- Gio
- pdfcpu
- Apache PDFBox 3.0.7

## 보안 하드닝

- 입력 경로 정규화 및 절대경로 확인
- `.pdf` 확장자, 실제 파일, 빈 파일 여부 확인
- PDF 헤더 `%PDF-` 시그니처 확인
- 파일 크기 상한: 256MB
- 페이지 수 상한: 2000페이지
- `pdfcpu` 유효성 검사
- 암호화 PDF 차단
- 첨부파일 포함 PDF 차단
- `OpenAction`, `AA`, `JavaScript`, `Launch`, `SubmitForm`, `ImportData`, `GoToR`, `GoToE`, `XFA`, `RichMedia` 등 위험 기능 차단
- 분할 저장 시 임시 파일 후 확정 방식 적용
- 썸네일 렌더링 타임아웃 및 Java 힙 상한 적용
- 썸네일 캐시 자동 정리
- 저장 위치 열기 시 셸 문자열 조합 없이 디렉터리 실존 여부 확인 후 직접 실행

## 요구 사항

- Go 1.26+
- Java Runtime

`PDFBox` JAR은 저장소의 `tools/pdfbox`에 포함되어 있고, Windows 빌드 시 `dist/windows/tools/pdfbox`로 같이 복사됩니다.

## 실행

```powershell
go run ./cmd/pdfsplitter
```

## 빌드

Windows:

```powershell
./scripts/build.ps1
```

macOS:

```bash
chmod +x ./scripts/build-macos.sh
./scripts/build-macos.sh
```

생성 경로:

- Windows: `dist/windows/pdf-splitter.exe`
- macOS: `dist/macos/pdf-splitter`

## 릴리즈 패키징

초기 릴리즈 패키지는 Windows 기준으로 아래 스크립트로 생성합니다.

```powershell
./scripts/release.ps1
```

생성 경로:

- ZIP: `dist/release/v0.1.0/pdf-splitter-v0.1.0-windows-x64.zip`
- 체크섬: `dist/release/v0.1.0/SHA256SUMS.txt`
- 릴리즈 노트: `dist/release/v0.1.0/RELEASE_NOTES.md`

## 점검

```powershell
./scripts/security-check.ps1
```

`govulncheck`가 설치되어 있으면 같이 실행되고, 없으면 빌드/테스트와 PDFBox 자산 확인까지만 수행합니다.

## 구조

```text
cmd/pdfsplitter      앱 진입점
internal/pdf         PDF 분석, 분할, 썸네일, 보안 정책
internal/platform    OS별 파일 선택, 폴더 열기, 테마 감지
internal/ui          GUI 화면과 사용자 상호작용
scripts              빌드 및 보안 점검 스크립트
tools/pdfbox         썸네일 렌더러 자산
```

## 남은 작업

- macOS 실기기 빌드 및 실행 검증
- macOS 드래그앤드롭
- 설정 저장
- 최근 작업 이력
- 배포 서명
