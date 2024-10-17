import React from "react";
import { TaskCard } from "./TaskCard";
import { Task } from "../types";

interface TaskGridProps {
  tasks: Task[];
}

const TaskGrid: React.FC<TaskGridProps> = ({ tasks }) => {
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
