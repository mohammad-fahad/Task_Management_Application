"use client";

import React, { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDebounce } from "@/hooks/use-debounce";
import { Plus, Search, SlidersHorizontal } from "lucide-react";
import type { TaskFilter, TaskStatus } from "@/types";
import { getStatusLabel } from "@/lib/utils";

interface TaskFilterBarProps {
  filter: TaskFilter;
  onFilterChange: (filter: TaskFilter) => void;
  onCreateClick: () => void;
}

const statusOptions: { value: TaskStatus | "all"; label: string }[] = [
  { value: "all", label: "All Statuses" },
  { value: "pending", label: "Todo" },
  { value: "in_progress", label: "In Progress" },
  { value: "completed", label: "Completed" },
  { value: "cancelled", label: "Cancelled" },
];

const sortOptions = [
  { value: "created_at", label: "Created At" },
  { value: "due_date", label: "Due Date" },
  { value: "priority", label: "Priority" },
  { value: "title", label: "Title" },
  { value: "status", label: "Status" },
];

export function TaskFilterBar({ filter, onFilterChange, onCreateClick }: TaskFilterBarProps) {
  const [search, setSearch] = useState(filter.search || "");
  const debouncedSearch = useDebounce(search, 300);

  useEffect(() => {
    onFilterChange({ ...filter, search: debouncedSearch || "" });
  }, [debouncedSearch]);

  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      {/* Search */}
      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search tasks..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9 h-10"
        />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-2">
        <Select
          value={filter.status || "all"}
          onValueChange={(value) =>
            onFilterChange({
              ...filter,
              status: (value === "all" ? "" : value) as TaskStatus | "",
              page: 1,
            })
          }
        >
          <SelectTrigger className="w-[140px] h-10">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            {statusOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={filter.sort_by || "created_at"}
          onValueChange={(value) =>
            onFilterChange({
              ...filter,
              sort_by: value as TaskFilter["sort_by"],
              page: 1,
            })
          }
        >
          <SelectTrigger className="w-[140px] h-10">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            {sortOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button
          variant="outline"
          size="icon"
          className="h-10 w-10"
          onClick={() =>
            onFilterChange({
              ...filter,
              sort_dir: filter.sort_dir === "asc" ? "desc" : "asc",
            })
          }
          title={`Sort ${filter.sort_dir === "asc" ? "descending" : "ascending"}`}
        >
          <SlidersHorizontal className="h-4 w-4" />
        </Button>

        <Button onClick={onCreateClick} className="h-10 gap-1.5">
          <Plus className="h-4 w-4" />
          <span className="hidden sm:inline">New Task</span>
        </Button>
      </div>
    </div>
  );
}