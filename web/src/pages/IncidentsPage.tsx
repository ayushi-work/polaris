import { useState } from "react";
import { Link } from "react-router-dom";
import { useIncidents, useCreateIncident } from "@/hooks/useIncidents";
import { SeverityBadge, StatusBadge } from "@/components/incidents/SeverityBadge";
import { formatDate } from "@/lib/utils";
import { Plus, AlertTriangle } from "lucide-react";

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

  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Incidents</h1>
        <p className="mt-1 text-sm text-stone-500">Active and resolved incidents across the cluster.</p>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex gap-3">
          <select
            className="rounded-lg border border-border bg-white px-3 py-2 text-sm text-stone-700 focus:outline-none focus:ring-2 focus:ring-stone-200"
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
            className="rounded-lg border border-border bg-white px-3 py-2 text-sm text-stone-700 focus:outline-none focus:ring-2 focus:ring-stone-200"
            value={severity}
            onChange={(e) => setSeverity(e.target.value)}
          >
            <option value="">All Severities</option>
            <option value="critical">Critical</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </select>
          <input
            className="rounded-lg border border-border bg-white px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-200"
            placeholder="Filter by service..."
            value={service}
            onChange={(e) => setService(e.target.value)}
          />
        </div>
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
          className="flex items-center gap-2 rounded-lg bg-stone-800 px-4 py-2 text-sm font-medium text-white hover:bg-stone-700 transition-colors"
        >
          <Plus className="h-4 w-4" />
          New Incident
        </button>
      </div>

      {isLoading ? (
        <div className="text-sm text-stone-400">Loading...</div>
      ) : !incidents?.length ? (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-border bg-white py-16">
          <AlertTriangle className="h-8 w-8 text-stone-300" />
          <p className="text-sm text-stone-400">No incidents found</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border bg-white overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-stone-50 text-left">
                <th className="px-5 py-3 text-xs font-medium uppercase tracking-wider text-stone-400">ID</th>
                <th className="px-5 py-3 text-xs font-medium uppercase tracking-wider text-stone-400">Severity</th>
                <th className="px-5 py-3 text-xs font-medium uppercase tracking-wider text-stone-400">Service</th>
                <th className="px-5 py-3 text-xs font-medium uppercase tracking-wider text-stone-400">Type</th>
                <th className="px-5 py-3 text-xs font-medium uppercase tracking-wider text-stone-400">Status</th>
                <th className="px-5 py-3 text-xs font-medium uppercase tracking-wider text-stone-400">Detected</th>
              </tr>
            </thead>
            <tbody>
              {incidents.map((inc) => (
                <tr key={inc.id} className="border-b border-border hover:bg-stone-50/50 transition-colors">
                  <td className="px-5 py-3">
                    <Link to={`/incidents/${inc.id}`} className="font-mono text-xs text-stone-600 hover:text-stone-900">
                      {inc.id}
                    </Link>
                  </td>
                  <td className="px-5 py-3">
                    <SeverityBadge severity={inc.severity} />
                  </td>
                  <td className="px-5 py-3 font-medium text-stone-700">{inc.resource_name}</td>
                  <td className="px-5 py-3 text-stone-500">{inc.incident_type}</td>
                  <td className="px-5 py-3">
                    <StatusBadge status={inc.status} />
                  </td>
                  <td className="px-5 py-3 text-xs text-stone-400">
                    {formatDate(inc.detected_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
