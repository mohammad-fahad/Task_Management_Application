import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from "@tanstack/react-query";
import { tasksApi, getApiError } from "@/lib/api";
import type { Task, TaskFilter, TaskInput, TaskListResponse } from "@/types";
import { toast } from "sonner";

// ============================================================================
// Query keys
// ============================================================================

const TASKS_KEY = "tasks";
const TASK_KEY = "task";
const ACTIVITY_KEY = "task-activity";

export function taskKeys() {
  return {
    all: [TASKS_KEY] as const,
    list: (filter: TaskFilter) => [TASKS_KEY, "list", filter] as const,
    detail: (id: string) => [TASK_KEY, id] as const,
    activity: (id: string) => [ACTIVITY_KEY, id] as const,
  };
}

// ============================================================================
// List tasks
// ============================================================================

export function useTasks(filter: TaskFilter = {}) {
  return useQuery<TaskListResponse>({
    queryKey: taskKeys().list(filter),
    queryFn: () => tasksApi.list(filter),
    placeholderData: (prev) => prev,
  });
}

// ============================================================================
// Single task
// ============================================================================

export function useTask(id: string) {
  return useQuery<Task>({
    queryKey: taskKeys().detail(id),
    queryFn: () => tasksApi.get(id),
    enabled: !!id,
  });
}

// ============================================================================
// Task mutations with Optimistic UI
// ============================================================================

export function useCreateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: TaskInput) => tasksApi.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys().all });
      toast.success("Task created successfully");
    },
    onError: (err) => {
      const apiErr = getApiError(err);
      toast.error(apiErr.message);
    },
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: TaskInput }) =>
      tasksApi.update(id, input),

    onMutate: async ({ id, input }) => {
      // Cancel outgoing queries for this task
      await queryClient.cancelQueries({ queryKey: taskKeys().detail(id) });

      // Snapshot previous value
      const previousTask = queryClient.getQueryData<Task>(taskKeys().detail(id));

      // Optimistically update the cache
      if (previousTask) {
        queryClient.setQueryData<Task>(taskKeys().detail(id), {
          ...previousTask,
          ...input,
          due_date: input.due_date !== undefined ? input.due_date : previousTask.due_date,
        });
      }

      return { previousTask };
    },

    onError: (err, { id }, context) => {
      // Rollback on error
      if (context?.previousTask) {
        queryClient.setQueryData<Task>(taskKeys().detail(id), context.previousTask);
      }
      const apiErr = getApiError(err);
      toast.error(apiErr.message);
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys().all });
    },

    onSuccess: () => {
      toast.success("Task updated successfully");
    },
  });
}

export function useToggleTaskStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, status }: { id: string; status: Task["status"] }) => {
      // Get current task data
      const current = queryClient.getQueryData<Task>(taskKeys().detail(id));
      if (!current) throw new Error("Task not in cache");

      return tasksApi.update(id, {
        title: current.title,
        description: current.description,
        status,
        priority: current.priority,
        due_date: current.due_date,
      });
    },

    onMutate: async ({ id, status }) => {
      await queryClient.cancelQueries({ queryKey: taskKeys().detail(id) });

      const previousTask = queryClient.getQueryData<Task>(taskKeys().detail(id));

      if (previousTask) {
        queryClient.setQueryData<Task>(taskKeys().detail(id), {
          ...previousTask,
          status,
        });
      }

      // Also update in list cache
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: [TASKS_KEY, "list"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((t) =>
              t.id === id ? { ...t, status } : t
            ),
          };
        }
      );

      return { previousTask };
    },

    onError: (err, { id }, context) => {
      if (context?.previousTask) {
        queryClient.setQueryData<Task>(taskKeys().detail(id), context.previousTask);
      }
      const apiErr = getApiError(err);
      toast.error(apiErr.message);
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys().all });
    },

    onSuccess: () => {
      toast.success("Task status updated");
    },
  });
}

export function useDeleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => tasksApi.delete(id),

    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: taskKeys().all });

      // Remove from all list caches optimistically
      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: [TASKS_KEY, "list"],
      });

      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: [TASKS_KEY, "list"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.filter((t) => t.id !== id),
            total: old.total - 1,
          };
        }
      );

      return { previousLists };
    },

    onError: (_err, _id, context) => {
      // Rollback all list caches
      if (context?.previousLists) {
        for (const [queryKey, data] of context.previousLists) {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        }
      }
      toast.error("Failed to delete task");
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys().all });
    },

    onSuccess: () => {
      toast.success("Task deleted successfully");
    },
  });
}

// ============================================================================
// Activity Log
// ============================================================================

export function useTaskActivity(taskId: string) {
  return useQuery({
    queryKey: taskKeys().activity(taskId),
    queryFn: () => tasksApi.getActivity(taskId),
    enabled: !!taskId,
  });
}