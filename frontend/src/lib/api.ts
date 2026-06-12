import axios, { AxiosError, type AxiosInstance } from "axios";
import type {
  APIResponse,
  AuthResponse,
  Task,
  TaskInput,
  TaskFilter,
  TaskListResponse,
  ActivityLogResponse,
  User,
} from "@/types";

// ============================================================================
// Axios instance – secure cookie-based auth
// ============================================================================

const api: AxiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1",
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

// ============================================================================
// Error helpers
// ============================================================================

export class ApiError extends Error {
  status: number;
  errors?: { field: string; message: string }[];

  constructor(message: string, status: number, errors?: { field: string; message: string }[]) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.errors = errors;
  }
}

export function getApiError(err: unknown): ApiError {
  if (err instanceof AxiosError && err.response) {
    const data = err.response.data as APIResponse & { errors?: { field: string; message: string }[] };
    const status = err.response.status;
    if (data?.errors) {
      return new ApiError(data.message || "Validation failed", status, data.errors);
    }
    return new ApiError(data?.error || data?.message || "Request failed", status);
  }
  if (err instanceof Error) {
    return new ApiError(err.message, 0);
  }
  return new ApiError("Unknown error", 0);
}

// ============================================================================
// Auth API
// ============================================================================

export const authApi = {
  async register(input: { email: string; password: string }): Promise<User> {
    const { data } = await api.post<APIResponse<User>>("/auth/register", input);
    if (!data.success || !data.data) throw new ApiError(data.error || "Registration failed", 400);
    return data.data;
  },

  async login(input: { email: string; password: string }): Promise<AuthResponse> {
    const { data } = await api.post<APIResponse<AuthResponse>>("/auth/login", input);
    if (!data.success || !data.data) throw new ApiError(data.error || "Login failed", 401);
    return data.data;
  },

  async logout(): Promise<void> {
    await api.post("/auth/logout");
  },

  async me(): Promise<User> {
    const { data } = await api.get<APIResponse<User>>("/auth/me");
    if (!data.success || !data.data) throw new ApiError(data.error || "Not authenticated", 401);
    return data.data;
  },
};

// ============================================================================
// Tasks API
// ============================================================================

function buildTaskFilterQuery(filter: TaskFilter): string {
  const params = new URLSearchParams();
  if (filter.page) params.set("page", String(filter.page));
  if (filter.per_page) params.set("per_page", String(filter.per_page));
  if (filter.status) params.set("status", filter.status);
  if (filter.priority) params.set("priority", filter.priority);
  if (filter.search) params.set("search", filter.search);
  if (filter.sort_by) params.set("sort_by", filter.sort_by);
  if (filter.sort_dir) params.set("sort_dir", filter.sort_dir);
  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

export const tasksApi = {
  async list(filter: TaskFilter = {}): Promise<TaskListResponse> {
    const qs = buildTaskFilterQuery(filter);
    const { data } = await api.get<APIResponse<TaskListResponse>>(`/tasks${qs}`);
    if (!data.success || !data.data) throw new ApiError(data.error || "Failed to fetch tasks", 500);
    return data.data;
  },

  async get(id: string): Promise<Task> {
    const { data } = await api.get<APIResponse<Task>>(`/tasks/${id}`);
    if (!data.success || !data.data) throw new ApiError(data.error || "Task not found", 404);
    return data.data;
  },

  async create(input: TaskInput): Promise<Task> {
    const { data } = await api.post<APIResponse<Task>>("/tasks", input);
    if (!data.success || !data.data) throw new ApiError(data.error || "Failed to create task", 500);
    return data.data;
  },

  async update(id: string, input: TaskInput): Promise<Task> {
    const { data } = await api.put<APIResponse<Task>>(`/tasks/${id}`, input);
    if (!data.success || !data.data) throw new ApiError(data.error || "Failed to update task", 500);
    return data.data;
  },

  async delete(id: string): Promise<void> {
    const { data } = await api.delete<APIResponse>(`/tasks/${id}`);
    if (!data.success) throw new ApiError(data.error || "Failed to delete task", 500);
  },

  async getActivity(taskId: string, limit = 50, offset = 0): Promise<ActivityLogResponse> {
    const { data } = await api.get<APIResponse<ActivityLogResponse>>(
      `/tasks/${taskId}/activity?limit=${limit}&offset=${offset}`
    );
    if (!data.success || !data.data) throw new ApiError(data.error || "Failed to fetch activity", 500);
    return data.data;
  },
};

export default api;