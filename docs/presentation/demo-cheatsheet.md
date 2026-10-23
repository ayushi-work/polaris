# POLARIS — DEMO DAY CHEAT SHEET (print this)

## ⏱ Run sheet (12-13 min)
| Time | Action | If it fails |
|---|---|---|
| T-10m before | Boot backend + web, `rm -f data/polaris.db` for clean state, open tabs | — |
| 0:00 | Hook: "3 AM pager → 20 min to root cause. Polaris: an investigation that heals." | — |
| 0:30 | Problem slide (alerts w/o context, MTTR) | — |
| 1:30 | Architecture slide (event bus, agent loop, runbooks) | — |
| 3:00 | Chaos Lab → execute `stress_memory` (or delete_pod) LIVE | Use seeded scenario from demo.sh |
| 4:00 | Dashboard: incident appears LIVE via WebSocket (don't refresh!) | Fallback: incident already created by demo.sh |
| 5:00 | Incident detail → timeline story → RCA panel → agent trail + root cause + confidence | WiFi/key dead: RCA still returns 30% heuristic + detail page still shines |
| 6:30 | Show `configs/runbooks/oom-recovery.yaml` (30s) — "3-rung ladder" | — |
| 7:00 | Healing page → pending remediation → click APPROVE → success row + output text | If blocked by breaker: demo the 429 message itself — it's a feature |
| 8:30 | Alerting: show 1-2 Slack messages if webhoook configured | Skip silently — don't show config screens |
| 9:00 | Flash: Postmortems → Topology → Metrics (real MTTR) → Postmortems 5s each | — |
| 9:45 | Engineering choices table + tests slide (`go test ./...` green) | — |
| 11:30 | Roadmap → close: "Detect. Diagnose. Heal. Verify." | — |

## 💻 Boot sequence (before presenting)
```bash
# terminal 1 — backend (fake cluster, no k8s needed)
rm -f data/polaris.db          # clean state
go run ./cmd/polaris serve --dry-run      # API on :8080

# terminal 2 — frontend
cd web && npm run dev                     # UI on http://localhost:5173

# terminal 3 — (optional) seed a scenario + incidents so the slow-loading
# demo always has fallback material
./scripts/demo.sh
```

## 🔑 Key numbers (quote these — they prove depth)
- 6 detection rules — 5 symptom rules + **anomaly rule (z-score ≥ 2.5 vs pod's own restart baseline)**
- **4-layer RCA fallback**: agent loop → 4 specialists (logs/events/deploy/metrics) → single call → heuristics. "Diagnosis never fully fails."
- **5 agent tools**: `get_pod_logs`, `get_k8s_events`, `get_deployment_status`, `get_pod_metrics`, `search_similar_incidents`; max 6 investigation rounds
- **Runbooks**: YAML escalation ladders (`configs/runbooks/`) — failed step auto-escalates to next rung
- **SafetyGate**: allowed_actions whitelist + hourly circuit breaker (`max_actions_per_hour: 20`) + target pre-checks + idempotent restart (pod-gone = success)
- Dedupе window: 5 min (orchestrator cross-checks DB before persisting detections)
- Chaos: `delete_pod`, `stress_cpu`, `stress_memory` (runner pods), `network_delay/loss` (gated tc netem), `config_fault` (env patch)
- First-class fallbacks: RCA works even with an invalid LLM key (confidence 0.3 heuristics)
- Storage: SQLite default, **PostgreSQL** behind `POLARIS_STORAGE_TYPE=postgres`
- Stack: Go + Fiber + Cobra + Viper + GORM + client-go | React 18 + TS + Vite + TanStack Query + Recharts + React Flow
- `go test ./...` green: safety gate, anomaly rule, runbooks, event bus

## 🎯 The 20-second thesis (memorize)
"Polaris watches a Kubernetes cluster, runs four detection rules plus a statistical anomaly detector, and on every critical incident the orchestrator auto-triages it into a runbook. An autonomous agent then investigates using real evidence — logs, events, metrics, past incidents — and declares a root cause with confidence. A safety-gated engine heals the issue, and if that step fails, the runbook escalates automatically. We prove it by breaking the cluster on purpose."

## 🛡️ Q&A quick answers
- **"Is it really on Kubernetes?"** → client-go everywhere: 3 modes — kubeconfig, in-cluster, fake (`--dry-run`), and the chaos engine's runner pods exist inside it. `scripts/setup-kind.sh` gives you a real kind cluster.
- **"What if the LLM hallucinates?"** → Evidence-by-construction: the agent can only cite what its tools returned. Actions are rule-driven and whitelisted — the LLM never executes anything. Human approval ON by default; circuit breaker caps at 20 actions/hour.
- **"Why an agent loop instead of one prompt?"** → One prompt = one answer, no evidence trail. The agent gathers logs/events/metrics itself and the UI shows every tool call — it's auditable. If the model fails, the fallback chain keeps the pipeline alive.
- **"How does detection work? Prometheus?"** → Direct pod/node watch via client-go every 30s + rule engine, PLUS a stateful anomaly rule z-scoring each pod's own restart baseline against a rolling window. Prometheus enriches RCA context (optional), doesn't gate detection.
- **"What happens if the same alert fires 50 times?"** → Two layers: detector fires once per pod/type per 5 min; the orchestrator dedupes against the DB inside the dedupe window before persisting.
- **"Say the healing action fires and the pod is already gone?"** → Restart is idempotent: 'already gone (recreated after detection)' counts as success. Show the exact output string in the Healing page.
- **"Postgres?"** → One `Store` interface; sqlite/postgres behind `POLARIS_STORAGE_TYPE`. Same model, AutoMigrate either.
- **"What would you build next?"** → Prometheus full ingestion for anomaly-of-metrics (memory z-score like restarts), two-way Slack ack, multi-cluster.

## ⚠️ Day-before checklist
- [ ] Real DeepSeek key in `.env` (agent loop is the money shot with a key)
- [ ] Run the full flow once: chaos → incident → RCA → approve → success
- [ ] Keep one cached RCA result on screen as WiFi fallback — but remember: **heuristics still produce an RCA at 30% confidence with a dead key**. You're never dead in the water.
- [ ] `go test ./...` green in a terminal, ready for the tests slide
- [ ] Tabs ordered: Overview / Incidents / Incident detail / Chaos Lab / Healing / Metrics
- [ ] Big terminal font, dark theme, browser zoom 110%
- [ ] Timed run at home once — cut Slide 11 roadmap if over 10 min

## 🗣 One-liners that land
- "Components don't call each other — they talk through an event bus. New features are new subscribers."
- "The agent investigates like an engineer: events first, logs second, deployment status third. Cited evidence, not vibes."
- "AI diagnoses, rules act — that's how you get autonomy without recklessness."
- "If a healing step fails, the runbook doesn't retry it — it ESCALATES to the next rung."
- "We break things on purpose, so you never have to trust a slide deck."
