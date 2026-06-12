"use client";

import React, { useState, useCallback } from "react";
import { useAuth } from "@/lib/auth-context";
import { useTasks } from "@/hooks/use-tasks";
import { Sidebar } from "@/components/dashboard/sidebar";
import { TaskFilterBar } from "@/components/dashboard/task-filter-bar";
import { TaskFormModal } from "@/components/dashboard/task-form-modal";
import { TaskCard } from "@/components/dashboard/task-card";
import { ActivityLog } from "@/components/dashboard/activity-log";
import { Button } from "@/components/ui/button";
import type { Task, TaskFilter } from "@/types";
import {
  Loader2,
  Menu,
  ClipboardList,
  History,
  ChevronLeft,
  ChevronRight,
  ArrowUpDown,
} from "lucide-react";
import { cn } from "@/lib/utils";

export default function DashboardPage() {
  const { user, isLoading: authLoading } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [filter, setFilter] = useState<TaskFilter>({
    page: 1,
    per_page: 20,
    sort_by: "created_at",
    sort_dir: "desc",
  });
  const [formOpen, setFormOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [activityTaskId, setActivityTaskId] = useState<string | null>(null);

  const { data, isLoading, isError } = useTasks(filter);

  const handleFilterChange = useCallback((newFilter: TaskFilter) => {
    setFilter(newFilter);
  }, []);

  const handleEdit = useCallback((task: Task) => {
    setEditingTask(task);
    setFormOpen(true);
  }, []);

  const handleCreate = useCallback(() => {
    setEditingTask(null);
    setFormOpen(true);
  }, []);

  const handleFormClose = useCallback(() => {
    setFormOpen(false);
    setEditingTask(null);
  }, []);

  const handlePageChange = useCallback((page: number) => {
    setFilter((prev) => ({ ...prev, page }));
  }, []);

  if (authLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-3">
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Prevent render if not authenticated (redirect handled in auth context)
  if (!user) {
    return null;
  }

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar */}
      <Sidebar isOpen={sidebarOpen} onToggle={() => setSidebarOpen(false)} />

      {/* Main content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Top bar */}
        <header className="flex h-16 items-center gap-4 border-b bg-card/50 px-6 backdrop-blur-sm">
          <button
            onClick={() => setSidebarOpen(true)}
            className="rounded-lg p-1.5 text-muted-foreground hover:bg-accent lg:hidden"
          >
            <Menu className="h-5 w-5" />
          </button>

          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <ClipboardList className="h-4 w-4 text-primary" />
            </div>
            <h1 className="text-lg font-semibold">Tasks</h1>
          </div>

          <div className="flex-1" />

          {/* Quick activity access placeholder */}
          <p className="text-xs text-muted-foreground hidden sm:block">
            {data?.total ?? 0} task{data?.total !== 1 ? "s" : ""}
          </p>
        </header>

        {/* Content scroll area */}
        <div className="flex-1 overflow-y-auto">
          <div className="mx-auto w-full max-w-4xl space-y-6 px-6 py-6">
            {/* Filter bar */}
            <TaskFilterBar
              filter={filter}
              onFilterChange={handleFilterChange}
              onCreateClick={handleCreate}
            />

            {/* Task list */}
            {isLoading ? (
              <div className="space-y-4">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div key={i} className="rounded-xl border p-5">
                    <div className="flex items-start gap-4">
                      <div className="h-5 w-5 rounded-full skeleton" />
                      <div className="flex-1 space-y-3">
                        <div className="h-5 w-3/4 skeleton" />
                        <div className="h-4 w-1/2 skeleton" />
                        <div className="flex gap-2">
                          <div className="h-5 w-20 skeleton" />
                          <div className="h-5 w-16 skeleton" />
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : isError ? (
              <div className="flex flex-col items-center justify-center py-20 text-center">
                <div className="rounded-full bg-destructive/10 p-4 mb-4">
                  <ClipboardList className="h-8 w-8 text-destructive" />
                </div>
                <h3 className="text-lg font-medium mb-1">Failed to load tasks</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Something went wrong. Please try again.
                </p>
                <Button onClick={() => setFilter({ ...filter })}>Retry</Button>
              </div>
            ) : data?.data?.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20 text-center">
                <div className="rounded-full bg-muted p-4 mb-4">
                  <ClipboardList className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-medium mb-1">No tasks found</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  {filter.search
                    ? "No tasks match your search. Try different keywords."
                    : "You haven't created any tasks yet. Get started!"}
                </p>
                <Button onClick={handleCreate}>Create your first task</Button>
              </div>
            ) : (
              <>
                <div className="space-y-3">
                  {data?.data?.map((task) => (
                    <div key={task.id} className="relative">
                      <TaskCard task={task} onEdit={handleEdit} />
                      <button
                        onClick={() =>
                          setActivityTaskId(
                            activityTaskId === task.id ? null : task.id
                          )
                        }
                        className="absolute top-3 right-12 p-1 text-muted-foreground hover:text-primary opacity-0 group-hover:opacity-100 transition-opacity"
                        title="View activity"
                      >
                        <History className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  ))}
                </div>

                {/* Pagination */}
                {data && data.total_pages > 1 && (
                  <div className="flex items-center justify-center gap-2 pt-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(data.page - 1)}
                      disabled={data.page <= 1}
                    >
                      <ChevronLeft className="h-4 w-4" />
                      Previous
                    </Button>

                    <span className="text-sm text-muted-foreground px-4">
                      Page {data.page} of {data.total_pages}
                    </span>

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(data.page + 1)}
                      disabled={data.page >= data.total_pages}
                    >
                      Next
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      {/* Activity log drawer */}
      <ActivityLog
        taskId={activityTaskId || ""}
        isOpen={!!activityTaskId}
        onClose={() => setActivityTaskId(null)}
      />

      {/* Task form modal */}
      <TaskFormModal
        open={formOpen}
        onOpenChange={handleFormClose}
        task={editingTask}
      />
    </div>
  );
}