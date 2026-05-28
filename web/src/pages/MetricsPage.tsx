import { useQuery } from "@tanstack/react-query";
import api from "@/api/client";

export default function MetricsPage() {
  const { data: incidents } = useQuery({
    queryKey: ["incidents"],
    queryFn: async () => {
      const { data } = await api.get("/incidents", { params: { limit: "50" } });
      return data;
    },
  });

  const critical = incidents?.filter((i: { severity: string }) => i.severity === "critical").length || 0;
  const warning = incidents?.filter((i: { severity: string }) => i.severity === "warning").length || 0;
  const total = incidents?.length || 1;
  const uptimePct = Math.max(0, 100 - (critical * 2 + warning * 0.5)).toFixed(1);

  const metrics = [
    { label: "Estimated Uptime", value: `${uptimePct}%`, color: "text-emerald-600" },
    { label: "Total Incidents", value: String(total), color: "text-stone-800" },
    { label: "Critical", value: String(critical), color: critical > 0 ? "text-red-600" : "text-emerald-600" },
    { label: "Warnings", value: String(warning), color: "text-amber-600" },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Metrics</h1>
        <p className="mt-1 text-sm text-stone-500">Reliability and performance overview.</p>
      </div>
      <div className="grid grid-cols-2 gap-5">
        {metrics.map(({ label, value, color }) => (
          <div key={label} className="rounded-xl border border-border bg-white p-5">
            <div className="text-xs font-medium uppercase tracking-wider text-stone-400">{label}</div>
            <div className={`mt-3 font-serif text-2xl italic ${color}`}>{value}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
