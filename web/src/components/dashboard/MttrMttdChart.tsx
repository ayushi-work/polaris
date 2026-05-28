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
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-4 text-sm text-muted-foreground">MTTR / MTTD Trend (seconds)</div>
      <ResponsiveContainer width="100%" height={200}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(215 28% 17%)" />
          <XAxis dataKey="name" stroke="hsl(215 20% 65%)" fontSize={12} />
          <YAxis stroke="hsl(215 20% 65%)" fontSize={12} />
          <Tooltip
            contentStyle={{
              backgroundColor: "hsl(215 28% 8%)",
              border: "1px solid hsl(215 28% 17%)",
              borderRadius: "0.5rem",
            }}
          />
          <Bar dataKey="mttr" fill="hsl(217 91% 60%)" name="MTTR" radius={[4, 4, 0, 0]} />
          <Bar dataKey="mttd" fill="hsl(160 84% 39%)" name="MTTD" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
