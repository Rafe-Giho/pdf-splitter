# AGENTS.md

## 목적

이 저장소는 `Go + Gio + pdfcpu + PDFBox` 기반의 GUI PDF 분할기입니다.

현재 작업 브랜치 `codex/macos-bringup`의 1차 목표는 아래입니다.

- macOS에서 앱이 실제로 실행되게 만들기
- 파일 선택, PDF 분석, 썸네일 생성, 분할 저장이 macOS에서 동작하는지 검증하기
- `.app` 번들링 전 단계까지 필요한 문제를 정리하고 수정하기

## 작업 원칙

- 항상 현재 브랜치 목적과 직접 관련된 변경만 한다.
- Windows 안정 동작을 깨는 변경은 피한다.
- 관련 없는 리팩터링은 하지 않는다.
- macOS 이슈를 고치되, 공용 코드 수정이 필요하면 가장 작은 범위로 한다.
- 빌드 스크립트, 플랫폼 코드, 문서를 우선적으로 점검한다.

## 우선순위

1. macOS에서 `go run ./cmd/pdfsplitter` 실행
2. macOS 파일 선택 동작 확인
3. PDF 분석 동작 확인
4. 썸네일 렌더링 동작 확인
5. PDF 분할 저장 동작 확인
6. `scripts/build-macos.sh` 빌드 산출물 검증
7. 이후 `.app` 번들링, 서명, 노타리제이션 준비

## 수정 전 확인

- 먼저 `MACOS_SETUP.md`와 `PROJECT.md`를 읽는다.
- macOS 관련 수정이면 `internal/platform`, `scripts/build-macos.sh`, `internal/ui`, `internal/pdf`를 우선 점검한다.
- 의존성 추가, 버전 변경, 릴리즈 정책 변경은 꼭 필요할 때만 한다.

## 검증 원칙

- 가장 작은 관련 검증만 실행한다.
- 기본 검증 순서:
  - `go test ./...`
  - `go build ./cmd/pdfsplitter`
  - macOS에서는 필요 시 `./scripts/build-macos.sh`
- GUI는 가능하면 실제 수동 확인 결과를 문서에 남긴다.

## 문서 반영

아래 파일은 작업하면서 같이 최신 상태로 유지한다.

- `PROJECT.md`: 진행 상태와 남은 과제
- `MACOS_SETUP.md`: 맥북 세팅과 실행 절차
- `CONTRIBUTING.md`: 브랜치/커밋/PR 규칙

## 금지 사항

- 근거 없이 macOS 완료로 표기하지 않는다.
- 검증하지 않은 기능을 배포 완료처럼 문서화하지 않는다.
- 브랜치 목적과 무관한 대규모 UI 개편을 하지 않는다.
