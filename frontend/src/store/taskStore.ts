import { create } from "zustand";
import { Task, TaskExecution, TaskStatus } from "../types/taskTypes";
import { config } from "../config";
import TaskExecution from "../pages/TaskExecution";

interface TaskStore {
  tasks: Task[];
  currentTask: Task | null;
  taskExecution: TaskExecution | null;
  taskExecutionUrl: string | null;
  taskStats: {
    inProgress: number;
    completed: number;
    failed: number;
  };
  socket: WebSocket | null;
  fetchTasks: () => Promise<void>;
  setCurrentTask: (task: Task | null) => void;
  executeTask: (
    taskId: string,
    formData: Record<string, unknown>,
  ) => Promise<void>;
  updateTaskExecution: (log: string) => void;
  setTaskExecutionUrl: (url: string | null) => void;
  connectWebSocket: (taskId: string) => Promise<void>;
  disconnectWebSocket: () => void;
  updateTaskStats: (status: TaskStatus) => void;
}

type WebSocketEventType = "launch" | "log" | "status" | "error" | "close";

interface WebSocketEvent {
  type: WebSocketEventType;
  message?: string;
  status?: TaskStatus;
}

export const useTaskStore = create<TaskStore>((set, get) => ({
  tasks: [],
  currentTask: null,
  taskExecution: "initializing",
  taskExecutionUrl: null,
  taskStats: {
    inProgress: 0,
    completed: 0,
    failed: 0,
  },
  socket: null,

  fetchTasks: async () => {
    try {
      const response = await fetch(`${config.apiUrl}/tasks`);
      if (!response.ok) {
        throw new Error("Failed to fetch tasks");
      }
      const tasks = await response.json();
      set({ tasks });
    } catch (error) {
      console.error("Error fetching tasks:", error);
    }
  },

  setCurrentTask: (task) => set({ currentTask: task }),

  executeTask: async (taskId, formData) => {
    try {
      const response = await fetch(`${config.apiUrl}/tasks/${taskId}/execute`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ form_data: formData }),
      });
      if (!response.ok) {
        throw new Error("Failed to execute task");
      }
      const executionData = await response.json();
      set({
        taskExecution: {
          id: executionData.task_id,
          taskId: taskId,
          status: executionData.status,
          logs: [],
        },
        taskExecutionUrl: `/task-execution/${taskId}`,
      });
      return executionData;
    } catch (error) {
      console.error("Error executing task:", error);
      throw error;
    }
  },

  connectWebSocket: async (taskId: string) => {
    await new Promise((resolve) => setTimeout(resolve, config.websocket_delay));

    const socket = new WebSocket(
      `ws://${config.apiUrl.replace(/^https?:\/\//, "")}/ws/task/${taskId}`,
    );

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data) as WebSocketEvent;
      switch (data.type) {
        case "launch":
          console.log("Task launched:", data.message);
          set((state) => ({
            taskExecution: state.taskExecution
              ? { ...state.taskExecution, status: "running" }
              : null,
          }));
          get().updateTaskStats("running");
          break;
        case "log":
          if (data.message) {
            set((state) => ({
              taskExecution: state.taskExecution
                ? {
                    ...state.taskExecution,
                    logs: [...state.taskExecution.logs, data.message],
                  }
                : null,
            }));
          }
          break;
        case "status":
          if (data.status) {
            set((state) => ({
              taskExecution: state.taskExecution
                ? { ...state.taskExecution, status: data.status }
                : null,
            }));
            get().updateTaskStats(data.status);
            if (data.status === "Succeeded") {
              get().disconnectWebSocket();
            }
          }
          break;
        case "error":
          console.error("Task error:", data.message);
          set((state) => ({
            taskExecution: state.taskExecution
              ? {
                  ...state.taskExecution,
                  error: data.message,
                  status: "failed",
                }
              : null,
          }));
          get().updateTaskStats("failed");
          get().disconnectWebSocket();
          break;
        case "close":
          console.log("Task execution finished:", data.message);
          set((state) => ({
            taskExecution: state.taskExecution
              ? { ...state.taskExecution, status: "closed" }
              : null,
          }));
          get().updateTaskStats("closed");
          get().disconnectWebSocket();

          break;
        default:
          console.warn("Unknown event type:", (data as WebSocketEvent).type);
      }
    };

    socket.onclose = (event) => {
      if (!event.wasClean) {
        console.error("WebSocket connection closed unexpectedly");
        set((state) => ({
          taskExecution: state.taskExecution
            ? {
                ...state.taskExecution,
                error: "Connection closed unexpectedly",
                status: "failed",
              }
            : null,
        }));
        get().updateTaskStats("failed");
      }
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      set((state) => ({
        taskExecution: state.taskExecution
          ? {
              ...state.taskExecution,
              error: "WebSocket error occurred",
              status: "failed",
            }
          : null,
      }));
      get().updateTaskStats("failed");
      get().disconnectWebSocket();
    };

    set({ socket });
  },

  disconnectWebSocket: () => {
    const { socket } = get();
    if (socket) {
      socket.close();
      set({ socket: null });
    }
  },

  updateTaskStats: (status: TaskStatus) => {
    set((state) => {
      const newStats = { ...state.taskStats };
      if (status === "running") {
        newStats.inProgress += 1;
      } else if (status === "Succeeded") {
        newStats.inProgress -= 1;
        newStats.completed += 1;
      } else if (status === "failed") {
        newStats.inProgress -= 1;
        newStats.failed += 1;
      }
      return { taskStats: newStats };
    });
  },

  updateTaskExecution: (log: string) => {
    set((state) => ({
      taskExecution: state.taskExecution
        ? { ...state.taskExecution, logs: [...state.taskExecution.logs, log] }
        : null,
    }));
  },

  setTaskExecutionUrl: (url: string | null) => set({ taskExecutionUrl: url }),
}));
