import { Activity } from "lucide-react";

interface Props {
  health: string;
}

export function ClusterHealthCard({ health }: Props) {
  const colors: Record<string, string> = {
    Healthy: "text-emerald-400",
    Warning: "text-amber-400",
    Degraded: "text-red-400",
  };

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Activity className="h-4 w-4" />
        Cluster Health
      </div>
      <div className={`mt-2 text-2xl font-bold ${colors[health] || "text-gray-400"}`}>
        {health}
      </div>
    </div>
  );
}
