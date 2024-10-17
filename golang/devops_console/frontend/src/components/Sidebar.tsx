import React from "react";
import { Link } from "react-router-dom";
import {
  FaUser,
  FaTachometerAlt,
  FaTasks,
  FaStream,
  FaCog,
  FaLayerGroup, // Nuevo ícono para Workspaces
  FaBuilding, // Nuevo ícono para Tenants
} from "react-icons/fa";

interface SidebarProps {
  username: string;
}

const Sidebar: React.FC<SidebarProps> = ({ username }) => {
  return (
    <div className="bg-gray-800 text-white w-64 min-h-screen p-4">
      <div className="flex items-center mb-6">
        <FaUser className="text-2xl mr-2" />
        <span className="text-lg font-semibold">{username}</span>
      </div>
      <ul>
        <li className="mb-4">
          <Link to="/" className="flex items-center hover:text-blue-300">
            <FaTachometerAlt className="mr-2" />
            Dashboard
          </Link>
        </li>
        <li className="mb-4">
          <Link to="/tasks" className="flex items-center hover:text-blue-300">
            <FaTasks className="mr-2" />
            Tasks
          </Link>
        </li>
        <li className="mb-4">
          <Link
            to="/pipelines"
            className="flex items-center hover:text-blue-300"
          >
            <FaStream className="mr-2" />
            Pipelines
          </Link>
        </li>
        <li className="mb-4">
          <Link to="/tenants" className="flex items-center hover:text-blue-300">
            <FaBuilding className="mr-2" />
            Tenants
          </Link>
        </li>
        <li className="mb-4">
          <Link
            to="/workspaces"
            className="flex items-center hover:text-blue-300"
          >
            <FaLayerGroup className="mr-2" />
            Workspaces
          </Link>
        </li>
        <li className="mb-4">
          <Link
            to="/settings"
            className="flex items-center hover:text-blue-300"
          >
            <FaCog className="mr-2" />
            Settings
          </Link>
        </li>
      </ul>
    </div>
  );
};

export default Sidebar;
