import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTaskStore } from "../store/taskStore";

interface TaskWizardProps {
  task: Task;
  onClose: () => void;
}

const TaskWizard: React.FC<TaskWizardProps> = ({ task, onClose }) => {
  const [currentStep, setCurrentStep] = useState(0);
  const [formData, setFormData] = useState<Record<string, any>>({});
  const navigate = useNavigate();
  const executeTask = useTaskStore((state) => state.executeTask);

  const steps = task.form?.steps || [];

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else {
      handleExecute();
    }
  };

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleExecute = async () => {
    try {
      await executeTask(task.id, formData);
      navigate(`/task-execution/${task.id}`);
    } catch (error) {
      console.error("Error executing task:", error);
      // Aquí podrías mostrar un mensaje de error al usuario
    }
  };

  return (
    <div className="p-4 bg-white rounded-lg shadow">
      <h3 className="text-lg font-semibold mb-4">Execute Task: {task.title}</h3>
      {steps[currentStep] && (
        <div>
          <h4 className="font-medium mb-2">
            Step {currentStep + 1}: {steps[currentStep].title}
          </h4>
          {steps[currentStep].fields.map((field: any) => (
            <div key={field.name} className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {field.label}
              </label>
              <input
                type={field.type}
                name={field.name}
                value={formData[field.name] || ""}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          ))}
        </div>
      )}
      <div className="flex justify-between mt-6">
        {currentStep > 0 && (
          <button
            onClick={handlePrevious}
            className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
          >
            Previous
          </button>
        )}
        <button
          onClick={handleNext}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          {currentStep === steps.length - 1 ? "Execute" : "Next"}
        </button>
      </div>
    </div>
  );
};

export default TaskWizard;
