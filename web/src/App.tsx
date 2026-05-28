import { Routes, Route } from "react-router-dom";
import { Shell } from "@/components/layout/Shell";
import { useWebSocket } from "@/hooks/useWebSocket";
import DashboardPage from "@/pages/DashboardPage";
import IncidentsPage from "@/pages/IncidentsPage";
import IncidentDetailPage from "@/pages/IncidentDetailPage";
import ChaosLabPage from "@/pages/ChaosLabPage";
import TopologyPage from "@/pages/TopologyPage";
import LogsPage from "@/pages/LogsPage";
import MetricsPage from "@/pages/MetricsPage";
import HealingPage from "@/pages/HealingPage";
import PostmortemsPage from "@/pages/PostmortemsPage";

function WSListener() {
  useWebSocket();
  return null;
}

export function App() {
  return (
    <>
      <WSListener />
      <Routes>
        <Route element={<Shell />}>
          <Route index element={<DashboardPage />} />
          <Route path="incidents" element={<IncidentsPage />} />
          <Route path="incidents/:id" element={<IncidentDetailPage />} />
          <Route path="chaos" element={<ChaosLabPage />} />
          <Route path="topology" element={<TopologyPage />} />
          <Route path="logs" element={<LogsPage />} />
          <Route path="metrics" element={<MetricsPage />} />
          <Route path="healing" element={<HealingPage />} />
          <Route path="postmortems" element={<PostmortemsPage />} />
        </Route>
      </Routes>
    </>
  );
}
