import { useIncidents } from "@/hooks/useIncidents";
import { formatDate } from "@/lib/utils";
import { SeverityBadge, StatusBadge } from "@/components/incidents/SeverityBadge";

export function RecentActivity() {
  const { data: incidents } = useIncidents({ limit: "5" });

  if (!incidents?.length) {
    return (
      <div className="rounded-lg border border-border bg-card p-4">
        <div className="text-sm text-muted-foreground">Recent Activity</div>
        <div className="mt-4 text-center text-sm text-muted-foreground">No recent incidents</div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-4 text-sm text-muted-foreground">Recent Activity</div>
      <div className="space-y-3">
        {incidents.slice(0, 5).map((inc) => (
          <div key={inc.id} className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <SeverityBadge severity={inc.severity} />
              <span className="text-foreground">{inc.resource_name}</span>
              <span className="text-muted-foreground">{inc.incident_type}</span>
            </div>
            <div className="flex items-center gap-3">
              <StatusBadge status={inc.status} />
              <span className="text-xs text-muted-foreground">{formatDate(inc.detected_at)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
