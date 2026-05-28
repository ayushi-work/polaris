import { useState } from "react";
import { Zap, Loader2 } from "lucide-react";
import { useChaosScenarios, useExecuteScenario, useCreateScenario } from "@/hooks/useChaosScenarios";
import { StatusBadge } from "@/components/incidents/SeverityBadge";
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
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="mb-4 flex items-center gap-2 text-lg font-semibold">
          <Zap className="h-5 w-5 text-amber-400" />
          Inject Failure
        </h2>
        <div className="grid grid-cols-4 gap-4">
          <div>
            <label className="mb-1 block text-xs text-muted-foreground">Target Service</label>
            <input
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              placeholder="payment-service"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-muted-foreground">Failure Type</label>
            <select
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
              value={action}
              onChange={(e) => setAction(e.target.value)}
            >
              {FAILURE_TYPES.map((f) => (
                <option key={f.value} value={f.value}>
                  {f.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs text-muted-foreground">Duration</label>
            <select
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
            >
              {DURATIONS.map((d) => (
                <option key={d.value} value={d.value}>
                  {d.label}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <button
              onClick={handleLaunch}
              disabled={create.isPending}
              className="flex w-full items-center justify-center gap-2 rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700 disabled:opacity-50"
            >
              {create.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Zap className="h-4 w-4" />
              )}
              Launch
            </button>
          </div>
        </div>
        <div>
          <label className="mb-1 mt-3 block text-xs text-muted-foreground">Description</label>
          <input
            className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Optional description..."
          />
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Scenario History</h2>
        {!scenarios?.length ? (
          <p className="text-sm text-muted-foreground">No scenarios yet</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left">
                  <th className="pb-2 font-medium text-muted-foreground">Name</th>
                  <th className="pb-2 font-medium text-muted-foreground">Action</th>
                  <th className="pb-2 font-medium text-muted-foreground">Runs</th>
                  <th className="pb-2 font-medium text-muted-foreground">Last Run</th>
                  <th className="pb-2 font-medium text-muted-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {scenarios.map((sc) => (
                  <tr key={sc.id} className="border-b border-border">
                    <td className="py-2">{sc.name}</td>
                    <td className="py-2">{sc.action}</td>
                    <td className="py-2">{sc.run_count}</td>
                    <td className="py-2 text-muted-foreground">
                      {sc.last_run_at ? formatDate(sc.last_run_at) : "Never"}
                    </td>
                    <td className="py-2">
                      <button
                        onClick={() => execute.mutate(sc.id)}
                        className="rounded bg-primary/10 px-2 py-1 text-xs text-primary hover:bg-primary/20"
                      >
                        Re-run
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
