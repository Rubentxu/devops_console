import React, { useEffect } from "react";
import { useParams } from "react-router-dom";
import { useTaskStore } from "../store/taskStore";

const TaskExecution: React.FC = () => {
  const { taskId } = useParams<{ taskId: string }>();
  const { taskExecution, updateTaskExecution } = useTaskStore();

  useEffect(() => {
    // Simular la recepción de logs (en un caso real, esto sería un WebSocket)
    const interval = setInterval(() => {
      updateTaskExecution(
        `[${new Date().toISOString()}] Task ${taskId}: ${Math.random().toString(36).substring(7)}`,
      );
    }, 1000);

    return () => clearInterval(interval);
  }, [taskId, updateTaskExecution]);

  if (!taskExecution) return <div>Loading...</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Task Execution: {taskId}</h1>
      <div className="bg-black text-white p-4 rounded h-96 overflow-y-auto font-mono">
        {taskExecution.logs.map((log, index) => (
          <div key={index} className={getLogColor(log)}>
            {log}
          </div>
        ))}
      </div>
    </div>
  );
};

const getLogColor = (log: string) => {
  if (log.includes("ERROR")) return "text-red-500";
  if (log.includes("WARNING")) return "text-yellow-500";
  if (log.includes("DEBUG")) return "text-gray-500";
  return "text-green-500";
};

export default TaskExecution;
