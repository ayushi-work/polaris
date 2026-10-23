import { useQuery } from "@tanstack/react-query";
import api from "@/api/client";
import type { DashboardMetrics } from "@/types";

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  const m = Math.floor(seconds / 60);
  const s = Math.round(seconds % 60);
  if (m < 60) return `${m}m ${s}s`;
  const h = Math.floor(m / 60);
  return `${h}h ${m % 60}m`;
}

export function useDashboardMetrics() {
  return useQuery({
    queryKey: ["dashboard"],
    queryFn: async (): Promise<DashboardMetrics> => {
      const { data: incidents } = await api.get("/incidents", { params: { limit: "100" } });
      const active = incidents.filter((i: { status: string }) =>
        ["detected", "investigating", "remediating"].includes(i.status)
      ).length;
      const resolvedIncidents = incidents.filter((i: { status: string }) =>
        i.status === "resolved"
      );
      const resolved = resolvedIncidents.length;
      const today = incidents.filter((i: { detected_at: string }) => {
        const d = new Date(i.detected_at);
        const now = new Date();
        return d.toDateString() === now.toDateString();
      }).length;

      // Reliability: start at 100, penalize unresolved incidents by severity.
      const penalty = incidents.reduce((acc: number, i: { status: string; severity: string }) => {
        if (!["detected", "investigating", "remediating"].includes(i.status)) return acc;
        return acc + (i.severity === "critical" ? 15 : i.severity === "warning" ? 7 : 3);
      }, 0);
      const reliability = Math.max(0, 100 - penalty);

      // MTTR: mean time from detection to resolution across resolved incidents.
      const mttrSeconds = resolvedIncidents.reduce((acc: number, i: { detected_at: string; resolved_at: string }) => {
        const ms = new Date(i.resolved_at).getTime() - new Date(i.detected_at).getTime();
        return acc + Math.max(0, ms) / 1000;
      }, 0) / (resolved || 1);

      // MTTD: mean time from detection to last status transition (investigation latency).
      const mttdSeconds = incidents.reduce((acc: number, i: { detected_at: string; updated_at: string }) => {
        const ms = new Date(i.updated_at).getTime() - new Date(i.detected_at).getTime();
        return acc + Math.max(0, ms) / 1000;
      }, 0) / (incidents.length || 1);

      return {
        cluster_health: active > 2 ? "Degraded" : active > 0 ? "Warning" : "Healthy",
        incidents_today: today,
        active_incidents: active,
        auto_recovered: resolved,
        reliability_score: reliability,
        mttr: resolved > 0 ? formatDuration(mttrSeconds) : "—",
        mttd: incidents.length > 0 ? formatDuration(mttdSeconds) : "—",
      };
    },
    refetchInterval: 10000,
  });
}
