#!/usr/bin/env bash
# GitOps "bad release": merge a PR that brings back the nil map bug in pricing.
# CI builds the image and bumps deploy/; Flux rolls it out (~3-5 min).
set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/env.sh"
REPO=$DEMO_REPO
SCRIPTS=$(cd "$(dirname "$0")" && pwd)
DIR=$(cd "$SCRIPTS/../.." && pwd)
cd "$DIR"
# The Watcher must scan every 5 min, or nobody notices the bug for up to an hour.
iv=$(gcx --context "$GCX_CONTEXT" assistant watchers get "$WATCHER" -o yaml 2>/dev/null | awk '/^  interval:/{print $2}' || true)
if [ "$iv" != "5m0s" ] && [ -z "${ALLOW_ANY_INTERVAL:-}" ]; then
  echo "Watcher interval is ${iv:-unknown}, not 5m. Set it to 5m in Grafana (AI -> Watchers -> Customer checkout journey) first, or set ALLOW_ANY_INTERVAL=1."
  exit 1
fi
git checkout -q main && git pull -q --ff-only
FIX=$(git log --format=%H -n1 --grep='^fix(pricing)' -- internal/pricing/pricing.go)
[ -n "$FIX" ] || { echo "no fix(pricing) commit to undo: is main already broken?"; exit 1; }
BR="refactor/pricing-lazy-receipt-$(date +%H%M%S)"
git checkout -q -b "$BR"
git revert --no-edit --no-commit "$FIX"
git commit -q -m "refactor(pricing): allocate receipt map lazily"
git push -q -u origin "$BR"
gh pr create -R $REPO --head "$BR" --title "refactor(pricing): allocate receipt map lazily" \
  --body "Only allocate the receipt map when the cart has items."
gh pr merge -R $REPO "$BR" --merge --delete-branch
git checkout -q main && git pull -q --ff-only && git branch -q -D "$BR" 2>/dev/null || true
echo "bad release merged at $(date -u +%H:%M:%SZ). Flux deploys in ~2 min; the Watcher picks it up on its next scheduled scan."
