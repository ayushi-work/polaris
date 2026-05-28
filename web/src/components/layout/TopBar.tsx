import { useDashboardMetrics } from "@/hooks/useMetrics";

export function TopBar() {
  const { data } = useDashboardMetrics();

  const statusColor =
    data?.cluster_health === "Healthy"
      ? "bg-emerald-500"
      : data?.cluster_health === "Warning"
        ? "bg-amber-400"
        : "bg-red-500";

  return (
    <header className="flex h-14 items-center justify-between border-b border-border bg-white px-6">
      <div className="flex items-center gap-4 text-sm">
        <div className="flex items-center gap-2">
          <div className={`h-2 w-2 rounded-full ${statusColor}`} />
          <span className="text-stone-500">
            Cluster:{" "}
            <span className="text-stone-700 font-medium">{data?.cluster_health || "Unknown"}</span>
          </span>
        </div>
        <span className="text-stone-300">|</span>
        <span className="text-stone-500">
          <span className="text-stone-700 font-medium">{data?.active_incidents || 0}</span> active
        </span>
      </div>
      <span className="text-xs text-stone-400 font-mono">v0.1.0</span>
    </header>
  );
}
