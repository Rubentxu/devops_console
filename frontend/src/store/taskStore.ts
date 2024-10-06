import { create } from "zustand";
import { Task, TaskExecution, TaskStatus } from "../types/taskTypes";
import { config } from "../config";

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
  executeTask: (taskId: string, formData: Record<string, any>) => Promise<void>;
  updateTaskExecution: (log: string) => void;
  setTaskExecutionUrl: (url: string | null) => void;
  connectWebSocket: (taskId: string) => Promise<void>;
  disconnectWebSocket: () => void;
  updateTaskStats: (status: TaskStatus) => void;
}

export const useTaskStore = create<TaskStore>((set, get) => ({
  tasks: [],
  currentTask: null,
  taskExecution: null,
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
      const data = JSON.parse(event.data);
      if (data.type === "log") {
        set((state) => ({
          taskExecution: state.taskExecution
            ? {
                ...state.taskExecution,
                logs: [...state.taskExecution.logs, data.message],
              }
            : null,
        }));
      } else if (data.type === "status") {
        set((state) => ({
          taskExecution: state.taskExecution
            ? { ...state.taskExecution, status: data.status }
            : null,
        }));
        get().updateTaskStats(data.status);
      } else if (data.type === "error") {
        set((state) => ({
          taskExecution: state.taskExecution
            ? { ...state.taskExecution, error: data.message, status: "failed" }
            : null,
        }));
        get().updateTaskStats("failed");
        get().disconnectWebSocket();
        socket.send(
          JSON.stringify({ type: "close", reason: "Error event received" }),
        );
      } else {
        console.error("Unknown event type received:", data.type);
        get().disconnectWebSocket();
        socket.send(
          JSON.stringify({ type: "close", reason: "Unknown event type" }),
        );
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
      socket.send(JSON.stringify({ type: "close", reason: "WebSocket error" }));
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
      if (status === "in_progress") {
        newStats.inProgress += 1;
      } else if (status === "completed") {
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
