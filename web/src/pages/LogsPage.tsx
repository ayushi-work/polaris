import { LiveLogs } from "@/components/logs/LiveLogs";

export default function LogsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Live Logs</h1>
        <p className="mt-1 text-sm text-stone-500">Real-time log streaming from cluster pods.</p>
      </div>
      <LiveLogs />
    </div>
  );
}
