import { useParams } from "react-router-dom";
import { useIncident, useAcknowledgeIncident, useResolveIncident } from "@/hooks/useIncidents";
import { SeverityBadge, StatusBadge } from "@/components/incidents/SeverityBadge";
import { IncidentTimeline } from "@/components/incidents/IncidentTimeline";
import { RCAPanel } from "@/components/rca/RCAPanel";
import { formatDate } from "@/lib/utils";
import { AlertTriangle } from "lucide-react";

export default function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: incident, isLoading } = useIncident(id || "");
  const acknowledge = useAcknowledgeIncident();
  const resolve = useResolveIncident();

  if (isLoading) {
    return <div className="text-sm text-stone-400">Loading...</div>;
  }

  if (!incident) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-stone-400">
        <AlertTriangle className="h-8 w-8 text-stone-300" />
        <p className="text-sm">Incident not found</p>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="font-serif text-3xl italic text-stone-900">{incident.id}</h1>
          <p className="mt-1 text-sm text-stone-500">
            Detected {formatDate(incident.detected_at)}
          </p>
        </div>
        <div className="flex gap-2">
          {incident.status === "detected" && (
            <button
              onClick={() => acknowledge.mutate(incident.id)}
              className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-2 text-sm font-medium text-amber-700 hover:bg-amber-100 transition-colors"
            >
              Acknowledge
            </button>
          )}
          {["detected", "investigating", "remediating"].includes(incident.status) && (
            <button
              onClick={() => resolve.mutate(incident.id)}
              className="rounded-lg bg-stone-800 px-4 py-2 text-sm font-medium text-white hover:bg-stone-700 transition-colors"
            >
              Resolve
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <div className="rounded-xl border border-border bg-white p-4">
          <div className="text-xs font-medium uppercase tracking-wider text-stone-400">Severity</div>
          <div className="mt-2"><SeverityBadge severity={incident.severity} /></div>
        </div>
        <div className="rounded-xl border border-border bg-white p-4">
          <div className="text-xs font-medium uppercase tracking-wider text-stone-400">Status</div>
          <div className="mt-2"><StatusBadge status={incident.status} /></div>
        </div>
        <div className="rounded-xl border border-border bg-white p-4">
          <div className="text-xs font-medium uppercase tracking-wider text-stone-400">Service</div>
          <p className="mt-2 text-sm font-medium text-stone-800">{incident.resource_name}</p>
        </div>
        <div className="rounded-xl border border-border bg-white p-4">
          <div className="text-xs font-medium uppercase tracking-wider text-stone-400">Type</div>
          <p className="mt-2 text-sm font-medium text-stone-800">{incident.incident_type}</p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-white p-5">
        <h2 className="mb-2 text-xs font-medium uppercase tracking-wider text-stone-400">Message</h2>
        <p className="text-sm text-stone-600">{incident.message}</p>
      </div>

      <div className="rounded-xl border border-border bg-white p-5">
        <h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-stone-400">Timeline</h2>
        <IncidentTimeline entries={[
          { timestamp: incident.detected_at, event: "Incident Detected", details: incident.message },
          ...(incident.rca_result ? [{ timestamp: incident.rca_result.created_at, event: "RCA Completed", details: incident.rca_result.summary }] : []),
          ...(incident.remediations?.map(r => ({ timestamp: r.executed_at || r.created_at, event: `Remediation ${r.status}`, details: `${r.type} on ${r.target_name}` })) || []),
          ...(incident.resolved_at ? [{ timestamp: incident.resolved_at, event: "Incident Resolved", details: "Service restored" }] : []),
        ]} />
      </div>

      {incident.rca_result ? (
        <RCAPanel rca={incident.rca_result} />
      ) : (
        <div className="rounded-xl border border-border bg-white p-5 text-sm text-stone-400">
          No RCA analysis available yet.
        </div>
      )}

      {incident.remediations && incident.remediations.length > 0 && (
        <div className="rounded-xl border border-border bg-white p-5">
          <h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-stone-400">Remediations</h2>
          <div className="space-y-2">
            {incident.remediations.map((rem) => (
              <div key={rem.id} className="flex items-center justify-between rounded-lg bg-stone-50 px-4 py-2.5 text-sm">
                <span className="font-medium text-stone-700">{rem.type}</span>
                <StatusBadge status={rem.status} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
