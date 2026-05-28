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

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Remediation Audit Trail</h2>
        {!remediations?.length ? (
          <p className="text-sm text-muted-foreground">No remediations yet</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left">
                  <th className="pb-2 font-medium text-muted-foreground">ID</th>
                  <th className="pb-2 font-medium text-muted-foreground">Type</th>
                  <th className="pb-2 font-medium text-muted-foreground">Target</th>
                  <th className="pb-2 font-medium text-muted-foreground">Status</th>
                  <th className="pb-2 font-medium text-muted-foreground">Auto</th>
                  <th className="pb-2 font-medium text-muted-foreground">Created</th>
                </tr>
              </thead>
              <tbody>
                {remediations.map((rem) => (
                  <tr key={rem.id} className="border-b border-border">
                    <td className="py-2 font-mono text-xs">{rem.id}</td>
                    <td className="py-2">{rem.type}</td>
                    <td className="py-2">{rem.target_name}</td>
                    <td className="py-2">
                      <StatusBadge status={rem.status} />
                    </td>
                    <td className="py-2">{rem.is_automated ? "Yes" : "No"}</td>
                    <td className="py-2 text-muted-foreground">
                      {formatDate(rem.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="grid grid-cols-3 gap-4">
        <MetricCard label="Total Actions" value={String(remediations?.length || 0)} />
        <MetricCard
          label="Success Rate"
          value={
            remediations?.length
              ? `${Math.round(
                  (remediations.filter((r) => r.status === "success").length /
                    remediations.length) *
                    100
                )}%`
              : "N/A"
          }
        />
        <MetricCard label="Auto-Healing" value="87%" />
      </div>
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-1 text-xl font-bold text-foreground">{value}</div>
    </div>
  );
}
