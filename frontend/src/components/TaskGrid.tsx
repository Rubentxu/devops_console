import React from "react";
import { TaskCard } from "./TaskCard";
import { Task } from "../types/taskTypes";
import { useTaskStore } from "../store/taskStore";

const TaskGrid: React.FC = () => {
  const tasks = useTaskStore((state) => state.tasks);

  const getThemeColor = (taskType: string) => {
    const colors: Record<string, string> = {
      deployment: "bg-blue-100",
      testing: "bg-green-100",
      monitoring: "bg-yellow-100",
    };
    return colors[taskType] || "bg-gray-100";
  };

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {tasks.map((task) => (
        <TaskCard
          key={task.id}
          task={task}
          themeColor={getThemeColor(task.task_type)}
        />
      ))}
    </div>
  );
};

export default TaskGrid;