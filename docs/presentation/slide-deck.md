# Polaris — Tech Fest Slide Deck (~12 min, 13 slides)

Paste each slide block into your slide tool. "VISUAL" = what's on screen; "SAY" = talking script; time budget on the right.

---

## Slide 1 — Title
**POLARIS**
*AI-powered Kubernetes incident response: Detect → Diagnose → Heal → Verify (autonomously)*

Your name • College • Tech Fest 2026

- **VISUAL:** Logo/title on dark background (match the dashboard theme).
- **SAY:** "When a pod dies at 3 AM, an on-call engineer spends 20+ minutes just finding the root cause. Polaris investigates, explains, and heals — in seconds."
- ⏱ 30s

---

## Slide 2 — The Problem
**Production breaks. Humans are slow.**

- Kubernetes fails in confusing ways: OOMKilled, CrashLoopBackOff, ImagePullBackOff, node pressure…
- Alerts are noisy and arrive with **zero context**
- Root cause is buried in logs, events, metrics, and tribal knowledge
- Result: high MTTR, burnt-out on-call, manual postmortems

- **VISUAL:** chaotic "alerts wall" graphic or raw kubectl spam screenshot.
- **SAY:** "And the first fix humans reach for is usually 'restart and pray' — because diagnosing takes too long."
- ⏱ 1 min

---

## Slide 3 — Meet Polaris
**An SRE copilot that:**

1. **Detects** — 6 detection rules including *statistical anomaly detection*
2. **Diagnoses** — an **agent loop** gathers its own evidence, then explains the root cause with a confidence score
3. **Heals** — runbook escalation ladders: restart → scale up → roll back, wrapped in a safety gate
4. **Verifies** — built-in Chaos Lab (delete pods, CPU/memory stress, config faults) proves the loop closes

- **VISUAL:** the 4-step pipeline as a horizontal arrow flow.
- **SAY:** "Everything is one Go binary plus a web dashboard. Runs against a real cluster — or a fake one, so you can demo the whole thing on a laptop with no infrastructure."
- ⏱ 45s

---

## Slide 4 — Architecture
**One Go server, event-driven, one nervous system**

```
CLI (Cobra) → API Server (Fiber + WebSocket)
   ├─ Detector       6 rules, incl. z-score anomaly rule
   ├─ Event Bus      in-memory pub/sub — the nervous system
   ├─ Orchestrator   persistence + dedupe + runbook triage + escalation
   ├─ RCA Engine     AGENT LOOP → 4 specialists → single LLM → heuristics
   ├─ Prometheus     real memory/CPU metrics feed RCA evidence (optional)
   ├─ Runbook Reg.   YAML escalation ladders, auto-escalate on failure
   ├─ Remediation    SafetyGate (whitelist + circuit breaker) + executor
   ├─ Chaos Engine   delete / stress_cpu / stress_memory / netem / config_fault
   ├─ Alerting       Slack-compatible webhook on every lifecycle event
   └─ Store          SQLite or PostgreSQL | UI: React + TS + Tailwind
```

- **VISUAL:** redraw the diagram nicely.
- **SAY:** "Components never call each other directly — they talk through an event bus. Adding a capability means adding one subscriber. Detector incidents are persisted with a dedupe window, and every critical incident is automatically triaged into a runbook."
- ⏱ 1.5 min

---

## Slide 5 — LIVE DEMO — Break it
**Chaos Lab**

- Execute a scenario live: pod deletion or **memory stress runner pod**
- The chaos engine really runs in the fake cluster — the runner pod exists, the result is captured in the incident details

- **VISUAL:** Chaos Lab page; show the created runner in the logs if time allows.
- **SAY:** "Netflix popularized chaos engineering. I built it in so self-healing is *proven*, not claimed."
- ⏱ 2 min

---

## Slide 6 — LIVE DEMO — Detect
**Real-time, persisted, deduped**

- Incident appears on the Overview **without refresh** (WebSocket push)
- Detector runs 6 rules: 5 symptom rules + **anomaly rule** that z-scores restart rates against each pod's own baseline and fires on outliers
- 5-minute dedupe window prevents alert storms

