import { AlertTriangle } from "lucide-react";

interface Props {
  active: number;
  today: number;
}

export function ActiveIncidentsCard({ active, today }: Props) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <AlertTriangle className="h-4 w-4" />
        Incidents
      </div>
      <div className="mt-2 flex items-baseline gap-3">
        <span className="text-2xl font-bold text-red-400">{active}</span>
        <span className="text-sm text-muted-foreground">active</span>
        <span className="text-sm text-muted-foreground">| {today} today</span>
      </div>
    </div>
  );
}
