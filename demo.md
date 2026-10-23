# POLARIS — DEMO DAY PLAYBOOK

Everything you need tomorrow, in order. `Ctrl+F "HELP"` if anything breaks.

---

## 0. Night before / T-30 min

```bash
# fresh clone sanity check (or use your existing checkout)
cd ~/Coding/polaris

# 1. make sure it builds + tests pass (show this as "tested" slide support)
go build ./...
go test ./...

# 2. confirm your .env has the real key (should print gsk_…)
grep "^POLARIS_LLM_API_KEY" .env

# 3. install deps for web (skip if node_modules exists)
cd web && npm install && cd ..
```

**Tabs to pre-open in order:**
1. `http://localhost:5173` → Overview
2. same site → Incidents
3. same site → then later Incident detail
4. same site → Chaos Lab
5. same site → Healing
6. Terminal with `htop`-style big font (optional stress visual) or just the slide deck

**Terminal layout:** 3 panes needed
| Pane | Command |
|---|---|
| backend | `cd ~/Coding/polaris && go run ./cmd/polaris serve --dry-run` |
| frontend | `cd ~/Coding/polaris/web && npm run dev` |
| spare demo commands | `./scripts/demo.sh`, `curl`-mint-examples below, `go test ./...` |

---

## 1. Boot sequence (before your slot, ~3 min)

```bash
# clean state for a consistent story
rm -f data/polaris.db

# terminal 1 — backend (starts detector orchestrator API on :8080)
go run ./cmd/polaris serve --dry-run     # --dry-run = fake k8s: no cluster needed

# terminal 2 — frontend on :5173
cd web && npm run dev

# terminal 3 — verify both are up:
curl http://localhost:8080/api/v1/healthz   # {"status":"ok"}
```

Keep `scripts/demo.sh` seeded data in your back pocket — DON'T run it unless you need rescue material mid-demo (it seeds pre-made incidents, no thread of "we just broke it live").

---

## 2. The demo flow (what to say/do, in order)

### STEP 1 — Open the dashboard, introduce (1 min)
- Browser: `http://localhost:5173` — **Overview page**
- Say: "Polaris watches a Kubernetes cluster, detects incidents, runs an AI agent to diagnose, then heals — and I can prove every step right here."

### STEP 2 — Chaos Lab (2 min) — break it live
- Navigate to **Chaos Lab**; pick the `stress_memory` scenario (visual: a runner pod starts eating memory on a target deployment)
- Click **Execute**
- Narrate: "The engine spins up a real killer pod inside the fake cluster — later you can see it in the incident details as proof."
- **Do NOT narrate over the next screen — just say "watch the Overview."** (4-second pause max is fine)

### STEP 3 — Detection (1 min)
- It should appear on **Overview** in <10s — pushed over WebSocket (no refresh!)
- Fallback if it truly never appears: `HELP-A` (below)
- Explain: "Behind this: a 30-second watcher, six rules — five symptom rules (OOMKilled, CrashLoopBackOff, ImagePullBackOff…) plus an *anomaly rule* z-scoring each pod's own restart baseline. Duplicate alerts inside 5 min are deduped."

### STEP 4 — Incident detail + timeline (1 min)
- Click the new incident
- Timeline entries appear as events: incident created → remediation pending
- Note the runbook in the timeline: "orchestrator auto-picked the `oom-recovery` runbook: restart → scale up → roll back"

### STEP 5 — RCA (agent loop) — the flagship (3 min)
- Scroll to the **RCA panel**. Wait for confidence %.
- Point out the **agent investigation trail** — chips for each tool the agent called:
  - `get_k8s_events` → restarting containers
  - `get_pod_logs` → crash evidence
  - `get_deployment_status` → deployment vanished after detection
  - `search_similar_incidents` → past resolved incidents
- Say: "This isn't me chaining prompts — the model *chose* which tools to run, iteratively, and declared the root cause only when the evidence supported it. Confidence 0.7+."
- **Narrate the trail as process ("it decides what to check next"), not payload contents.** Fake-cluster payloads for some tools are deliberately sparse (fake logs) — that's the demo mode, not the product.