- **VISUAL:** Dashboard + Incidents page; filter by severity.
- ⏱ 1 min

---

## Slide 7 — LIVE DEMO — Diagnose
**The agent loop (the flagship)**

- On `rca.triggered`, an **autonomous investigation agent** starts with a toolset:
  `get_pod_logs` · `get_k8s_events` · `get_deployment_status` · `get_pod_metrics` · `search_similar_incidents`
- It iterates up to 6 rounds, gathering its own evidence
- Then declares root cause + confidence as a final JSON answer
- The UI shows the **full investigation trail** — every tool call and result

- **VISUAL:** Incident detail page → RCA panel: "Agent investigation trail" chips + outputs.
- **SAY:** "This isn't one prompt dumped into ChatGPT. The model *investigates like an engineer* — looks at events, tails logs, checks the deployment, compares against past incidents — and cites exact evidence. And if the LLM can't do it, the pipeline degrades: multi-agent → single call → rule-based heuristics. Diagnosis never fully fails."
- ⏱ 2.5 min

---

## Slide 8 — LIVE DEMO — Heal
**Runbooks + Safety Gate + human approval**

- Critical incident → orchestrator auto-triages into `oom-recovery` runbook → step 1: restart (pending)
- **Approve from the UI** → SafetyGate checks (whitelist, circuit breaker, pre-checks) → engine executes → success/failed captured with output
- If a step fails → the NEXT runbook step auto-escalates (scale_up → rollback)
- Flash config: `auto_approve`, `max_actions_per_hour`, `allowed_actions`

- **VISUAL:** Healing page — Approve/Cancel buttons, then the success row.
- **SAY:** "AI diagnoses; deterministic rules act. Autonomy with brakes."
- ⏱ 2.5 min

---

## Slide 9 — What flows around it
- **Alerting** — every incident, RCA, remediation and chaos event → Slack-compatible webhook
- **Postmortems** — generate & download incident reports
- **Topology** — service dependency graph (React Flow)
- **Metrics** — real computed MTTR/MTTD + reliability score

- **VISUAL:** 2×2 screenshot grid. 5s each.
- ⏱ 1 min

---

## Slide 10 — Engineering Choices
| Decision | Why |
|---|---|
| Event-driven pub/sub | Decoupled; every new feature is a subscriber |
| Agent loop with tools, not one-shot prompting | Evidence-grounded diagnosis; trail is auditable |
| 4-layer RCA fallback chain | Diagnosis degrades, never dies |
| Runbook ladders + SafetyGate | Deterministic action, human approval, rate limits |
| Fake Kubernetes client | Entire stack demos with zero infra (and full tests run against it) |
| SQLite/Postgres swap behind one Store interface | Dev → prod with a config flag |

- ⏱ 1 min

---

## Slide 11 — Roadmap (honest: what's next)
- Prometheus full ingestion (cAdvisor + kube-state-metrics scrapes)
- Slack/PagerDuty two-way ack from the alert thread
- Multi-cluster federation
- LLM *selects* the runbook (constrained choice), engine still executes
- Postmortem drafting agent (timeline + RCA → SEV-deck)
- Chaos scheduling engine (cron-based steady-state verification)

- ⏱ 30s

---

## Slide 12 — Tested, not vibes
- Regression tests for the safety gate (whitelist, circuit breaker), the anomaly rule (storm detected / steady churn silent), runbook selection/escalation, and the event bus
- `go test ./...` — the fake cluster makes the whole system testable

- **VISUAL:** terminal running `go test ./...` all green.
- ⏱ 30s

---

## Slide 13 — Close
**"Detect. Diagnose. Heal. Verify."**

Polaris — GitHub: <your repo link>

- **SAY:** "The next incident at 3 AM doesn't wake a human. It wakes an investigation."
- ⏱ 20s

**Total ≈ 13 min live. Trim to fit: merge Slides 9+12, and cut Slide 11 for a 10-min slot.**
