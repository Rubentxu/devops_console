import React, { useEffect } from "react";
import TaskGrid from "../components/TaskGrid";
import { useTaskStore } from "../store/taskStore";
import { config } from "../config";

const Tasks: React.FC = () => {
  const { tasks, fetchTasks } = useTaskStore();

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Tasks</h1>
      <TaskGrid tasks={tasks} />
    </div>
  );
};

export default Tasks;
