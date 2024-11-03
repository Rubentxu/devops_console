export interface Task {
  id: string;
  title: string;
  description: string;
  task_type: string;
  technology: string;
  tags: string[];
  metadata: Record<string, string>;
  form?: any;
   status: TaskStatus;
    executions?: TaskExecution[];
}

export interface TaskExecution {
  id: string;
  taskId: string;
  status: "running" | "completed" | "failed";
  logs: string[];
  error?: string;
}

export interface TaskStatus {
  inProgress: number;
  completed: number;
  failed: number;
  running: number;
  pending: number;
  paused: number;
  unknown: number;
}

export interface TaskType {
  id: string;
  name: string;
  description: string;
  form?: any;
}
