#!/usr/bin/env bash
# Preflight for the AI Camp demo. Exit 0 only when the shop is healthy and
# its telemetry reaches your-stack.grafana.net. Usage: ./preflight.sh [max_error_ratio]
set -uo pipefail
. "$(cd "$(dirname "$0")" && pwd)/env.sh"
KCTX=$KUBE_CONTEXT; GCTX=$GCX_CONTEXT; NS=ai-camp; MAX=${1:-0.01}
fail() { echo "FAIL: $*"; exit 1; }

for d in checkout storefront loadgen; do
  ready=$(kubectl --context $KCTX -n $NS get deploy $d -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
  [ "${ready:-0}" -ge 1 ] || fail "deployment $d not ready"
done
echo "ok: pods ready ($(kubectl --context $KCTX -n $NS get deploy checkout -o jsonpath='{.spec.template.spec.containers[0].image}'))"

# Always ask gcx for JSON: its default output is a table in a normal terminal.
root=$(gcx --context $GCTX traces search '{resource.service.name="storefront"}' --since 5m --limit 3 -o json --jq '.traces[0].rootServiceName // ""' 2>/dev/null || true)
[ "$root" = '"storefront"' ] || fail "no storefront traces on otlp in the last 5 min"
echo "ok: traces arriving"

q='sum(rate(http_server_request_duration_seconds_count{service_name="checkout",http_response_status_code=~"5.."}[2m])) / sum(rate(http_server_request_duration_seconds_count{service_name="checkout"}[2m]))'
ratio=$(gcx --context $GCTX metrics query "$q" -o json --jq '.data.result[0].value[1] // "0"' 2>/dev/null | tr -d '"' || true)
[ -n "$ratio" ] || fail "metrics query failed on $GCTX"
echo "checkout 5xx ratio: $ratio (max $MAX)"
python3 -I -c "import sys;sys.exit(0 if float('$ratio')<=float('$MAX') else 1)" || fail "checkout 5xx ratio $ratio above $MAX"
echo "PREFLIGHT OK"
