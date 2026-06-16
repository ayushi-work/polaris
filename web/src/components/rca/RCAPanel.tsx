import { Lightbulb, AlertCircle, ListChecks, FileText, Activity } from "lucide-react";
import type { RCAResult, Evidence } from "@/types";

interface Props {
  rca: RCAResult;
}

function parseEvidence(jsonStr: string): string[] {
  try {
    const parsed = JSON.parse(jsonStr);
    if (Array.isArray(parsed)) return parsed;
  } catch {}
  return jsonStr ? [jsonStr] : [];
}

export function RCAPanel({ rca }: Props) {
  const confPct = Math.round(rca.confidence * 100);
  const citedLogs = parseEvidence(rca.logs_snippet);
  const citedEvents = parseEvidence(rca.events_summary);
  const hasEvidence = citedLogs.length > 0 || citedEvents.length > 0;

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-border bg-white p-5">
        <div className="flex items-center gap-2 mb-3">
          <Lightbulb className="h-4 w-4 text-amber-500" />
          <span className="text-xs font-medium uppercase tracking-wider text-stone-400">Root Cause</span>
          <span className={`ml-auto rounded-full px-2 py-0.5 text-xs font-medium ${
            confPct >= 70 ? "bg-emerald-50 text-emerald-700" :
            confPct >= 40 ? "bg-amber-50 text-amber-700" :
            "bg-red-50 text-red-700"
          }`}>
            {confPct}% confidence
          </span>
        </div>
        <p className="text-sm text-stone-600 leading-relaxed">{rca.root_cause}</p>
      </div>

      {hasEvidence && (
        <div className="rounded-xl border border-border bg-white p-5">
          <div className="flex items-center gap-2 mb-4">
            <AlertCircle className="h-4 w-4 text-blue-500" />
            <span className="text-xs font-medium uppercase tracking-wider text-stone-400">Evidence cited by LLM</span>
          </div>

          {citedLogs.length > 0 && (
            <div className="mb-4">
              <div className="flex items-center gap-1.5 mb-2">
                <FileText className="h-3.5 w-3.5 text-stone-400" />
                <span className="text-xs font-medium text-stone-500">Log citations</span>
              </div>
              <div className="space-y-1">
                {citedLogs.map((line, i) => (
                  <div key={i} className="flex gap-2 rounded-md bg-red-50/50 px-3 py-1.5 text-xs font-mono text-red-800 border border-red-100">
                    <span className="text-red-400 select-none">{i + 1}.</span>
                    <span className="truncate">{line}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {citedEvents.length > 0 && (
            <div>
              <div className="flex items-center gap-1.5 mb-2">
                <Activity className="h-3.5 w-3.5 text-stone-400" />
                <span className="text-xs font-medium text-stone-500">Event citations</span>
              </div>
              <div className="space-y-1">
                {citedEvents.map((line, i) => (
                  <div key={i} className="flex gap-2 rounded-md bg-amber-50/50 px-3 py-1.5 text-xs font-mono text-amber-800 border border-amber-100">
                    <span className="text-amber-400 select-none">{i + 1}.</span>
                    <span className="truncate">{line}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      <div className="rounded-xl border border-border bg-white p-5">
        <div className="flex items-center gap-2 mb-3">
          <ListChecks className="h-4 w-4 text-emerald-500" />
          <span className="text-xs font-medium uppercase tracking-wider text-stone-400">Recommendation</span>
        </div>
        <p className="text-sm text-stone-600 leading-relaxed">{rca.summary}</p>
        {rca.suggested_actions && (
          <div className="mt-3 flex gap-2">
            {rca.suggested_actions.split(",").map((action) => (
              <span key={action} className="rounded-lg bg-stone-100 px-3 py-1 text-xs font-medium text-stone-700">
                {action.trim()}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
