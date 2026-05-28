export interface Incident {
  id: string;
  cluster: string;
  namespace: string;
  kind: string;
  resource_name: string;
  incident_type: string;
  severity: "info" | "warning" | "critical";
  status: "detected" | "investigating" | "remediating" | "resolved" | "failed";
  message: string;
  details: string;
  detected_at: string;
  resolved_at: string | null;
  updated_at: string;
  remediations?: Remediation[];
  rca_result?: RCAResult;
}

export interface Remediation {
  id: string;
  incident_id: string;
  type: string;
  status: "pending" | "running" | "success" | "failed" | "skipped";
  target_kind: string;
  target_name: string;
  namespace: string;
  parameters: string;
  output: string;
  error_message: string;
  is_automated: boolean;
  created_at: string;
  executed_at: string | null;
  completed_at: string | null;
}

export interface RCAResult {
  id: string;
  incident_id: string;
  summary: string;
  root_cause: string;
  confidence: number;
  logs_snippet: string;
  events_summary: string;
  suggested_actions: string;
  raw_llm_output: string;
  llm_model: string;
  prompt_tokens: number;
  completion_tokens: number;
  created_at: string;
}

export interface ChaosScenario {
  id: string;
  name: string;
  description: string;
  action: string;
  target: string;
  parameters: string;
  schedule: string;
  enabled: boolean;
  run_count: number;
  last_run_at: string | null;
  created_at: string;
}

export interface TimelineEntry {
  timestamp: string;
  event: string;
  details: string;
}

export interface DashboardMetrics {
  cluster_health: string;
  incidents_today: number;
  active_incidents: number;
  auto_recovered: number;
  reliability_score: number;
  mttr: string;
  mttd: string;
}

export interface WSEvent {
  type: string;
  timestamp: string;
  payload: unknown;
}
