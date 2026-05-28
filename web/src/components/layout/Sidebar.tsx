import { NavLink } from "react-router-dom";
import {
  LayoutDashboard,
  AlertTriangle,
  Zap,
  GitGraph,
  ScrollText,
  BarChart3,
  HeartPulse,
  FileText,
} from "lucide-react";

const links = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard },
  { to: "/incidents", label: "Incidents", icon: AlertTriangle },
  { to: "/chaos", label: "Chaos Lab", icon: Zap },
  { to: "/topology", label: "Topology", icon: GitGraph },
  { to: "/logs", label: "Live Logs", icon: ScrollText },
  { to: "/metrics", label: "Metrics", icon: BarChart3 },
  { to: "/healing", label: "Self-Healing", icon: HeartPulse },
  { to: "/postmortems", label: "Postmortems", icon: FileText },
];

export function Sidebar() {
  return (
    <aside className="flex w-56 flex-col border-r border-border bg-card">
      <div className="flex h-14 items-center gap-2 border-b border-border px-4">
        <Zap className="h-5 w-5 text-primary" />
        <span className="font-semibold text-sm">Polaris</span>
      </div>
      <nav className="flex-1 space-y-1 p-3">
        {links.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === "/"}
            className={({ isActive }) =>
              `flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-accent hover:text-foreground"
              }`
            }
          >
            <Icon className="h-4 w-4" />
            {label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
