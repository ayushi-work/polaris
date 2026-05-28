import { useState } from "react";
import { ScrollText, Search } from "lucide-react";

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
      <div className="flex gap-4">
        <select className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground">
          <option>default</option>
          <option>kube-system</option>
          <option>monitoring</option>
        </select>
        <input
          className="flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
          placeholder="Filter logs..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
        <button className="flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm text-white">
          <Search className="h-4 w-4" />
          Search
        </button>
      </div>

      <div className="rounded-lg border border-border bg-gray-950 p-4 font-mono text-xs">
        {filteredLogs.map((log, i) => {
          const isError = log.includes("ERROR");
          const isWarn = log.includes("WARN");
          return (
            <div
              key={i}
              className={`py-0.5 ${
                isError
                  ? "text-red-400"
                  : isWarn
                    ? "text-amber-400"
                    : "text-gray-400"
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
