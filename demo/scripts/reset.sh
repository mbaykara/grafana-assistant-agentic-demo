#!/usr/bin/env bash
# After a run: close open fix PRs, delete assistant/* branches.
# If main is still broken (fix not merged), undo the bad release via a PR.
set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/env.sh"
REPO=$DEMO_REPO
DIR=$(cd "$(dirname "$0")/../.." && pwd)
for pr in $(gh pr list -R $REPO --state open --json number --jq '.[].number'); do
  gh pr close -R $REPO "$pr" --delete-branch --comment "Closed by reset.sh after a demo run."
done
for b in $(git ls-remote --heads "https://github.com/$REPO" 'assistant/*' | awk '{print $2}' | sed 's#refs/heads/##'); do
  git -C "$DIR" push -q origin --delete "$b" || true
done
cd "$DIR" && git checkout -q main && git pull -q --ff-only
last_pricing=$(git log --format=%s -n1 -- internal/pricing/pricing.go)
if [[ "$last_pricing" == refactor\(pricing\)* ]]; then
  BAD=$(git log --format=%H -n1 -- internal/pricing/pricing.go)
  BR="fix/pricing-reset-$(date +%H%M%S)"
  git checkout -q -b "$BR" && git revert --no-edit "$BAD" >/dev/null
  git commit -q --amend -m "fix(pricing): restore eager receipt map (demo reset)"
  git push -q -u origin "$BR"
  gh pr create -R $REPO --head "$BR" --title "fix(pricing): restore eager receipt map (demo reset)" --body "Demo reset."
  gh pr merge -R $REPO "$BR" --merge --delete-branch
  git checkout -q main && git pull -q --ff-only && git branch -q -D "$BR" 2>/dev/null || true
  echo "main was broken: reset PR merged. Flux deploys the healthy build in ~3-5 min."
else
  echo "main already healthy ($last_pricing)."
fi
