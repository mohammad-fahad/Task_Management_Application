"use client";

import React from "react";
import { useTaskActivity } from "@/hooks/use-tasks";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDateTime } from "@/lib/utils";
import { Loader2, History, Clock, User } from "lucide-react";

interface ActivityLogProps {
  taskId: string;
  isOpen: boolean;
  onClose: () => void;
}

export function ActivityLog({ taskId, isOpen, onClose }: ActivityLogProps) {
  const { data, isLoading, isError } = useTaskActivity(taskId);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-y-0 right-0 z-50 w-full max-w-md border-l bg-card shadow-2xl animate-slide-in-right">
      <div className="flex h-16 items-center justify-between border-b px-6">
        <div className="flex items-center gap-2">
          <History className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Activity Log</h2>
        </div>
        <button
          onClick={onClose}
          className="rounded-lg p-1.5 text-muted-foreground hover:bg-accent transition-colors"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M18 6 6 18" />
            <path d="m6 6 12 12" />
          </svg>
        </button>
      </div>

      <ScrollArea className="h-[calc(100vh-4rem)] p-6">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : isError ? (
          <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
            <p>Failed to load activity log</p>
          </div>
        ) : !data?.activity_logs?.length ? (
          <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
            <History className="h-10 w-10 mb-3 opacity-30" />
            <p className="text-sm">No activity recorded yet</p>
          </div>
        ) : (
          <div className="relative">
            {/* Timeline line */}
            <div className="absolute left-[11px] top-2 bottom-2 w-px bg-border" />

            <div className="space-y-6">
              {data.activity_logs.map((log) => (
                <div key={log.id} className="relative flex gap-4 pl-8">
                  {/* Timeline dot */}
                  <div className="absolute left-0 top-1.5 flex h-[22px] w-[22px] items-center justify-center rounded-full border-2 border-background bg-primary/10">
                    <div className="h-2 w-2 rounded-full bg-primary" />
                  </div>

                  {/* Content */}
                  <div className="flex-1">
                    <div className="rounded-lg border bg-card/50 p-3">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mb-1.5">
                        <Clock className="h-3 w-3" />
                        {formatDateTime(log.created_at)}
                      </div>
                      <p className="text-sm font-medium capitalize">{log.action.replace(/_/g, " ")}</p>
                      {log.details && Object.keys(log.details).length > 0 && (
                        <div className="mt-1.5 space-y-0.5">
                          {Object.entries(log.details).map(([key, value]) => (
                            <p key={key} className="text-xs text-muted-foreground">
                              <span className="font-medium">{key}:</span>{" "}
                              {typeof value === "string" ? value : JSON.stringify(value)}
                            </p>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </ScrollArea>
    </div>
  );
}