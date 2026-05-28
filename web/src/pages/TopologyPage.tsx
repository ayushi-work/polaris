import { ClusterTopology } from "@/components/topology/ClusterTopology";

export default function TopologyPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl italic text-stone-900">Cluster Topology</h1>
        <p className="mt-1 text-sm text-stone-500">Service dependency graph with health indicators.</p>
      </div>
      <div className="h-[500px] rounded-xl border border-border bg-white overflow-hidden">
        <ClusterTopology />
      </div>
    </div>
  );
}
