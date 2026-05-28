import { Lightbulb, AlertCircle, ListChecks } from "lucide-react";
import type { RCAResult } from "@/types";

interface Props {
  rca: RCAResult;
}

export function RCAPanel({ rca }: Props) {
  const confPct = Math.round(rca.confidence * 100);

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

      <div className="rounded-xl border border-border bg-white p-5">
        <div className="flex items-center gap-2 mb-3">
          <AlertCircle className="h-4 w-4 text-blue-500" />
          <span className="text-xs font-medium uppercase tracking-wider text-stone-400">Evidence</span>
        </div>
        {rca.logs_snippet && (
          <div className="mb-3">
            <span className="text-xs text-stone-400">Logs</span>
            <pre className="mt-1 max-h-32 overflow-y-auto rounded-lg bg-stone-50 p-3 text-xs font-mono text-stone-600 leading-relaxed">
              {rca.logs_snippet.slice(0, 500)}
            </pre>
          </div>
        )}
        {rca.events_summary && (
          <div>
            <span className="text-xs text-stone-400">Events</span>
            <pre className="mt-1 max-h-32 overflow-y-auto rounded-lg bg-stone-50 p-3 text-xs font-mono text-stone-600 leading-relaxed">
              {rca.events_summary.slice(0, 500)}
            </pre>
          </div>
        )}
      </div>

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
