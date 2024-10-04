import create from "zustand";
import { Task, TaskExecution } from "../types";
import { config } from "../config";

interface TaskStore {
  tasks: Task[];
  currentTask: Task | null;
  taskExecution: TaskExecution | null;
  fetchTasks: () => Promise<void>;
  setCurrentTask: (task: Task | null) => void;
  executeTask: (taskId: string, formData: Record<string, any>) => Promise<void>;
  updateTaskExecution: (log: string) => void;
}

export const useTaskStore = create<TaskStore>((set, get) => ({
  tasks: [],
  currentTask: null,
  taskExecution: null,

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
          id: executionData.id,
          status: executionData.status,
          logs: [],
        },
        taskExecutionUrl: `/task-execution/${taskId}`,
      });
    } catch (error) {
      console.error("Error executing task:", error);
      throw error;
    }
  },

  updateTaskExecution: (log) => {
    set((state) => ({
      taskExecution: state.taskExecution
        ? { ...state.taskExecution, logs: [...state.taskExecution.logs, log] }
        : null,
    }));
  },
}));
