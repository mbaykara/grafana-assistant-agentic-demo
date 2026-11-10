# Setup and run guide

Everything needed to deploy, run and reset the demo. Short overview: the root `README.md`.

## How the pieces fit

- **App.** `storefront` forwards each cart to `checkout`, which prices it. `loadgen`
  sends about 5 carts per second. All three send traces, metrics and logs over OTLP
  (`OTEL_*` variables in `deploy/`; metric export every 15s).
- **The bug.** `break.sh` reverts the last `fix(pricing)` commit in a PR titled
  "refactor(pricing): allocate receipt map lazily" and merges it. An empty cart with
  code `CAMP10` then hits `panic: assignment to entry in nil map`, about 20% of
  checkouts fail with HTTP 500.
- **Delivery.** Every push to main builds images tagged `sha-<7>`, and the release job
  commits the new tag to `deploy/kustomization.yaml`. Flux deploys it within about
  2 minutes. Changes only under `demo/` do not build.
- **Detection.** The Watcher `customer-checkout-journey` checks the shop (every 1h by default,
  5m while demoing) and picks up the errors on its next scan after the bad image is live. When unhealthy it starts an
  Investigation, which hands a code fix to the coding agent in a sandbox. The agent
  opens a PR on an `assistant/*` branch.
- **Review and merge.** `autoheal-merge.yaml` runs only for PRs by
  `grafana-assistant[bot]` on `assistant/*` that change only `internal/`. It runs
  vet and tests, marks the PR ready, asks Copilot for a review and waits up to
  5 minutes. Comments: it posts "@grafana please address the Copilot review comments"
  (max 2 rounds), the Assistant pushes a fix, Copilot reviews again. Clean review:
  squash-merge and build. Anything else is left for a person.

## Prerequisites

