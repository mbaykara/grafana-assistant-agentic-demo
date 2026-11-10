# Agent notes

- Go is not preinstalled in the sandbox. The sandbox is Alpine on arm64: install with
  `curl -sSfL https://go.dev/dl/go1.27.3.linux-arm64.tar.gz | tar xz -C /workspace && export PATH=/workspace/go/bin:$PATH`.
- Language: Go. Build: `go build ./...`. Test: `go test ./...`. Lint: `go vet ./...`.
- Business logic lives in `internal/pricing`. HTTP handlers in `cmd/<service>/main.go`.
- A fix must come with a test in the package it touches that fails before the fix.
- Keep changes minimal. Do not touch `internal/telemetry` or `deploy/` unless the bug is there. `deploy/kustomization.yaml` is updated by CI only.
- Commit messages follow Conventional Commits (`fix(pricing): ...`).
- Open pull requests against `main` with the label `autoheal`.
- Always push to a new branch named `assistant/fix-<topic>-<YYYYMMDDHHMM>` (UTC time). Never
  reuse an existing branch name: an older PR on the same name hides the new one.
- `demo/` is demo material (slides, scripts, notes). It is not part of the
  application. Do not read or change it; diagnose from telemetry and the code.
