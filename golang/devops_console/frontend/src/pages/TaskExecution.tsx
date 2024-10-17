import React, { useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTaskStore } from "../store/taskStore";
import { toast } from "react-toastify";
import { FaSpinner, FaCheckCircle, FaTimesCircle } from "react-icons/fa";

const TaskExecution: React.FC = () => {
  const { taskId } = useParams<{ taskId: string }>();
  const navigate = useNavigate();
  const { taskExecution, connectWebSocket, disconnectWebSocket } =
    useTaskStore();
  const [isFinished, setIsFinished] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (taskId) {
      setIsLoading(true);
      connectWebSocket(taskId).finally(() => {
        console.log("WebSocket finished");
        setIsLoading(false);
      });
    }

    return () => {
      disconnectWebSocket();
    };
  }, [taskId, connectWebSocket, disconnectWebSocket]);

  useEffect(() => {
    console.log("TaskExecution status: ", taskExecution?.status);
    if (
      taskExecution?.status === "completed" ||
      taskExecution?.status === "closed" ||
      taskExecution?.status === "failed"
    ) {
      setIsFinished(true);
      if (taskExecution.status === "failed" && taskExecution.error) {
        toast.error(`Task failed: ${taskExecution.error}`);
      }
    }
  }, [taskExecution?.status, taskExecution?.error]);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [taskExecution?.logs]);

  const handleBackToDashboard = () => {
    navigate("/");
  };

  const getLogColor = (log: string) => {
    if (log.includes("ERROR")) return "text-red-500";
    if (log.includes("WARNING")) return "text-yellow-500";
    if (log.includes("DEBUG")) return "text-gray-500";
    return "text-green-500";
  };

  return (
    <div className="flex flex-col h-full">
      <div className="p-6 flex-grow flex flex-col">
        <h1 className="text-2xl font-bold mb-4">Task Execution: {taskId}</h1>
        <div
          className="flex-grow flex flex-col bg-black rounded-lg shadow-lg overflow-hidden"
          style={{ maxHeight: "calc(100vh - 240px)" }}
        >
          <div className="bg-gray-800 text-white px-4 py-2 flex justify-between items-center">
            <span>Console Output</span>
            {!isFinished && <FaSpinner className="animate-spin" />}
          </div>
          <div
            ref={terminalRef}
            className="flex-grow p-4 font-mono text-sm text-white overflow-y-auto"
          >
            {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <FaSpinner className="animate-spin text-4xl" />
              </div>
            ) : (
              <>
                {taskExecution?.logs.map((log, index) => (
                  <div key={index} className={`${getLogColor(log)} mb-1`}>
                    {log}
                  </div>
                ))}
                {!isFinished && (
                  <div className="flex items-center mt-2">
                    <FaSpinner className="animate-spin mr-2" />
                    <span className="text-gray-400">
                      Waiting for more data...
                    </span>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>
      <div className="p-6 bg-gray-100 border-t border-gray-200">
        {isFinished ? (
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              {taskExecution?.status === "completed" ? (
                <FaCheckCircle className="text-green-500 text-2xl mr-2" />
              ) : (
                <FaTimesCircle className="text-red-500 text-2xl mr-2" />
              )}
              <span
                className={`text-lg font-semibold ${taskExecution?.status === "completed" ? "text-green-600" : "text-red-600"}`}
              >
                Task {taskExecution?.status}
              </span>
            </div>
            <button
              onClick={handleBackToDashboard}
              className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
            >
              Back to Dashboard
            </button>
          </div>
        ) : (
          <div className="flex items-center justify-center">
            <FaSpinner className="animate-spin text-blue-500 mr-2" />
            <span>Task in progress...</span>
          </div>
        )}
        {taskExecution?.error && (
          <div
            className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mt-4"
            role="alert"
          >
            <strong className="font-bold">Error: </strong>
            <span className="block sm:inline">{taskExecution.error}</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default TaskExecution;
