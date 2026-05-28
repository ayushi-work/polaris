import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
} from "recharts";

const data = [
  { name: "Mon", mttr: 240, mttd: 30 },
  { name: "Tue", mttr: 180, mttd: 25 },
  { name: "Wed", mttr: 320, mttd: 20 },
  { name: "Thu", mttr: 200, mttd: 18 },
  { name: "Fri", mttr: 150, mttd: 15 },
  { name: "Sat", mttr: 100, mttd: 12 },
  { name: "Sun", mttr: 90, mttd: 10 },
];

export function MttrMttdChart() {
  return (
    <div className="rounded-xl border border-border bg-white p-5">
      <div className="mb-1 text-xs font-medium uppercase tracking-wider text-stone-400">
        MTTR / MTTD Trend
      </div>
      <p className="mb-4 text-xs text-stone-400">Seconds, last 7 days</p>
      <ResponsiveContainer width="100%" height={200}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(30 10% 90%)" vertical={false} />
          <XAxis dataKey="name" stroke="hsl(30 5% 60%)" fontSize={12} tickLine={false} axisLine={false} />
          <YAxis stroke="hsl(30 5% 60%)" fontSize={12} tickLine={false} axisLine={false} />
          <Tooltip
            contentStyle={{
              backgroundColor: "#fff",
              border: "1px solid hsl(30 10% 88%)",
              borderRadius: "0.625rem",
              boxShadow: "0 4px 12px rgba(0,0,0,0.06)",
              fontSize: "13px",
            }}
          />
          <Bar dataKey="mttr" fill="hsl(220 15% 30%)" name="MTTR" radius={[4, 4, 0, 0]} />
          <Bar dataKey="mttd" fill="hsl(220 15% 55%)" name="MTTD" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
