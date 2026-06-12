"use client";

import React from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useToggleTaskStatus, useDeleteTask } from "@/hooks/use-tasks";
import type { Task } from "@/types";
import {
  formatDate,
  getPriorityColor,
  getPriorityLabel,
  getStatusColor,
  getStatusLabel,
} from "@/lib/utils";
import {
  Pencil,
  Trash2,
  Clock,
  ArrowUp,
  ArrowDown,
  CheckCircle2,
  Circle,
  Loader2,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useTaskActivity } from "@/hooks/use-tasks";
import { ScrollArea } from "@/components/ui/scroll-area";

interface TaskCardProps {
  task: Task;
  onEdit: (task: Task) => void;
}

export function TaskCard({ task, onEdit }: TaskCardProps) {
  const toggleStatus = useToggleTaskStatus();
  const deleteTask = useDeleteTask();
  const [showActivity, setShowActivity] = React.useState(false);

  const isCompleted = task.status === "completed";

  const handleToggle = () => {
    const newStatus = isCompleted ? "pending" : "completed";
    toggleStatus.mutate({ id: task.id, status: newStatus });
  };

  const handleDelete = () => {
    if (window.confirm("Are you sure you want to delete this task?")) {
      deleteTask.mutate(task.id);
    }
  };

  return (
    <div className="group relative rounded-xl border bg-card p-5 shadow-sm transition-all duration-200 hover:shadow-md hover:border-primary/20">
      <div className="flex items-start gap-4">
        {/* Status toggle */}
        <button
          onClick={handleToggle}
          disabled={toggleStatus.isPending}
          className="mt-0.5 shrink-0 text-muted-foreground hover:text-primary transition-colors"
        >
          {toggleStatus.isPending ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : isCompleted ? (
            <CheckCircle2 className="h-5 w-5 text-emerald-500" />
          ) : (
            <Circle className="h-5 w-5" />
          )}
        </button>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <h3
              className={cn(
                "text-base font-medium leading-snug",
                isCompleted && "line-through text-muted-foreground/60"
              )}
            >
              {task.title}
            </h3>

            {/* Actions */}
            <div className="flex items-center gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => onEdit(task)}
              >
                <Pencil className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-destructive hover:text-destructive"
                onClick={handleDelete}
                disabled={deleteTask.isPending}
              >
                {deleteTask.isPending ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <Trash2 className="h-3.5 w-3.5" />
                )}
              </Button>
            </div>
          </div>

          {/* Description */}
          {task.description && (
            <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
              {task.description}
            </p>
          )}

          {/* Meta badges */}
          <div className="mt-3 flex flex-wrap items-center gap-2">
            <span
              className={cn(
                "inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium",
                getStatusColor(task.status)
              )}
            >
              {getStatusLabel(task.status)}
            </span>

            <span
              className={cn(
                "inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium",
                getPriorityColor(task.priority)
              )}
            >
              {getPriorityLabel(task.priority)}
            </span>

            {task.due_date && (
              <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                <Clock className="h-3 w-3" />
                {formatDate(task.due_date)}
              </span>
            )}
          </div>

          {/* Created date */}
          <p className="mt-2 text-[11px] text-muted-foreground">
            Created {formatDate(task.created_at)}
          </p>
        </div>
      </div>
    </div>
  );
}