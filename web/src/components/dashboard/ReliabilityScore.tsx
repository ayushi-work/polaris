interface Props {
  score: number;
  mttr: string;
  mttd: string;
}

export function ReliabilityScore({ score, mttr, mttd }: Props) {
  const color =
    score >= 90
      ? "text-emerald-600"
      : score >= 70
        ? "text-amber-600"
        : "text-red-600";

  return (
    <div className="rounded-xl border border-border bg-white p-5">
      <div className="text-xs font-medium uppercase tracking-wider text-stone-400">
        Reliability Score
      </div>
      <div className={`mt-3 font-serif text-2xl italic ${color}`}>{score}/100</div>
      <div className="mt-3 flex gap-5 text-xs text-stone-400">
        <span>
          MTTR: <span className="text-stone-600 font-medium">{mttr}</span>
        </span>
        <span>
          MTTD: <span className="text-stone-600 font-medium">{mttd}</span>
        </span>
      </div>
    </div>
  );
}
