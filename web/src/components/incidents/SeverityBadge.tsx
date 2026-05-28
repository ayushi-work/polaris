import { cn } from "@/lib/utils";

interface SeverityBadgeProps {
  severity: string;
  className?: string;
}

const colors: Record<string, string> = {
  critical: "bg-red-500/10 text-red-400 border-red-500/30",
  warning: "bg-amber-500/10 text-amber-400 border-amber-500/30",
  info: "bg-blue-500/10 text-blue-400 border-blue-500/30",
};

export function SeverityBadge({ severity, className }: SeverityBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium",
        colors[severity] || colors.info,
        className
      )}
    >
      {severity}
    </span>
  );
}

export function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    detected: "bg-amber-500/10 text-amber-400",
    investigating: "bg-blue-500/10 text-blue-400",
    remediating: "bg-purple-500/10 text-purple-400",
    resolved: "bg-emerald-500/10 text-emerald-400",
    failed: "bg-red-500/10 text-red-400",
    pending: "bg-gray-500/10 text-gray-400",
    running: "bg-blue-500/10 text-blue-400",
    success: "bg-emerald-500/10 text-emerald-400",
    skipped: "bg-gray-500/10 text-gray-400",
  };

  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
        styles[status] || styles.detected
      }`}
    >
      {status}
    </span>
  );
}