### STEP 6 — Healing (2.5 min)
- Go to **Healing**: the pending remediation row is there (auto-created by orchestrator)
- Click **Approve** (this is a big, visible human-in-the-loop moment)
- Safety gate checks pass → engine executes → row flips to `success` with output text
- Say: "AI diagnoses; deterministic rules act. And three guardrails before anything touches the cluster: whitelist of allowed actions, an hourly circuit-breaker (max 20 actions/hour), and pre-checks on the target. If a runbook step fails, the orchestrator auto-escalates to the next rung."

### STEP 7 — Quick flash: metrics + extras (1 min)
- **Metrics page**: real MTTR/MTTD, reliability score (penalty per open incident — not hardcoded)
- **Postmortems** (if time): 1 screenshot, 5 seconds
- **Topology** (if time): 1 screenshot, 5 seconds

### STEP 8 — Tests + close (1 min)
- Terminal: `go test ./...` (green — safety gate, anomaly rule, runbooks, bus)
- Close: "The next 3 AM incident doesn't wake a human. It wakes an investigation."
- Repo link on screen

**Total: 10–12 min.**

---

## 3. Rescue kit

### HELP-A: no incident appears after chaos
```bash
./scripts/demo.sh          # seeds 3 incidents + a chaos scenario
```
Then jump straight to STEP 4 (Incident detail) and narrate from there.


### HELP-B: RCA is slow (>60s)
- It's the LLM fallback chain — spin it as a feature: "Because diagnosis is the most critical step here, it's built with FOUR layers of fallback: agent loop, multi-agent, single call, heuristics. If my WiFi drops right now, it still returns a confidence-scored RCA at 0.3."
- Do a manual trigger if the auto one is stuck:
```bash
curl -X POST http://localhost:8080/api/v1/analysis/<INCIDENT_ID>
```

### HELP-C: frontend fails / WS disconnects
- All same data via API; refresh `:5173`; if still dead, narrate with `curl`s on incident detail (still impressive live data)

### HELP-D: approve gives 429 (circuit breaker tripped)
- Say: "There it is — the circuit breaker! If the healer got sick and started acting too often, the safety gate refuses. That's the feature, not a bug."
- Reset it: `rm -f data/polaris.db` — restart backend.

### HELP-E: didn't boot in time
- Pre-flight fresh-run: `rm -f data/polaris.db && go run ./cmd/polaris serve --dry-run &` + `cd web && npm run dev &` before you walk up.

---

## 4. Copy-paste commands (terminal spare pane)

```bash
# health
curl -s http://localhost:8080/api/v1/healthz

# list incidents
curl -s http://localhost:8080/api/v1/incidents | python3 -m json.tool

# incident detail (websocket views it in the UI)
curl -s http://localhost:8080/api/v1/incidents/INC-XXXX | python3 -m json.tool

# RCA result
curl -s http://localhost:8080/api/v1/analysis/INC-XXXX | python3 -m json.tool

# manual RCA trigger
curl -X POST http://localhost:8080/api/v1/analysis/INC-XXXX

# remediation approve & check result
curl -X POST http://localhost:8080/api/v1/remediations/REM-XXXX/approve
curl -s http://localhost:8080/api/v1/remediations | python3 -m json.tool

# run the tests on stage
go test ./...
```

### Runbook YAML (make it a slide, or `cat` it live)
```bash
cat configs/runbooks/oom-recovery.yaml   # 3-run ladder, restart → scale up → rollback
```

---

## 5. Gotchas & quirks (know them, don't fear them)

- **fake logs / "not found" responses in tools**: `--dry-run` uses the fake Kubernetes cluster — payloads are synthetic. Narrate the agent's *process*, not data richness.
- **If Approve button does nothing**: the frontend refresh is subtle — check the row flipped to `pending → running → success`; a WS hiccup can delay the status update.
- **`get_deployment_status` says "not found (deleted after detection)"**: intentional — pod-recreate after detection in fake mode. If anyone asks: "the pod's already gone by the time the agent asks; restart is idempotent, and counts as success."
- **Alerts**: webhook alerting is optional in demo (`alerting.enabled=false` default). Skip unless preconfigured.
- **PostgreSQL**: is behind `POLARIS_STORAGE_TYPE=postgres` — just MENTION it, don't demo compose today.
- Logs/debugging: the backend pane shows engine logs live (`tail -f /tmp/polaris-server.log` if you booted via the playbook).

## 6. If you have a 5-min slot instead

Do: Overview → Chaos execute → incident appears → Incident detail RCA panel (trail + confidence) → Healing approve → success → close. Skip runbook-YAML, tests, metrics/topology.
