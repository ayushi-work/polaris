import { Activity } from "lucide-react";

interface Props {
  health: string;
}

const colors: Record<string, string> = {
  Healthy: "text-emerald-600",
  Warning: "text-amber-600",
  Degraded: "text-red-600",
};

const dots: Record<string, string> = {
  Healthy: "bg-emerald-500",
  Warning: "bg-amber-400",
  Degraded: "bg-red-500",
};

export function ClusterHealthCard({ health }: Props) {
  return (
    <div className="rounded-xl border border-border bg-white p-5">
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-stone-400">
        <Activity className="h-3.5 w-3.5" />
        Cluster Health
      </div>
      <div className="mt-3 flex items-center gap-2">
        <div className={`h-2.5 w-2.5 rounded-full ${dots[health] || "bg-stone-400"}`} />
        <span className={`text-2xl font-serif italic ${colors[health] || "text-stone-500"}`}>
          {health}
        </span>
      </div>
    </div>
  );
}
