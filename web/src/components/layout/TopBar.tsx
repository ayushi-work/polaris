import { useDashboardMetrics } from "@/hooks/useMetrics";

export function TopBar() {
  const { data } = useDashboardMetrics();

  const statusColor =
    data?.cluster_health === "Healthy"
      ? "bg-emerald-500"
      : data?.cluster_health === "Warning"
        ? "bg-amber-500"
        : "bg-red-500";

  return (
    <header className="flex h-14 items-center justify-between border-b border-border bg-card px-6">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className={`h-2 w-2 rounded-full ${statusColor} animate-pulse`} />
          <span className="text-sm text-muted-foreground">
            Cluster: {data?.cluster_health || "Unknown"}
          </span>
        </div>
        <span className="text-xs text-muted-foreground">|</span>
        <span className="text-xs text-muted-foreground">
          {data?.active_incidents || 0} active incidents
        </span>
      </div>
      <span className="text-xs text-muted-foreground">v0.1.0</span>
    </header>
  );
}
