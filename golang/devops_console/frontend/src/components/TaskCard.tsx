import React, { useState } from "react";
import {
  FaRocket,
  FaFlask,
  FaChartBar,
  FaFolder,
  FaPlay,
} from "react-icons/fa";
import TaskWizard from "./TaskWizard";

import { useTaskStore } from "../store/taskStore";
import { Task } from "../types/taskTypes";

interface TaskCardProps {
  task: Task;
  themeColor: string;
}

export const TaskCard: React.FC<TaskCardProps> = ({ task, themeColor }) => {
  const [isFlipped, setIsFlipped] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const setCurrentTask = useTaskStore((state) => state.setCurrentTask);

  const flipCard = () => {
    if (!isExpanded) {
      setIsFlipped(!isFlipped);
    }
  };

  const handleExecute = () => {
    setCurrentTask(task);
    setIsExpanded(true);
  };

  const getIcon = (taskType: string) => {
    const icons: Record<string, React.ReactElement> = {
      deployment: <FaRocket />,
      testing: <FaFlask />,
      monitoring: <FaChartBar />,
    };
    return icons[taskType] || <FaFolder />;
  };

  const getIconColor = (taskType: string) => {
    const colors: Record<string, string> = {
      deployment: "text-blue-500",
      testing: "text-green-500",
      monitoring: "text-yellow-500",
    };
    return colors[taskType] || "text-gray-500";
  };

  return (
    <div
      className={`relative w-full ${isExpanded ? "h-auto" : "h-56"} perspective-1000`}
    >
      <div
        className={`w-full h-full transition-transform duration-500 transform-style-3d ${
          isFlipped ? "rotate-y-180" : ""
        }`}
      >
        {/* Front of the card */}
        <div
          className={`absolute w-full h-full backface-hidden rounded-lg shadow-lg p-4 ${themeColor} overflow-hidden group ${
            isExpanded ? "hidden" : ""
          }`}
        >
          <div className="absolute -bottom-6 -right-6 text-[150px] opacity-20 transform rotate-12 transition-all duration-300 group-hover:scale-150 group-hover:rotate-[-5deg]">
            <span className={getIconColor(task.task_type)}>
              {getIcon(task.task_type)}
            </span>
          </div>
          <div className="relative z-10">
            <h3 className="text-lg font-semibold mb-2 line-clamp-1">
              {task.title}
            </h3>
            <p className="text-gray-600 mb-2 text-sm line-clamp-2">
              {task.description}
            </p>
            <div className="flex flex-wrap gap-1 mb-2">
              {task.tags.slice(0, 3).map((tag, index) => (
                <span
                  key={index}
                  className="px-2 py-0.5 bg-white bg-opacity-30 text-gray-700 rounded-full text-xs"
                >
                  {tag}
                </span>
              ))}
              {task.tags.length > 3 && (
                <span className="px-2 py-0.5 bg-white bg-opacity-30 text-gray-700 rounded-full text-xs">
                  +{task.tags.length - 3}
                </span>
              )}
            </div>
          </div>
          <button
            onClick={flipCard}
            className="absolute bottom-2 right-2 bg-white text-gray-700 px-3 py-1 rounded text-sm hover:bg-opacity-80 transition-colors"
          >
            More Info
          </button>
          <button
            onClick={handleExecute}
            className="absolute bottom-2 left-2 bg-white text-gray-700 px-3 py-1 rounded text-sm hover:bg-opacity-80 transition-colors flex items-center"
          >
            <FaPlay className="mr-1" /> Execute
          </button>
        </div>

        {/* Back of the card */}
        <div
          className={`absolute w-full h-full backface-hidden rounded-lg shadow-lg p-4 ${themeColor} rotate-y-180 overflow-y-auto ${
            isExpanded ? "hidden" : ""
          }`}
        >
          <h4 className="text-md font-semibold mb-2">Additional Information</h4>
          <p className="text-sm">
            <strong>Technology:</strong> {task.technology}
          </p>
          <p className="text-sm">
            <strong>Type:</strong> {task.task_type}
          </p>
          <h5 className="font-semibold mt-2 mb-1 text-sm">Metadata:</h5>
          <ul className="text-xs">
            {Object.entries(task.metadata).map(([key, value]) => (
              <li key={key}>
                <strong>{key}:</strong> {value}
              </li>
            ))}
          </ul>
          <button
            onClick={flipCard}
            className="absolute bottom-2 right-2 bg-white text-gray-700 px-3 py-1 rounded text-sm hover:bg-opacity-80 transition-colors"
          >
            Back
          </button>
        </div>
      </div>

      {/* Expanded view for wizard */}
      {isExpanded && (
        <div className="absolute top-0 left-0 w-full bg-white rounded-lg shadow-lg p-6 z-20">
          <TaskWizard task={task} onClose={() => setIsExpanded(false)} />
        </div>
      )}
    </div>
  );
};
