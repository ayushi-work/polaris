import { Lightbulb, AlertCircle, ListChecks, Brain } from "lucide-react";
import type { RCAResult } from "@/types";
import { cn } from "@/lib/utils";

interface Props {
  rca: RCAResult;
}

export function RCAPanel({ rca }: Props) {
  const confPct = Math.round(rca.confidence * 100);

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-border bg-card p-4">
        <div className="flex items-center gap-2 mb-3">
          <Lightbulb className="h-4 w-4 text-amber-400" />
          <span className="font-medium text-sm">Root Cause</span>
          <span
            className={cn(
              "ml-auto rounded-full px-2 py-0.5 text-xs",
              confPct >= 70
                ? "bg-emerald-500/10 text-emerald-400"
                : confPct >= 40
                  ? "bg-amber-500/10 text-amber-400"
                  : "bg-red-500/10 text-red-400"
            )}
          >
            {confPct}% confidence
          </span>
        </div>
        <p className="text-sm text-muted-foreground">{rca.root_cause}</p>
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="flex items-center gap-2 mb-3">
          <AlertCircle className="h-4 w-4 text-blue-400" />
          <span className="font-medium text-sm">Evidence</span>
        </div>
        {rca.logs_snippet && (
          <div className="mb-2">
            <span className="text-xs text-muted-foreground">Logs:</span>
            <pre className="mt-1 max-h-32 overflow-y-auto rounded bg-gray-950 p-2 text-xs font-mono text-gray-400">
              {rca.logs_snippet.slice(0, 500)}
            </pre>
          </div>
        )}
        {rca.events_summary && (
          <div>
            <span className="text-xs text-muted-foreground">Events:</span>
            <pre className="mt-1 max-h-32 overflow-y-auto rounded bg-gray-950 p-2 text-xs font-mono text-gray-400">
              {rca.events_summary.slice(0, 500)}
            </pre>
          </div>
        )}
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="flex items-center gap-2 mb-3">
          <ListChecks className="h-4 w-4 text-emerald-400" />
          <span className="font-medium text-sm">Recommendation</span>
        </div>
        <p className="text-sm text-muted-foreground">{rca.summary}</p>
        {rca.suggested_actions && (
          <div className="mt-2 flex gap-2">
            {rca.suggested_actions.split(",").map((action) => (
              <span
                key={action}
                className="rounded bg-primary/10 px-2 py-0.5 text-xs text-primary"
              >
                {action.trim()}
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="flex items-center gap-2 mb-1">
          <Brain className="h-4 w-4 text-purple-400" />
          <span className="text-xs text-muted-foreground">
            Model: {rca.llm_model} | Completion tokens: {rca.completion_tokens}
          </span>
        </div>
      </div>
    </div>
  );
}
