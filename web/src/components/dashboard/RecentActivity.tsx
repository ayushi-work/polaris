import { useIncidents } from "@/hooks/useIncidents";
import { formatDate } from "@/lib/utils";
import { SeverityBadge, StatusBadge } from "@/components/incidents/SeverityBadge";

export function RecentActivity() {
  const { data: incidents } = useIncidents({ limit: "5" });

  return (
    <div className="rounded-xl border border-border bg-white p-5">
      <div className="mb-4 text-xs font-medium uppercase tracking-wider text-stone-400">
        Recent Activity
      </div>
      {!incidents?.length ? (
        <p className="py-8 text-center text-sm text-stone-400">No recent incidents</p>
      ) : (
        <div className="space-y-1">
          {incidents.slice(0, 5).map((inc) => (
            <div
              key={inc.id}
              className="flex items-center justify-between rounded-lg px-3 py-2.5 hover:bg-stone-50 transition-colors"
            >
              <div className="flex items-center gap-3 text-sm">
                <SeverityBadge severity={inc.severity} />
                <span className="text-stone-700 font-medium">{inc.resource_name}</span>
                <span className="text-stone-400">{inc.incident_type}</span>
              </div>
              <div className="flex items-center gap-3">
                <StatusBadge status={inc.status} />
                <span className="text-xs text-stone-400">{formatDate(inc.detected_at)}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
