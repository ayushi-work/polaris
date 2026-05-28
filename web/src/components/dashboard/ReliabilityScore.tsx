interface Props {
  score: number;
  mttr: string;
  mttd: string;
}

export function ReliabilityScore({ score, mttr, mttd }: Props) {
  const color = score >= 90 ? "text-emerald-400" : score >= 70 ? "text-amber-400" : "text-red-400";

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="text-sm text-muted-foreground">Reliability Score</div>
      <div className={`mt-2 text-2xl font-bold ${color}`}>{score}/100</div>
      <div className="mt-3 flex gap-4 text-xs text-muted-foreground">
        <span>MTTR: {mttr}</span>
        <span>MTTD: {mttd}</span>
      </div>
    </div>
  );
}
