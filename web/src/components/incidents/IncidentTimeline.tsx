import { formatDate } from "@/lib/utils";
import type { TimelineEntry } from "@/types";

interface Props {
  entries: TimelineEntry[];
}

export function IncidentTimeline({ entries }: Props) {
  return (
    <div className="space-y-4">
      {entries.map((entry, i) => (
        <div key={i} className="flex gap-4">
          <div className="flex flex-col items-center">
            <div className="h-2.5 w-2.5 rounded-full border-2 border-primary bg-background" />
            {i < entries.length - 1 && <div className="mt-1 h-full w-px bg-border" />}
          </div>
          <div className="pb-4">
            <div className="text-sm font-medium text-foreground">{entry.event}</div>
            <div className="text-sm text-muted-foreground">{entry.details}</div>
            <div className="text-xs text-muted-foreground">{formatDate(entry.timestamp)}</div>
          </div>
        </div>
      ))}
    </div>
  );
}
