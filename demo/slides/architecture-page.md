# Architecture page: "How the Auto-Heal Demo Works"

A Grafana Page with an animated flow diagram, so people understand the demo before
they watch it. Pages cannot be created with gcx: paste the prompt below into the
Grafana Assistant chat (AI -> Assistant) on your-stack.grafana.net. To update the Page,
paste it into the same conversation that created it.

## Prompt

Create (or, if it already exists, replace the content of) the Grafana Page
"How the Auto-Heal Demo Works" and publish it. Category: presentation.
Tags: ai-camp, architecture. Audience: people who have not seen the demo yet. They
should understand the whole loop in under two minutes.

### Layout

1. Title and one sentence: "A bad release breaks a small web shop. Grafana Assistant
   finds the root cause, its coding agent opens a fix pull request, Copilot reviews it,
   a workflow merges it and Flux deploys it. Nobody touches anything."
2. A full-width animated flow diagram (below).
3. A "Live now" row with two small panels from the dashboard "AI Camp shop"
   (folder AI Camp): checkout request rate by status, and 5xx ratio. They show the
   loop for real while the demo runs.
4. A short numbered list of the six steps (below), each one highlighting its part of
   the diagram on hover.

### Animated flow diagram

Use one self-contained inline SVG with CSS animation only (no external scripts, no
external images). If the Page cannot render custom SVG or CSS animation, fall back to
a static SVG with the same nodes and labelled arrows, and say so.

Three horizontal lanes, tinted differently, with small uppercase lane labels:
Kubernetes cluster, Grafana Cloud, GitHub.

Nodes (rounded boxes, title plus one short subtitle):

- Kubernetes lane, left to right: Flux ("syncs every 15s"), storefront ("5 carts/s
  from loadgen"), checkout ("prices the cart").
- Grafana Cloud lane, left to right: Telemetry ("traces, metrics, logs") below
  checkout, Watcher ("scans on a schedule"), Investigation ("finds the root cause").
- GitHub lane, right to left: Coding agent ("sandbox, writes fix"), Fix PR
  ("assistant/* branch"), Copilot review ("automatic"), autoheal-merge ("guardrails,
  merge"), Build + release ("new image tag"), placed so Build + release sits below Flux.

Arrows, each with a short label: storefront -> checkout "cart"; checkout -> Telemetry
"OTLP"; Telemetry -> Watcher "scan"; Watcher -> Investigation "alert"; Investigation
-> Coding agent "fix task"; Coding agent -> Fix PR "push"; Fix PR -> Copilot review
"review"; Copilot review -> autoheal-merge "clean"; autoheal-merge -> Build + release
"merge"; Build + release -> Flux "bump deploy/ tag"; Flux -> storefront "deploy".
Add a dashed red arrow from Copilot review back to Fix PR labelled "comments: agent
pushes a fix (max 2 rounds)".

### Motion

- Idle: small dots travel continuously along the storefront -> checkout -> Telemetry
  arrows (the shop running). Slow, subtle.
- Auto-play story, 24 seconds, looping, with a visible Play/Pause button and a
  "Restart" button:
  1. 0-3s: everything calm and green. Caption: "A healthy shop."
  2. 3-5s: a "bad release" badge drops onto checkout, which turns red and pulses.
     Caption: "A refactor goes live. About 20% of checkouts fail."
  3. 5-11s: a glowing packet follows Telemetry -> Watcher -> Investigation ->
     Coding agent, each node lighting up as it is reached.
     Caption: "The Watcher sees the errors. The Investigation names the root cause."
  4. 11-17s: the packet continues Fix PR -> Copilot review. The dashed red loop
     flashes once. Caption: "The agent opens a fix PR. Copilot reviews it."
  5. 17-21s: packet through autoheal-merge -> Build + release -> Flux -> storefront.
     Caption: "Guardrails pass, the PR is merged, Flux deploys the fix."
  6. 21-24s: checkout turns green again with a short ring pulse.
     Caption: "Healed. Break to fix: about 12 minutes in real runs."
- Hovering a step in the numbered list pauses the story and highlights the matching
  nodes and arrows; leaving resumes it.
- Respect `prefers-reduced-motion`: show the static diagram with all nodes visible
  and no looping.

### Style

Follow the Grafana theme (light and dark). Lanes: green for Kubernetes, orange for
Grafana Cloud, blue for GitHub. Red only for the bug and the review loop. Nodes
stay readable at 12px or larger. The diagram scrolls horizontally on narrow screens
instead of shrinking text. Keep it clean: no decorative gradients, no emojis.

### Six steps (numbered list under the diagram)

1. Break: a pull request reverts the last pricing fix. An empty cart with code CAMP10
   panics on a nil map, so about 20% of checkouts return HTTP 500.
2. Ship: the push to main builds images tagged sha-<7>. The release job commits the
   tag to deploy/kustomization.yaml and Flux rolls it out in about 2 minutes.
3. Detect: the Watcher scans the shop (every 5 minutes while demoing) and sees the 5xx
   errors in traces and logs. It starts an Investigation.
4. Diagnose and fix: the Investigation finds the nil map and hands a fix to the coding
   agent. The agent works in a sandbox, adds a test that fails before the fix and
   opens a pull request on an assistant/* branch.
5. Review: a workflow runs vet and tests and requests a Copilot review. On comments the
   Assistant pushes a fix and Copilot looks again, at most 2 rounds.
6. Merge and heal: a clean review triggers a squash merge and a build. Flux deploys
   the new tag and the error ratio drops back to 0.

### Guardrails box (small, under the list)

Merges only PRs by the Assistant bot on assistant/* branches, only when the diff
touches internal/, only when vet and tests pass and Copilot's review is clean.
Anything else waits for a person.
