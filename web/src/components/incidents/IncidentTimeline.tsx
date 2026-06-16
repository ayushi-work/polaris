import { formatDate } from "@/lib/utils";
import type { TimelineEntry } from "@/types";
import { Rocket, Box, AlertTriangle, Skull, Zap, Bell, Wrench, CheckCircle, XCircle, Loader, Brain, FileText } from "lucide-react";

interface Props {
  entries: TimelineEntry[];
}

const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
  rocket: Rocket,
  container: Box,
  "alert-triangle": AlertTriangle,
  skull: Skull,
  zap: Zap,
  bell: Bell,
  wrench: Wrench,
  "check-circle": CheckCircle,
  "x-circle": XCircle,
  loader: Loader,
  brain: Brain,
  "file-text": FileText,
};

const sourceColor: Record<string, string> = {
  kubernetes: "border-blue-400 bg-blue-400",
  prometheus: "border-amber-400 bg-amber-400",
  node: "border-red-400 bg-red-400",
  polaris: "border-stone-700 bg-white",
  remediation: "border-emerald-400 bg-emerald-400",
  llm: "border-purple-400 bg-purple-400",
};

function timeAgo(dateStr: string, relativeTo?: string): string {
  const d = new Date(dateStr);
  const ref = relativeTo ? new Date(relativeTo) : new Date();
  const diffMs = d.getTime() - ref.getTime();
  const mins = Math.round(diffMs / 60000);
  if (mins === 0) return "now";
  if (mins > 0) return `+${mins}m`;
  return `${mins}m`;
}

export function IncidentTimeline({ entries }: Props) {
  if (!entries.length) return null;

  const firstTs = entries[0].timestamp;

  return (
    <div className="space-y-0">
      {entries.map((entry, i) => {
        const Icon = iconMap[entry.icon] || Zap;
        const dotColor = sourceColor[entry.source] || sourceColor.polaris;
        const isLast = i === entries.length - 1;

        return (
          <div key={i} className="flex gap-3">
            <div className="flex flex-col items-center pt-0.5">
              <div className={`flex h-5 w-5 items-center justify-center rounded-full border-2 ${dotColor}`}>
                <Icon className="h-2.5 w-2.5 text-white" />
              </div>
              {!isLast && <div className="mt-0.5 w-px flex-1 bg-stone-200" />}
            </div>
            <div className={`${isLast ? "pb-0" : "pb-5"}`}>
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-stone-700">{entry.event}</span>
                <span className="rounded bg-stone-100 px-1.5 py-0.5 text-[10px] font-medium uppercase text-stone-400">
                  {entry.source}
                </span>
              </div>
              <div className="text-sm text-stone-500">{entry.details}</div>
              <div className="mt-0.5 flex items-center gap-2 text-xs text-stone-400">
                <span>{formatDate(entry.timestamp)}</span>
                <span className="text-stone-300">·</span>
                <span className="font-mono">{timeAgo(entry.timestamp, firstTs)}</span>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
