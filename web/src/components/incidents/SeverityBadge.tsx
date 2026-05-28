import { cn } from "@/lib/utils";

interface SeverityBadgeProps {
  severity: string;
  className?: string;
}

const colors: Record<string, string> = {
  critical: "bg-red-50 text-red-700 border-red-200",
  warning: "bg-amber-50 text-amber-700 border-amber-200",
  info: "bg-blue-50 text-blue-700 border-blue-200",
};

export function SeverityBadge({ severity, className }: SeverityBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium",
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
    detected: "bg-amber-50 text-amber-700 border-amber-200",
    investigating: "bg-blue-50 text-blue-700 border-blue-200",
    remediating: "bg-purple-50 text-purple-700 border-purple-200",
    resolved: "bg-emerald-50 text-emerald-700 border-emerald-200",
    failed: "bg-red-50 text-red-700 border-red-200",
    pending: "bg-stone-100 text-stone-600 border-stone-200",
    running: "bg-blue-50 text-blue-700 border-blue-200",
    success: "bg-emerald-50 text-emerald-700 border-emerald-200",
    skipped: "bg-stone-100 text-stone-500 border-stone-200",
  };

  return (
    <span
      className={`inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium ${
        styles[status] || styles.detected
      }`}
    >
      {status}
    </span>
  );
}
