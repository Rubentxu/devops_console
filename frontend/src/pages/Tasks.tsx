import React, { useEffect } from "react";
import TaskGrid from "../components/TaskGrid";
import { useTaskStore } from "../store/taskStore";

const Tasks: React.FC = () => {
  const { fetchTasks, taskStats } = useTaskStore();

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Tasks</h1>
      <div className="mb-4 flex gap-4">
        <div className="bg-blue-100 p-2 rounded">
          <span className="font-bold">In Progress:</span> {taskStats.inProgress}
        </div>
        <div className="bg-green-100 p-2 rounded">
          <span className="font-bold">Completed:</span> {taskStats.completed}
        </div>
        <div className="bg-red-100 p-2 rounded">
          <span className="font-bold">Failed:</span> {taskStats.failed}
        </div>
      </div>
      <TaskGrid />
    </div>
  );
};

export default Tasks;
