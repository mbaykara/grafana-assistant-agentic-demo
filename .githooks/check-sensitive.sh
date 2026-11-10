#!/usr/bin/env bash
# Pre-commit check: block customer data and internal links in staged changes.
#
# Generic rules live here. Customer-specific terms (people, domains, stack
# names) go in .git/info/sensitive-terms, one case-insensitive extended regex
# per line, '#' for comments. That file is never committed, so the list itself
# does not leak. Secrets are covered by the gitleaks hook.
set -euo pipefail

root=$(git rev-parse --show-toplevel)
terms_file="$root/.git/info/sensitive-terms"
self=".githooks/check-sensitive.sh"
fail=0

# File names that must never be committed.
bad_paths='(^|/)(call-notes|regular-calls)/|transcript|\.vtt$|\.eml$|\.msg$|\.xlsx$|\.docx$|\.pptx$|\.pdf$|kubeconfig|\.env($|\.)|\.tfvars$|\.pem$|\.key$|\.p12$'
while IFS= read -r f; do
  if printf '%s\n' "$f" | grep -qiE "$bad_paths"; then
    echo "blocked file: $f (customer document, transcript or credential file)"
    fail=1
  fi
done < <(git diff --cached --name-only --diff-filter=ACMR)

# Generic content rules: personal emails and links to internal tools.
generic=(
  '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}'
  '[a-z0-9-]+\.slack\.com/'
  '(docs|drive|sheets|slides)\.google\.com/'
  'rocketlane\.com'
  'zoom\.us/'
  'gong\.io'
  'salesforce\.com|force\.com/'
  'atlassian\.net'
)
allow='git@github\.com|noreply|users\.noreply\.github\.com|@example\.(com|org)|@grafana|@copilot'

patterns=("${generic[@]}")
if [ -f "$terms_file" ]; then
  while IFS= read -r t; do
    t="${t%%#*}"; t="$(printf '%s' "$t" | sed 's/[[:space:]]*$//')"
    [ -n "$t" ] && patterns+=("$t")
  done < "$terms_file"
else
  echo "note: $terms_file not found, checking generic rules only"
fi

# Only added lines, per file, excluding this script.
while IFS= read -r f; do
  [ "$f" = "$self" ] && continue
  added=$(git diff --cached -U0 --no-color -- "$f" | grep -E '^\+' | grep -vE '^\+\+\+ ' || true)
  [ -z "$added" ] && continue
  for p in "${patterns[@]}"; do
    hits=$(printf '%s\n' "$added" | grep -inE -- "$p" | grep -viE "$allow" || true)
    if [ -n "$hits" ]; then
      echo "blocked: $f matches /$p/:"
      printf '%s\n' "$hits" | cut -c1-160 | sed 's/^/    /'
      fail=1
    fi
  done
done < <(git diff --cached --name-only --diff-filter=ACMR)

if [ "$fail" -ne 0 ]; then
  echo
  echo "Commit blocked: remove customer data or internal links above."
  echo "If a match is a false positive, reword it. Bypass only after review: SKIP=customer-data git commit ..."
  exit 1
fi
