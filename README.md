# grafana-assistant-agentic-demo

A bad release breaks a small web shop. Grafana Assistant notices, finds the root
cause, and its coding agent opens a fix PR. Copilot reviews it, the Assistant
addresses the comments, a workflow merges it, and Flux deploys the fix. No human
touches anything. Break to healed takes about 12 minutes.

A self-contained demo of an alert-to-fix loop with Grafana Cloud and Grafana Assistant.
See [demo/slides/architecture-page.md](demo/slides/architecture-page.md) for a Grafana Page
prompt that draws the architecture as an animated flow.

```
break.sh ── PR merged ──> GitHub Actions ──> GHCR image + deploy/ bump ──> Flux ──> Kubernetes
                                                                                      │ OTLP
Flux <── build <── autoheal-merge <── Copilot review <── fix PR <── coding agent <── Watcher + Investigation
                   (guardrails)       (@grafana loop)                                  (Grafana Cloud)
```

## Run it

Once the one-time setup is done ([guide](demo/docs/guide.md)), from the repo root:

```bash
demo/scripts/reset.sh       # clean state
demo/scripts/preflight.sh   # PREFLIGHT OK
demo/scripts/break.sh       # merge the bad release, then watch Grafana and GitHub
```

## Docs

- [demo/docs/guide.md](demo/docs/guide.md): how it works, prerequisites, setup, run, reset, troubleshooting.
- [demo/slides/page.md](demo/slides/page.md): slides as a Grafana Page prompt.
- [demo/docs/test-runs.md](demo/docs/test-runs.md): test runs with timings.

## Layout

| Path | What |
|---|---|
| `cmd/`, `internal/` | The shop in Go: `storefront` -> `checkout`, plus `loadgen`. |
| `deploy/`, `flux/` | Kubernetes manifests and the Flux sync. |
| `.github/workflows/` | `build.yaml` (test, images, release) and `autoheal-merge.yaml` (guardrails, Copilot loop, merge). |
| `AGENTS.md` | Instructions for the coding agent. |
| `demo/` | Scripts, Grafana resources, slides and docs for the demo. |
