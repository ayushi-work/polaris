# Polaris

**AI-powered Kubernetes incident detection, root cause analysis, and self-healing.**

Polaris watches your cluster, spots when things break (OOMKilled, CrashLoopBackOff, etc.), investigates the root cause with an **agentic LLM investigation loop**, and can fix it automatically — restart, scale, or rollback, **escalating through runbook ladders**. It can also intentionally break things so you can verify the self-healing works.

All controllable from a web dashboard.

---

## Quick start

```bash
# 1. Add your DeepSeek API key
cp .env.example .env
# edit .env with your key

# 2. Start the server
go run ./cmd/polaris serve --dry-run

# 3. Start the frontend (separate terminal)
cd web && npm install && npm run dev
```

Open **http://localhost:5173**.

The `--dry-run` flag runs without a real Kubernetes cluster — detection and chaos run against a fake client.

---

## Architecture

```
CLI (Cobra)
  └─ API Server (Fiber + WebSocket)
       ├─ Detector ─── 6 rules incl. statistical restart anomaly (z-score)
       ├─ Correlator ── persists detections, 5m dedupe window
       ├─ RCA Engine ── agent loop → specialist agents → single call → heuristics
       │               agent tools: logs, events, deployment, prometheus, incident memory
       ├─ Runbooks ──── YAML escalation ladders (restart → scale_up → rollback)
       ├─ Remediation ─ Safety gate (whitelist, circuit breaker, pre-checks) + executor
       ├─ Chaos ─────── delete pods, CPU/memory stress, network netem, config faults
       ├─ Prometheus ── real resource metrics feed RCA context (optional)
       ├─ Alerting ──── Slack-compatible webhook on incident lifecycle (optional)
       └─ Orchestrator ─ central coordinator, subscribes to event bus

Storage: SQLite (default) or PostgreSQL, via GORM
K8s:     client-go (real cluster, kubeconfig, or fake for dev)
Frontend: React + TypeScript + Tailwind
```

## Project structure

```
cmd/polaris/         CLI entry point
internal/
  api/               HTTP handlers, WebSocket hub, middleware
  orchestrator/      Central coordinator: persistence, runbooks, escalation
  detector/          Pod watcher + detection rules (6, incl. anomaly z-score)
  rca/               Agent-loop RCA engine, LLM client, heuristics
  remediation/       Safety gate + restart / scale / rollback actions
  runbooks/          YAML runbook registry
  chaos/             Failure injection: delete/stress/netem/config-fault
  prometheus/        Prometheus API client (instant + range queries)
  alerting/          Slack-compatible webhook notifier
  eventbus/          In-memory pub/sub
  models/            Data models + SQLite/Postgres store
  kube/              Kubernetes client abstraction
  config/            Viper config loading
pkg/iforge/          Shared constants, error types
web/                 React frontend (Vite)
configs/
  runbooks/          Default runbook ladders
deployments/         Dockerfile, docker-compose, nginx, prometheus configs
```

## API

| Endpoint | Description |
|---|---|
| `GET /api/v1/incidents` | List incidents (filterable by status, severity, service) |
| `POST /api/v1/incidents` | Create an incident manually |
| `GET /api/v1/incidents/:id` | Incident detail with remediations and RCA |
| `PUT /api/v1/incidents/:id/acknowledge` | Acknowledge an incident |
| `PUT /api/v1/incidents/:id/resolve` | Resolve an incident |
| `GET /api/v1/incidents/:id/timeline` | Event timeline for an incident |
| `GET /api/v1/remediations` | List remediations |
| `POST /api/v1/remediations/:id/approve` | Approve a pending remediation |
| `POST /api/v1/remediations/:id/execute` | Execute a remediation |
| `GET /api/v1/analysis/:incident_id` | Get RCA result |
| `POST /api/v1/analysis/:incident_id` | Trigger RCA (calls DeepSeek) |
| `GET /api/v1/chaos/scenarios` | List chaos scenarios |
| `POST /api/v1/chaos/scenarios` | Create a chaos scenario |
| `POST /api/v1/chaos/scenarios/:id/execute` | Run a scenario immediately |
| `GET /api/v1/healthz` | Liveness probe |
| `GET /api/v1/readyz` | Readiness probe |
| `GET /api/v1/ws` | WebSocket (real-time events) |

## WebSocket events

`incident.created` `incident.updated` `remediation.started` `remediation.completed` `rca.completed` `chaos.executing` `chaos.completed`

## Detection rules

| Rule | What it catches |
|---|---|
| `oomkilled` | Containers killed by OOM killer |
| `crashloop` | Pods stuck in CrashLoopBackOff |
| `imagepull` | ImagePullBackOff or ErrImagePull |
| `podpending` | Pods stuck unschedulable |
| `nodepressure` | Nodes under disk/memory pressure |
| `anomaly-restarts` | Statistical restart-rate spikes (z-score vs own baseline) |

## Runbooks (escalation ladders)

Runbooks in `configs/runbooks/*.yaml` define ordered remediation steps. The
first step auto-creates on critical incidents; **failed steps escalate** to the
next rung (e.g. `restart → scale_up → rollback`). Human approval is required
unless `remediation.auto_approve: true`; a SafetyGate wraps every execution:
allowed-action whitelist, hourly circuit breaker, target pre-checks.

## Prometheus + Alerting (optional)

```yaml
metrics:
  prometheus:
    enabled: true
    url: "http://localhost:9090"

alerting:
  enabled: true
  webhook_url: "https://hooks.slack.com/services/..."
```

With Prometheus configured, RCA ships 30m memory/CPU/restart history as
evidence and a fourth specialist agent analyzes it. With a webhook URL, alerts
fire on every incident, RCA result, remediation and chaos injection.

## Kubernetes modes

| Mode | Flag | Behavior |
|---|---|---|
| `fake` | `--dry-run` | No cluster needed, uses fake clientset |
| `kubeconfig` | default | Uses local `~/.kube/config` |
| `in-cluster` | — | Uses pod service account (production) |

## Docker Compose

```bash
docker compose -f deployments/docker-compose.yml up       # sqlite + prometheus
docker compose -f deployments/docker-compose.yml --profile postgres up  # postgres backend
```

Polaris on :8080, UI on :5173, Prometheus on :9090.

## Configuration

Via `configs/polaris.yaml` or environment variables (prefixed with `POLARIS_`):

```yaml
llm:
  provider: deepseek
  model: deepseek-chat
  api_key: ""              # set POLARIS_LLM_API_KEY in .env
  base_url: https://api.deepseek.com/v1
```

## Frontend pages

| Route | Page |
|---|---|
| `/` | Dashboard — health, active incidents, reliability score, MTTR/MTTD |
| `/incidents` | Filterable incident list |
| `/incidents/:id` | Detail view with timeline, RCA panel, remediations |
| `/chaos` | Chaos Lab — inject failures from the UI |
| `/topology` | Service dependency graph (React Flow) |
| `/logs` | Live log viewer |
| `/metrics` | Uptime and incident metrics |
| `/healing` | Remediation audit trail |
| `/postmortems` | Generate and download postmortem reports |

## Tech stack

**Backend:** Go, Fiber, Cobra, Viper, GORM (SQLite/PostgreSQL), client-go
**Frontend:** React 18, TypeScript, Vite, Tailwind CSS, TanStack Query, Recharts, React Flow
**AI:** DeepSeek (OpenAI-compatible API) with tool-calling agent loop
**Observability:** Prometheus HTTP API (optional), Slack-compatible webhooks (optional)
