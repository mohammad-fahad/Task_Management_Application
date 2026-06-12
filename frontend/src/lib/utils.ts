import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(dateString: string | null): string {
  if (!dateString) return "—";
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatDateTime(dateString: string | null): string {
  if (!dateString) return "—";
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function getStatusColor(status: string): string {
  const map: Record<string, string> = {
    pending: "bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 border-yellow-500/30",
    in_progress: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30",
    completed: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30",
    cancelled: "bg-neutral-500/10 text-neutral-600 dark:text-neutral-400 border-neutral-500/30",
  };
  return map[status] ?? "bg-neutral-500/10 text-neutral-600";
}

export function getPriorityColor(priority: string): string {
  const map: Record<string, string> = {
    low: "bg-slate-500/10 text-slate-600 dark:text-slate-400",
    medium: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
    high: "bg-orange-500/10 text-orange-600 dark:text-orange-400",
    critical: "bg-red-500/10 text-red-600 dark:text-red-400",
  };
  return map[priority] ?? "bg-slate-500/10 text-slate-600";
}

export function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: "Todo",
    in_progress: "In Progress",
    completed: "Completed",
    cancelled: "Cancelled",
  };
  return labels[status] ?? status;
}

export function getPriorityLabel(priority: string): string {
  return priority.charAt(0).toUpperCase() + priority.slice(1);
}

export function cnJoin(...classes: (string | undefined | null | false)[]): string {
  return classes.filter(Boolean).join(" ");
}