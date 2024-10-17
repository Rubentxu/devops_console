export interface Task {
  id: string;
  title: string;
  description: string;
  task_type: string;
  technology: string;
  tags: string[];
  metadata: Record<string, string>;
  form?: any;
}

export interface TaskExecution {
  id: string;
  taskId: string;
  status: "running" | "completed" | "failed";
  logs: string[];
}

export interface TaskStatus {
  inProgress: number;
  completed: number;
  failed: number;
}

export interface TaskType {
  id: string;
  name: string;
  description: string;
  form?: any;
}
