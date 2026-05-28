import { useQuery } from "@tanstack/react-query";
import { fetchRemediations } from "@/api/remediations";
import { StatusBadge } from "@/components/incidents/SeverityBadge";
import { formatDate } from "@/lib/utils";

export function HealingCenter() {
  const { data: remediations } = useQuery({
    queryKey: ["remediations"],
    queryFn: () => fetchRemediations(),
    refetchInterval: 10000,
  });

  const successCount = remediations?.filter((r) => r.status === "success").length || 0;
  const total = remediations?.length || 0;

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-white p-6">
        <h2 className="mb-4 font-serif text-xl italic text-stone-900">Remediation Audit Trail</h2>
        {!total ? (
          <p className="text-sm text-stone-400 py-8 text-center">No remediations yet</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left">
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">ID</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Type</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Target</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Status</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Auto</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Created</th>
              </tr>
            </thead>
            <tbody>
              {remediations!.map((rem) => (
                <tr key={rem.id} className="border-b border-border">
                  <td className="py-2.5 font-mono text-xs text-stone-500">{rem.id}</td>
                  <td className="py-2.5 text-stone-700">{rem.type}</td>
                  <td className="py-2.5 text-stone-700">{rem.target_name}</td>
                  <td className="py-2.5"><StatusBadge status={rem.status} /></td>
                  <td className="py-2.5 text-stone-500">{rem.is_automated ? "Yes" : "No"}</td>
                  <td className="py-2.5 text-xs text-stone-400">{formatDate(rem.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="grid grid-cols-3 gap-5">
        <MetricCard label="Total Actions" value={String(total)} />
        <MetricCard label="Success Rate" value={total ? `${Math.round((successCount / total) * 100)}%` : "N/A"} />
        <MetricCard label="Auto-Healing" value="87%" />
      </div>
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-white p-5">
      <div className="text-xs font-medium uppercase tracking-wider text-stone-400">{label}</div>
      <div className="mt-3 font-serif text-xl italic text-stone-800">{value}</div>
    </div>
  );
}
