import React from "react";
import {
  FaSearch,
  FaCheckCircle,
  FaTimesCircle,
  FaSpinner,
} from "react-icons/fa";
import { useTaskStore } from "../store/taskStore";

const Navbar: React.FC = () => {
  const [searchTerm, setSearchTerm] = React.useState("");
  const taskStats = useTaskStore((state) => state.taskStats);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    // Implementar la lógica de búsqueda aquí
    console.log("Searching for:", searchTerm);
  };

  return (
    <nav className="bg-white shadow-md p-4">
      <div className="flex justify-between items-center">
        <form onSubmit={handleSearch} className="flex-grow max-w-xl">
          <div className="relative">
            <input
              type="text"
              placeholder="Search tasks..."
              className="w-full pl-10 pr-4 py-2 rounded-lg border focus:outline-none focus:border-blue-500"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
            <FaSearch className="absolute left-3 top-3 text-gray-400" />
          </div>
        </form>
        <div className="flex space-x-4 ml-4">
          <div className="flex items-center">
            <FaSpinner className="text-blue-500 mr-1 animate-spin" />
            <span>{taskStats.inProgress}</span>
          </div>
          <div className="flex items-center">
            <FaCheckCircle className="text-green-500 mr-1" />
            <span>{taskStats.completed}</span>
          </div>
          <div className="flex items-center">
            <FaTimesCircle className="text-red-500 mr-1" />
            <span>{taskStats.failed}</span>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
