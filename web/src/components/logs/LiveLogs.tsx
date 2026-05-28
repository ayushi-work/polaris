import { useState } from "react";

const SAMPLE_LOGS = [
  "12:03:01 INFO  Starting worker process",
  "12:03:15 INFO  Processing request from frontend",
  "12:03:42 WARN  Connection pool at 75% capacity",
  "12:04:02 ERROR OutOfMemoryError: Java heap space",
  "12:04:02 ERROR Container terminated with exit code 137",
  "12:05:01 INFO  Restarting container payment-7f8g9",
  "12:05:15 INFO  Container started successfully",
  "12:05:30 INFO  Health check passed",
  "12:06:00 INFO  Restored service to healthy state",
];

export function LiveLogs() {
  const [filter, setFilter] = useState("");

  const filteredLogs = SAMPLE_LOGS.filter((log) =>
    log.toLowerCase().includes(filter.toLowerCase())
  );

  return (
    <div className="space-y-4">
      <div className="flex gap-3">
        <select className="rounded-lg border border-border bg-white px-3 py-2 text-sm text-stone-700 focus:outline-none focus:ring-2 focus:ring-stone-200">
          <option>default</option>
          <option>kube-system</option>
          <option>monitoring</option>
        </select>
        <input
          className="flex-1 rounded-lg border border-border bg-white px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-200"
          placeholder="Filter logs..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      <div className="rounded-xl border border-border bg-stone-50 p-5 font-mono text-xs leading-relaxed">
        {filteredLogs.map((log, i) => {
          const isError = log.includes("ERROR");
          const isWarn = log.includes("WARN");
          return (
            <div
              key={i}
              className={`py-0.5 ${
                isError ? "text-red-600" : isWarn ? "text-amber-600" : "text-stone-600"
              }`}
            >
              {log}
            </div>
          );
        })}
      </div>
    </div>
  );
}
