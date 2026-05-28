import { formatDate } from "@/lib/utils";
import type { TimelineEntry } from "@/types";

interface Props {
  entries: TimelineEntry[];
}

export function IncidentTimeline({ entries }: Props) {
  return (
    <div className="space-y-0">
      {entries.map((entry, i) => (
        <div key={i} className="flex gap-4">
          <div className="flex flex-col items-center pt-1">
            <div className={`h-2.5 w-2.5 rounded-full border-2 ${
              i === 0 ? "border-stone-800 bg-white" :
              i === entries.length - 1 ? "border-emerald-500 bg-emerald-500" :
              "border-stone-300 bg-stone-300"
            }`} />
            {i < entries.length - 1 && <div className="mt-1 w-px flex-1 bg-stone-200" />}
          </div>
          <div className={`pb-5 ${i === entries.length - 1 ? "pb-0" : ""}`}>
            <div className="text-sm font-medium text-stone-700">{entry.event}</div>
            <div className="text-sm text-stone-500">{entry.details}</div>
            <div className="mt-0.5 text-xs text-stone-400">{formatDate(entry.timestamp)}</div>
          </div>
        </div>
      ))}
    </div>
  );
}
