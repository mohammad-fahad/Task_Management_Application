// ============================================================================
// Domain Types – matching Go backend models
// ============================================================================

export interface User {
  id: string;
  email: string;
  role: "user" | "admin";
  created_at: string;
}

export interface UserInput {
  email: string;
  password: string;
  role?: "user" | "admin";
}

export interface LoginInput {
  email: string;
  password: string;
}

export type TaskStatus = "pending" | "in_progress" | "completed" | "cancelled";
export type TaskPriority = "low" | "medium" | "high" | "critical";

export interface Task {
  id: string;
  user_id: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  due_date: string | null;
  created_at: string;
}

export interface TaskInput {
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  due_date?: string | null;
}

export interface TaskFilter {
  page?: number;
  per_page?: number;
  status?: TaskStatus | "";
  priority?: TaskPriority | "";
  search?: string;
  sort_by?: "title" | "status" | "priority" | "due_date" | "created_at";
  sort_dir?: "asc" | "desc";
}

export interface TaskListResponse {
  data: Task[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface ActivityLog {
  id: string;
  task_id: string;
  user_id: string;
  action: string;
  details: Record<string, unknown>;
  created_at: string;
}

export interface ActivityLogResponse {
  activity_logs: ActivityLog[];
  total: number;
  limit: number;
  offset: number;
}

// ============================================================================
// API Response Wrappers
// ============================================================================

export interface APIResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

export interface ErrorDetail {
  field: string;
  message: string;
}

export interface ValidationErrorResponse {
  success: boolean;
  message: string;
  errors: ErrorDetail[];
}

// ============================================================================
// Auth helpers
// ============================================================================

export interface AuthResponse {
  user: User;
  token: string;
}