- A Grafana Cloud stack with Grafana Assistant (Watchers, Investigations, coding agents).
- A Kubernetes cluster sending to that stack with the
  [k8s-monitoring Helm chart](https://github.com/grafana/k8s-monitoring-helm)
  (tested with 4.6.0, Application Observability enabled). The apps send OTLP to
  `http://grafana-cloud-otel-receiver.monitoring.svc:4318`; change it in `deploy/*.yaml`
  if your receiver differs.
- A GitHub account with a Copilot plan that includes code review.
- CLI tools: `git`, `gh`, `kubectl`, `flux` (2.7+), `gcx`, `go` (1.27+), `python3`.

## One-time setup

The values below are the ones this repo runs with. If you use your own fork, cluster
or stack, replace `mbaykara/grafana-assistant-agentic-demo` (`git grep -l mbaykara`
lists every file, including the image names) and set the script variables:

```bash
export KUBE_CONTEXT=my-cluster GCX_CONTEXT=otlp DEMO_REPO=mbaykara/grafana-assistant-agentic-demo
```

Defaults for these live in `demo/scripts/env.sh`.

### 1. Repository

```bash
gh repo clone mbaykara/grafana-assistant-agentic-demo && cd grafana-assistant-agentic-demo
gh label create autoheal --description "Fix PR opened by Grafana Assistant" || true
pre-commit install     # secret checks on commit
```

In the repo settings:

- Turn on Copilot automatic code review for pull requests (Settings -> Rules ->
  Rulesets, or Settings -> Copilot -> Code review).
- Create a fine-grained personal access token: this repository only,
  Pull requests: Read and write. Store it only as a repo secret:

  ```bash
  gh secret set AUTOHEAL_USER_TOKEN    # paste the token when asked
  ```

  It must be a user token: Copilot ignores review requests from `GITHUB_TOKEN`, and
  the Assistant ignores @grafana mentions from bots.

### 2. Images

Push to main (or run `gh workflow run build.yaml -f ref=main`) and wait for the
`build` workflow. It pushes `ghcr.io/<owner>/ai-camp-{checkout,storefront,loadgen}`
and commits the tag to `deploy/kustomization.yaml`.

### 3. Cluster

```bash
kubectl --context "$KUBE_CONTEXT" create namespace ai-camp
# Pull secret for the private GHCR images: classic PAT with read:packages only.
kubectl --context "$KUBE_CONTEXT" -n ai-camp create secret docker-registry ghcr \
  --docker-server=ghcr.io --docker-username=<github-user> --docker-password=<token>

flux install --context "$KUBE_CONTEXT" --components=source-controller,kustomize-controller
# Read-only deploy key for Flux. The command prints the public key.
flux create secret git ai-camp-repo -n flux-system --context "$KUBE_CONTEXT" \
  --url=ssh://git@github.com/$DEMO_REPO --ssh-key-algorithm=ed25519
gh repo deploy-key add <public-key-file> --title flux-ai-camp   # read-only by default
kubectl --context "$KUBE_CONTEXT" apply -f flux/ai-camp.yaml
flux --context "$KUBE_CONTEXT" get kustomizations ai-camp        # Ready=True
```

### 4. Grafana

```bash
gcx --context "$GCX_CONTEXT" api /api/folders -d '{"uid":"ai-camp","title":"AI Camp"}'
gcx --context "$GCX_CONTEXT" dashboards create -f demo/grafana/dashboard-ai-camp-shop.json \
  --api-version dashboard.grafana.app/v1beta1
gcx --context "$GCX_CONTEXT" api /api/v1/provisioning/alert-rules \
  -H 'X-Disable-Provenance: true' -d @demo/grafana/alert-checkout-errors.json
grep -v watcher-id demo/grafana/watcher-customer-checkout-journey.yaml \
  | gcx --context "$GCX_CONTEXT" assistant watchers create -f -
gcx --context "$GCX_CONTEXT" assistant watchers start customer-checkout-journey
```

The Watcher prompt names the repo URL; edit it in the YAML first if you use a fork.

In the Grafana UI:

- Install the Grafana Assistant GitHub App on this repository only and link your
  GitHub account.
- Add the model token the coding agent uses (Assistant settings -> Coding agents -> Secrets).
- Slides: paste `demo/slides/page.md` into the Assistant chat. It creates and publishes
  the Page. Pages cannot be created with gcx.

### 5. Check

```bash
demo/scripts/preflight.sh      # prints PREFLIGHT OK
```

It checks that the pods are ready, storefront traces arrive, and the checkout 5xx
ratio is below 1%.

## Run the demo

First set the Watcher interval to 5m in Grafana (AI -> Watchers -> Customer checkout journey);
`break.sh` refuses to run otherwise. Then, from the repo root:

```bash
demo/scripts/reset.sh                 # clean state: no open PRs, main healthy
demo/scripts/preflight.sh             # PREFLIGHT OK
demo/scripts/break.sh                 # minute 0: merge the bad release
```

Then watch, in this order:

| ~Minute | Where | What happens |
|---|---|---|
| 0-3 | GitHub Actions, dashboard "AI Camp shop" | Build, release commit, Flux deploys the bad image; 5xx rises to about 20% |
| 3-10 | AI -> Watchers -> Customer checkout journey | Next scheduled scan (every 5 min) finds the errors, starts an Investigation |
| 6-9 | AI -> Investigations | Root cause: nil map in `internal/pricing`, storefront errors are a symptom |
| 8-11 | GitHub pull requests | Assistant PR with fix and test; Copilot review; @grafana loop if it comments |
| 11-13 | Actions, dashboard | Auto-merge, build, Flux deploys the fix; 5xx back to 0, new version label |


## Reset and clean up

```bash
demo/scripts/reset.sh                 # closes open PRs, deletes assistant/* branches,
                                      # merges a reset PR if main is still broken
```

Set the Watcher interval back to 1h in Grafana. When no more demos are planned, revoke `AUTOHEAL_USER_TOKEN` in GitHub
(Settings -> Developer settings -> Personal access tokens).

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `ImagePullBackOff` | Pull secret `ghcr` missing or expired, or the tag was not pushed. Check the `build` run for that commit. |
| `PREFLIGHT` fails on traces | k8s-monitoring receiver not reachable or sending to another stack. Check `OTEL_EXPORTER_OTLP_ENDPOINT` and the chart's destination. |
| No Watcher scan | `gcx --context "$GCX_CONTEXT" assistant watchers run customer-checkout-journey`; check `watchers status customer-checkout-journey`. |
| No PR from the Investigation | Check the GitHub App is installed on the repo and the coding agent token is set. Fallback: push `demo/fallback/fix-850ab52.patch` on an `assistant/` branch, or show the Investigation's proposed diff. |
| Copilot does not review | Copilot skips draft PRs (the workflow marks it ready) and needs a request from a user with a Copilot seat (`AUTOHEAL_USER_TOKEN`). The workflow hands the PR to a human after 5 minutes. |
| Assistant ignores the @grafana comment | The comment must come from a user token, not `github-actions[bot]`. |
| PR "out of date" and not merging | `gh pr update-branch <number>`; the workflow runs again. |
| Flux slow | `flux --context "$KUBE_CONTEXT" reconcile kustomization ai-camp --with-source` |
| Rate panels empty right after start | Metrics export every 15s; wait 1-2 minutes for `rate([1m])` to have data. |
