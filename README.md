# Polaris

**AI-powered Kubernetes incident detection, root cause analysis, and self-healing.**

Polaris watches your cluster, spots when things break (OOMKilled, CrashLoopBackOff, etc.), asks an LLM to figure out why, and can fix it automatically — restart, scale, or rollback. It can also intentionally break things so you can verify the self-healing works.

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
       ├─ Detector ─── watches pods, fires incident events
       ├─ RCA Engine ── gathers logs/events, calls DeepSeek for root cause
       ├─ Remediation ─ restarts, scales, or rolls back broken workloads
       ├─ Chaos ─────── injects failures (delete pods, stress, network)
       └─ Orchestrator ─ central coordinator, subscribes to event bus

Storage: SQLite via GORM
K8s:      client-go (real cluster, kubeconfig, or fake for dev)
Frontend: React + TypeScript + Tailwind
```

## Project structure

```
cmd/polaris/         CLI entry point
internal/
  api/               HTTP handlers, WebSocket hub, middleware
  orchestrator/      Central coordinator
  detector/          Pod watcher + detection rules
  rca/               Root cause analysis engine + LLM client
  remediation/       Self-healing actions
  chaos/             Failure injection engine
  eventbus/          In-memory pub/sub
  models/            Data models + SQLite store
  kube/              Kubernetes client abstraction
  config/            Viper config loading
pkg/iforge/          Shared constants, error types
web/                 React frontend (Vite)
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

## Kubernetes modes

| Mode | Flag | Behavior |
|---|---|---|
| `fake` | `--dry-run` | No cluster needed, uses fake clientset |
| `kubeconfig` | default | Uses local `~/.kube/config` |
| `in-cluster` | — | Uses pod service account (production) |

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

**Backend:** Go, Fiber, Cobra, Viper, GORM, SQLite, client-go
**Frontend:** React 18, TypeScript, Vite, Tailwind CSS, TanStack Query, Recharts, React Flow
**AI:** DeepSeek (OpenAI-compatible API)
