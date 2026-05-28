import { useState } from "react";
import { Zap, Loader2 } from "lucide-react";
import { useChaosScenarios, useExecuteScenario, useCreateScenario } from "@/hooks/useChaosScenarios";
import { formatDate } from "@/lib/utils";

const FAILURE_TYPES = [
  { value: "delete_pod", label: "Delete Pod" },
  { value: "stress_cpu", label: "CPU Stress" },
  { value: "stress_memory", label: "Memory Stress" },
  { value: "network_delay", label: "Network Delay" },
  { value: "network_loss", label: "Network Loss" },
  { value: "config_fault", label: "Config Fault" },
];

const DURATIONS = [
  { value: "1m", label: "1 min" },
  { value: "5m", label: "5 min" },
  { value: "10m", label: "10 min" },
  { value: "30m", label: "30 min" },
];

export function ChaosLab() {
  const [target, setTarget] = useState("payment-service");
  const [action, setAction] = useState("delete_pod");
  const [duration, setDuration] = useState("5m");
  const [description, setDescription] = useState("");

  const { data: scenarios } = useChaosScenarios();
  const execute = useExecuteScenario();
  const create = useCreateScenario();

  const handleLaunch = () => {
    create.mutate({
      name: target,
      description: description || `${action} on ${target} for ${duration}`,
      action,
      parameters: JSON.stringify({ duration }),
      enabled: true,
    });
  };

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-white p-6">
        <h2 className="mb-5 flex items-center gap-2 font-serif text-xl italic text-stone-900">
          <Zap className="h-5 w-5 text-amber-500" />
          Inject Failure
        </h2>
        <div className="grid grid-cols-4 gap-4">
          <div>
            <label className="mb-1.5 block text-xs font-medium uppercase tracking-wider text-stone-400">Target</label>
            <input
              className="w-full rounded-lg border border-border bg-stone-50 px-3 py-2.5 text-sm text-stone-800 placeholder:text-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-200 focus:bg-white transition-colors"
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              placeholder="payment-service"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium uppercase tracking-wider text-stone-400">Type</label>
            <select
              className="w-full rounded-lg border border-border bg-stone-50 px-3 py-2.5 text-sm text-stone-800 focus:outline-none focus:ring-2 focus:ring-stone-200 transition-colors"
              value={action}
              onChange={(e) => setAction(e.target.value)}
            >
              {FAILURE_TYPES.map((f) => (
                <option key={f.value} value={f.value}>{f.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium uppercase tracking-wider text-stone-400">Duration</label>
            <select
              className="w-full rounded-lg border border-border bg-stone-50 px-3 py-2.5 text-sm text-stone-800 focus:outline-none focus:ring-2 focus:ring-stone-200 transition-colors"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
            >
              {DURATIONS.map((d) => (
                <option key={d.value} value={d.value}>{d.label}</option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <button
              onClick={handleLaunch}
              disabled={create.isPending}
              className="flex w-full items-center justify-center gap-2 rounded-lg bg-red-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50 transition-colors"
            >
              {create.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Zap className="h-4 w-4" />}
              Launch
            </button>
          </div>
        </div>
        <div className="mt-4">
          <label className="mb-1.5 block text-xs font-medium uppercase tracking-wider text-stone-400">Description</label>
          <input
            className="w-full rounded-lg border border-border bg-stone-50 px-3 py-2.5 text-sm text-stone-800 placeholder:text-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-200 focus:bg-white transition-colors"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Optional description..."
          />
        </div>
      </div>

      <div className="rounded-xl border border-border bg-white p-6">
        <h2 className="mb-4 font-serif text-xl italic text-stone-900">Scenario History</h2>
        {!scenarios?.length ? (
          <p className="text-sm text-stone-400 py-8 text-center">No scenarios yet. Create one above.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left">
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Name</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Action</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Runs</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400">Last Run</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-stone-400"></th>
              </tr>
            </thead>
            <tbody>
              {scenarios.map((sc) => (
                <tr key={sc.id} className="border-b border-border">
                  <td className="py-2.5 font-medium text-stone-700">{sc.name}</td>
                  <td className="py-2.5 text-stone-500">{sc.action}</td>
                  <td className="py-2.5 text-stone-500">{sc.run_count}</td>
                  <td className="py-2.5 text-xs text-stone-400">
                    {sc.last_run_at ? formatDate(sc.last_run_at) : "Never"}
                  </td>
                  <td className="py-2.5">
                    <button
                      onClick={() => execute.mutate(sc.id)}
                      className="rounded-lg bg-stone-100 px-3 py-1 text-xs font-medium text-stone-600 hover:bg-stone-200 transition-colors"
                    >
                      Re-run
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
