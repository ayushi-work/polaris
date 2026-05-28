import { useState } from "react";
import { Link } from "react-router-dom";
import { useIncidents } from "@/hooks/useIncidents";
import { SeverityBadge, StatusBadge } from "@/components/incidents/SeverityBadge";
import { formatDate } from "@/lib/utils";
import { Plus, AlertTriangle } from "lucide-react";
import { useCreateIncident } from "@/hooks/useIncidents";

export default function IncidentsPage() {
  const [status, setStatus] = useState("");
  const [severity, setSeverity] = useState("");
  const [service, setService] = useState("");

  const params: Record<string, string> = {};
  if (status) params.status = status;
  if (severity) params.severity = severity;
  if (service) params.service = service;

  const { data: incidents, isLoading } = useIncidents(params);
  const createIncident = useCreateIncident();

  if (isLoading) {
    return <div className="text-sm text-muted-foreground">Loading incidents...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">Incidents</h1>
        <button
          onClick={() =>
            createIncident.mutate({
              cluster: "local",
              namespace: "default",
              kind: "Pod",
              resource_name: "test-pod",
              incident_type: "CrashLoopBackOff",
              severity: "critical",
              message: "Manually created test incident",
              status: "detected",
            })
          }
          className="flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm text-white hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          Create Test Incident
        </button>
      </div>

      <div className="flex gap-4">
        <select
          className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">All Statuses</option>
          <option value="detected">Detected</option>
          <option value="investigating">Investigating</option>
          <option value="remediating">Remediating</option>
          <option value="resolved">Resolved</option>
        </select>
        <select
          className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
          value={severity}
          onChange={(e) => setSeverity(e.target.value)}
        >
          <option value="">All Severities</option>
          <option value="critical">Critical</option>
          <option value="warning">Warning</option>
          <option value="info">Info</option>
        </select>
        <input
          className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
          placeholder="Filter by service..."
          value={service}
          onChange={(e) => setService(e.target.value)}
        />
      </div>

      <div className="rounded-lg border border-border bg-card">
        {!incidents?.length ? (
          <div className="flex flex-col items-center gap-2 py-12 text-muted-foreground">
            <AlertTriangle className="h-8 w-8" />
            <p>No incidents found</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left">
                <th className="px-4 py-3 font-medium text-muted-foreground">ID</th>
                <th className="px-4 py-3 font-medium text-muted-foreground">Severity</th>
                <th className="px-4 py-3 font-medium text-muted-foreground">Service</th>
                <th className="px-4 py-3 font-medium text-muted-foreground">Type</th>
                <th className="px-4 py-3 font-medium text-muted-foreground">Status</th>
                <th className="px-4 py-3 font-medium text-muted-foreground">Detected</th>
              </tr>
            </thead>
            <tbody>
              {incidents.map((inc) => (
                <tr key={inc.id} className="border-b border-border hover:bg-accent/50">
                  <td className="px-4 py-3">
                    <Link to={`/incidents/${inc.id}`} className="font-mono text-primary hover:underline">
                      {inc.id}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <SeverityBadge severity={inc.severity} />
                  </td>
                  <td className="px-4 py-3">{inc.resource_name}</td>
                  <td className="px-4 py-3">{inc.incident_type}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={inc.status} />
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {formatDate(inc.detected_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
