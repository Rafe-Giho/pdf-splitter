# CONTRIBUTING.md

## 현재 작업 브랜치

- 안정 브랜치: `master`
- macOS 작업 브랜치: `codex/macos-bringup`

`master`는 항상 배포 가능한 Windows 기준 상태를 유지합니다.  
macOS 관련 작업은 먼저 `codex/macos-bringup`에서 진행합니다.

## 브랜치 규칙

- macOS 실행/수정: `codex/macos-bringup`
- 추가 기능: `codex/<feature-name>`
- 긴급 수정: `codex/fix-<topic>`

## 커밋 규칙

- 한 커밋에는 한 가지 목적만 담습니다.
- 커밋 메시지는 짧고 명확하게 씁니다.
- 권장 형식:

```text
macos: fix file dialog on darwin
build: adjust macOS build script
ui: fix thumbnail panel layout
pdf: handle PDFBox path on macOS
docs: update macOS setup guide
```

- 커밋 전에 최소 한 가지 관련 검증을 실행합니다.

## PR 규칙

PR에는 아래가 들어가야 합니다.

1. 무엇을 바꿨는지
2. 왜 바꿨는지
3. 어떤 환경에서 검증했는지
4. 아직 남은 리스크가 무엇인지

### PR 제목 예시

```text
macos: make PDFBox thumbnail rendering work on Apple Silicon
```

### PR 본문 체크리스트

- [ ] 브랜치 목적과 직접 관련된 변경만 포함
- [ ] 불필요한 리팩터링 없음
- [ ] `go test ./...` 결과 확인
- [ ] macOS 수동 실행 결과 확인 또는 미확인 사유 기록
- [ ] 문서 업데이트 반영

## macOS 작업 순서

1. `git checkout codex/macos-bringup`
2. `git pull origin codex/macos-bringup`
3. `MACOS_SETUP.md` 확인
4. 변경
5. 최소 검증
6. 문서 업데이트
7. 커밋
8. `git push`

## Codex 세션 시작 권장 흐름

맥북에서 Codex를 열고 저장소 루트에서 아래 순서로 진행합니다.

1. `AGENTS.md` 확인
2. `PROJECT.md` 확인
3. `MACOS_SETUP.md` 확인
4. 현재 브랜치와 최근 실패 지점 확인
5. 가장 작은 macOS 관련 수정부터 진행
