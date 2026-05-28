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
  { to: "/", label: "Overview", icon: LayoutDashboard },
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
    <aside className="flex w-52 flex-col border-r border-border bg-white">
      <div className="flex h-14 items-center gap-2.5 border-b border-border px-5">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
          <Zap className="h-4 w-4 text-primary-foreground" />
        </div>
        <span className="font-serif text-lg italic tracking-tight">Polaris</span>
      </div>
      <nav className="flex-1 space-y-0.5 p-3">
        {links.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === "/"}
            className={({ isActive }) =>
              `flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors ${
                isActive
                  ? "bg-stone-100 text-primary font-medium"
                  : "text-stone-500 hover:bg-stone-50 hover:text-stone-800"
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
