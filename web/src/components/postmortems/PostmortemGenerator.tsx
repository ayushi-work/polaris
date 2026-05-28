import { useState } from "react";
import { useIncidents } from "@/hooks/useIncidents";
import { FileText, Download } from "lucide-react";

export function PostmortemGenerator() {
  const [selectedId, setSelectedId] = useState("");
  const { data: incidents } = useIncidents({ limit: "50" });

  const incident = incidents?.find((i) => i.id === selectedId);

  const handleDownload = (format: string) => {
    if (!incident) return;

    const ext = format === "md" ? "md" : "txt";
    let content = "";

    if (ext === "md") {
      content = `# Postmortem: ${incident.id}

**Generated:** ${new Date().toISOString()}

## Overview
- **Service:** ${incident.resource_name}
- **Severity:** ${incident.severity}
- **Type:** ${incident.incident_type}
- **Status:** ${incident.status}

## Root Cause
${incident.rca_result?.root_cause || "No RCA available"}

## Timeline
- **Detected:** ${new Date(incident.detected_at).toISOString()}
${incident.resolved_at ? `- **Resolved:** ${new Date(incident.resolved_at).toISOString()}` : "- Not yet resolved"}

## Recovery Actions
${incident.remediations?.map((r) => `- [${r.status}] ${r.type} on ${r.target_name}`).join("\n") || "No remediations"}

## Summary
${incident.rca_result?.summary || incident.message}
`;
    } else {
      content = `Postmortem: ${incident.id}\nService: ${incident.resource_name}\nRoot Cause: ${incident.rca_result?.root_cause || "N/A"}`;
    }

    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `postmortem-${incident.id}.${ext}`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-white p-6">
        <h2 className="mb-4 flex items-center gap-2 font-serif text-xl italic text-stone-900">
          <FileText className="h-5 w-5" />
          Generate Postmortem
        </h2>
        <div className="flex gap-3">
          <select
            className="flex-1 rounded-lg border border-border bg-stone-50 px-3 py-2.5 text-sm text-stone-700 focus:outline-none focus:ring-2 focus:ring-stone-200"
            value={selectedId}
            onChange={(e) => setSelectedId(e.target.value)}
          >
            <option value="">Select an incident...</option>
            {incidents?.map((inc) => (
              <option key={inc.id} value={inc.id}>
                {inc.id} — {inc.resource_name} ({inc.incident_type})
              </option>
            ))}
          </select>
          <button
            disabled={!incident}
            onClick={() => handleDownload("md")}
            className="flex items-center gap-2 rounded-lg bg-stone-800 px-4 py-2.5 text-sm font-medium text-white hover:bg-stone-700 disabled:opacity-50 transition-colors"
          >
            <Download className="h-4 w-4" />
            Export Markdown
          </button>
        </div>
      </div>

      {incident && (
        <div className="rounded-xl border border-border bg-white p-6">
          <h3 className="mb-4 font-serif text-lg italic text-stone-900">Preview</h3>
          <div className="space-y-4 text-sm">
            <div className="grid grid-cols-2 gap-4">
              <div><span className="text-stone-400">Incident:</span> <span className="text-stone-700">{incident.id}</span></div>
              <div><span className="text-stone-400">Service:</span> <span className="text-stone-700">{incident.resource_name}</span></div>
              <div><span className="text-stone-400">Severity:</span> <span className="text-stone-700">{incident.severity}</span></div>
              <div><span className="text-stone-400">Status:</span> <span className="text-stone-700">{incident.status}</span></div>
              <div><span className="text-stone-400">Detected:</span> <span className="text-stone-700">{new Date(incident.detected_at).toLocaleString()}</span></div>
              <div><span className="text-stone-400">Resolved:</span> <span className="text-stone-700">{incident.resolved_at ? new Date(incident.resolved_at).toLocaleString() : "Pending"}</span></div>
            </div>
            <div className="rounded-lg bg-stone-50 p-4">
              <h4 className="mb-2 text-xs font-medium uppercase tracking-wider text-stone-400">Root Cause Analysis</h4>
              <p className="text-stone-600">{incident.rca_result?.root_cause || "No RCA available"}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
