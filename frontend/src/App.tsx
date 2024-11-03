import React from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Navbar from "./components/Navbar";
import Sidebar from "./components/Sidebar";
import Dashboard from "./pages/Dashboard";
import Tasks from "./pages/Tasks";
import Pipelines from "./pages/Pipelines";
import Settings from "./pages/Settings";
import TaskExecution from "./pages/TaskExecution";
import Workspaces from "./pages/Workspaces"; // Importa el nuevo componente Workspaces
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { useTaskStore } from "./store/taskStore";
import { Tenants } from "./pages/Tenants";

function App() {
  const taskStats = useTaskStore((state) => state.taskStats);

  return (
    <Router>
      <div className="flex h-screen bg-gray-100">
        <div className="w-64 bg-white shadow-md hidden lg:block">
          <Sidebar username="John Doe" />
        </div>

        <div className="flex-1 flex flex-col overflow-hidden">
          <Navbar taskStats={taskStats} />
          <main className="flex-1 overflow-x-hidden overflow-y-auto bg-gray-200">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/tasks" element={<Tasks />} />
              <Route path="/pipelines" element={<Pipelines />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="/workspaces" element={<Workspaces />} />{" "}
              <Route path="/tenants" element={<Tenants />} />
              {/* Nueva ruta para Workspaces */}
              <Route
                path="/task-execution/:taskId"
                element={<TaskExecution />}
              />
            </Routes>
          </main>
        </div>
      </div>
      <ToastContainer />
    </Router>
  );
}

export default App;
