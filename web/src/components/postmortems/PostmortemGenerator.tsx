import { useState } from "react";
import { useIncidents } from "@/hooks/useIncidents";
import { FileText, Download } from "lucide-react";

export function PostmortemGenerator() {
  const [selectedId, setSelectedId] = useState("");
  const { data: incidents } = useIncidents({ limit: "50" });

  const incident = incidents?.find((i) => i.id === selectedId);

  const handleDownload = (format: string) => {
    if (!incident) return;

    let content = "";
    const ext = format === "md" ? "md" : format === "html" ? "html" : "txt";

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
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="mb-4 flex items-center gap-2 text-lg font-semibold">
          <FileText className="h-5 w-5" />
          Generate Postmortem
        </h2>
        <div className="flex gap-4">
          <select
            className="flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
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
            className="flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm text-white disabled:opacity-50"
          >
            <Download className="h-4 w-4" />
            Export Markdown
          </button>
        </div>
      </div>

      {incident && (
        <div className="rounded-lg border border-border bg-card p-6">
          <h3 className="mb-4 font-semibold">Preview</h3>
          <div className="space-y-4 text-sm">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-muted-foreground">Incident:</span> {incident.id}
              </div>
              <div>
                <span className="text-muted-foreground">Service:</span> {incident.resource_name}
              </div>
              <div>
                <span className="text-muted-foreground">Severity:</span> {incident.severity}
              </div>
              <div>
                <span className="text-muted-foreground">Status:</span> {incident.status}
              </div>
              <div>
                <span className="text-muted-foreground">Detected:</span>{" "}
                {new Date(incident.detected_at).toLocaleString()}
              </div>
              <div>
                <span className="text-muted-foreground">Resolved:</span>{" "}
                {incident.resolved_at
                  ? new Date(incident.resolved_at).toLocaleString()
                  : "Pending"}
              </div>
            </div>
            <div className="rounded-lg bg-gray-950 p-4">
              <h4 className="mb-2 font-medium">Root Cause Analysis</h4>
              <p className="text-muted-foreground">
                {incident.rca_result?.root_cause || "No RCA available"}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
