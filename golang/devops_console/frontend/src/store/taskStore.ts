import { create } from "zustand";
import { Task, TaskStatus } from "../types/taskTypes";
import { 
  CreateTask, 
  GetAllTasks, 
  GetTaskByID, 
  UpdateTask, 
  DeleteTask, 
  ExecuteTask,
  ExecuteTaskWithTimeout,
  PauseTask,
  ResumeTask,
  GetTaskStatus,
  GetTaskStatistics,
  MonitorTask,
  StreamTaskLogs,
  ExecuteAndMonitorTask
} from '../wailsjs/go/interfaces/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

interface TaskExecution {
  id: string;
  taskId: string;
  status: TaskStatus;
  logs: string[];
  error?: string;
}

interface TaskStore {
  tasks: Task[];
  currentTask: Task | null;
  taskExecution: TaskExecution | null;
  taskStats: {
    inProgress: number;
    completed: number;
    failed: number;
  };
  fetchTasks: () => Promise<void>;
  setCurrentTask: (task: Task | null) => void;
  createTask: (taskCreate: any) => Promise<Task | null>;
  executeTask: (taskId: string) => Promise<void>;
  monitorTask: (taskId: string) => Promise<void>;
  streamTaskLogs: (taskId: string) => Promise<void>;
  executeAndMonitorTask: (taskId: string) => Promise<void>;
  pauseTask: (taskId: string) => Promise<void>;
  resumeTask: (taskId: string) => Promise<void>;
  getTaskStatus: (taskId: string) => Promise<string>;
  updateTaskStats: (status: TaskStatus) => void;
}

export const useTaskStore = create<TaskStore>((set, get) => ({
  tasks: [],
  currentTask: null,
  taskExecution: null,
  taskStats: {
    inProgress: 0,
    completed: 0,
    failed: 0,
  },

  fetchTasks: async () => {
    try {
      const tasks = await GetAllTasks();
      set({ tasks });
    } catch (error) {
      console.error("Error fetching tasks:", error);
    }
  },

  setCurrentTask: (task) => set({ currentTask: task }),

  createTask: async (taskCreate) => {
    try {
      const task = await CreateTask(taskCreate);
      set(state => ({ tasks: [...state.tasks, task] }));
      return task;
    } catch (error) {
      console.error("Error creating task:", error);
      return null;
    }
  },

  executeTask: async (taskId) => {
    try {
      await ExecuteTask(taskId);
      // La ejecución real y el monitoreo se manejan a través de eventos
    } catch (error) {
      console.error("Error executing task:", error);
    }
  },

  monitorTask: async (taskId) => {
    try {
      await MonitorTask(taskId);
      // Los resultados se manejarán a través de eventos
    } catch (error) {
      console.error("Error monitoring task:", error);
    }
  },

  streamTaskLogs: async (taskId) => {
    try {
      await StreamTaskLogs(taskId);
      // Los logs se manejarán a través de eventos
    } catch (error) {
      console.error("Error streaming task logs:", error);
    }
  },

  executeAndMonitorTask: async (taskId) => {
    try {
      await ExecuteAndMonitorTask(taskId);
      // Los resultados y actualizaciones se manejarán a través de eventos
    } catch (error) {
      console.error("Error executing and monitoring task:", error);
    }
  },

  pauseTask: async (taskId) => {
    try {
      await PauseTask(taskId);
    } catch (error) {
      console.error("Error pausing task:", error);
    }
  },

  resumeTask: async (taskId) => {
    try {
      await ResumeTask(taskId);
    } catch (error) {
      console.error("Error resuming task:", error);
    }
  },

  getTaskStatus: async (taskId) => {
    try {
      return await GetTaskStatus(taskId);
    } catch (error) {
      console.error("Error getting task status:", error);
      return "unknown";
    }
  },

  updateTaskStats: (status: TaskStatus) => {
    set((state) => {
      const newStats = { ...state.taskStats };
      if (status === "running") {
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
}));

// Configurar los listeners de eventos
EventsOn("task:created", (task: Task) => {
  useTaskStore.setState(state => ({ tasks: [...state.tasks, task] }));
});

EventsOn("task:updated", (task: Task) => {
  useTaskStore.setState(state => ({
    tasks: state.tasks.map(t => t.id === task.id ? task : t)
  }));
});

EventsOn("task:deleted", (taskId: string) => {
  useTaskStore.setState(state => ({
    tasks: state.tasks.filter(t => t.id !== taskId)
  }));
});

EventsOn("task:execution:added", (data: { taskId: string, execution: any }) => {
  useTaskStore.setState(state => ({
    tasks: state.tasks.map(t => {
      if (t.id === data.taskId) {
        return { ...t, taskExecutions: [...(t.taskExecutions || []), data.execution] };
      }
      return t;
    })
  }));
});

EventsOn("task:execution:result", (data: { taskId: string, result: any }) => {
  useTaskStore.setState(state => ({
    taskExecution: state.taskExecution && state.taskExecution.taskId === data.taskId
      ? { ...state.taskExecution, ...data.result }
      : state.taskExecution
  }));
});

EventsOn("task:execution:error", (data: { taskId: string, error: string }) => {
  useTaskStore.setState(state => ({
    taskExecution: state.taskExecution && state.taskExecution.taskId === data.taskId
      ? { ...state.taskExecution, status: "failed", error: data.error }
      : state.taskExecution
  }));
  useTaskStore.getState().updateTaskStats("failed");
});

EventsOn("task:status:update", (data: { taskId: string, status: TaskStatus }) => {
  useTaskStore.setState(state => ({
    taskExecution: state.taskExecution && state.taskExecution.taskId === data.taskId
      ? { ...state.taskExecution, status: data.status }
      : state.taskExecution
  }));
  useTaskStore.getState().updateTaskStats(data.status);
});

EventsOn("task:log", (data: { taskId: string, log: string }) => {
  useTaskStore.setState(state => ({
    taskExecution: state.taskExecution && state.taskExecution.taskId === data.taskId
      ? { ...state.taskExecution, logs: [...state.taskExecution.logs, data.log] }
      : state.taskExecution
  }));
});

// Asegúrate de limpiar los listeners cuando ya no sean necesarios
// Esto podría hacerse en un componente de nivel superior cuando se desmonte
const cleanupEventListeners = () => {
  EventsOff("task:created");
  EventsOff("task:updated");
  EventsOff("task:deleted");
  EventsOff("task:execution:added");
  EventsOff("task:execution:result");
  EventsOff("task:execution:error");
  EventsOff("task:status:update");
  EventsOff("task:log");
};