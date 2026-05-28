import { useDashboardMetrics } from "@/hooks/useMetrics";
import { ClusterHealthCard } from "@/components/dashboard/ClusterHealthCard";
import { ActiveIncidentsCard } from "@/components/dashboard/ActiveIncidentsCard";
import { ReliabilityScore } from "@/components/dashboard/ReliabilityScore";
import { MttrMttdChart } from "@/components/dashboard/MttrMttdChart";
import { RecentActivity } from "@/components/dashboard/RecentActivity";

export default function DashboardPage() {
  const { data, isLoading } = useDashboardMetrics();

  if (isLoading) {
    return <div className="text-sm text-muted-foreground">Loading dashboard...</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold">Overview</h1>

      <div className="grid grid-cols-3 gap-4">
        <ClusterHealthCard health={data?.cluster_health || "Unknown"} />
        <ActiveIncidentsCard
          active={data?.active_incidents || 0}
          today={data?.incidents_today || 0}
        />
        <ReliabilityScore
          score={data?.reliability_score || 100}
          mttr={data?.mttr || "N/A"}
          mttd={data?.mttd || "N/A"}
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <MttrMttdChart />
        <RecentActivity />
      </div>
    </div>
  );
}
