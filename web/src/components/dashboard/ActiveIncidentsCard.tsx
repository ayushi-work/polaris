import { AlertTriangle } from "lucide-react";

interface Props {
  active: number;
  today: number;
}

export function ActiveIncidentsCard({ active, today }: Props) {
  return (
    <div className="rounded-xl border border-border bg-white p-5">
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-stone-400">
        <AlertTriangle className="h-3.5 w-3.5" />
        Incidents
      </div>
      <div className="mt-3 flex items-baseline gap-2">
        <span className={`font-serif text-2xl italic ${active > 0 ? "text-red-600" : "text-stone-400"}`}>
          {active}
        </span>
        <span className="text-sm text-stone-500">active</span>
        <span className="text-xs text-stone-400">&middot; {today} today</span>
      </div>
    </div>
  );
}
