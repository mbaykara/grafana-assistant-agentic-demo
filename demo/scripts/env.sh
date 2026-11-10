# Shared settings for the demo scripts. Override any of them in your shell, e.g.
#   export KUBE_CONTEXT=my-cluster GCX_CONTEXT=my-stack DEMO_REPO=me/my-fork
KUBE_CONTEXT=${KUBE_CONTEXT:-my-cluster}
GCX_CONTEXT=${GCX_CONTEXT:-my-stack}
DEMO_REPO=${DEMO_REPO:-mbaykara/grafana-assistant-agentic-demo}
WATCHER=${WATCHER:-customer-checkout-journey}
