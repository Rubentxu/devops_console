import React from "react";
import { Workspace } from "../types/workspaceTypes";

interface WorkspaceGridProps {
  workspaces: Workspace[];
}

const WorkspaceGrid: React.FC<WorkspaceGridProps> = ({ workspaces }) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {workspaces.map((workspace) => (
        <div key={workspace.id} className="border p-4 rounded">
          <h2 className="text-xl font-bold">{workspace.name}</h2>
          <p>{workspace.description}</p>
          <p className="text-sm text-gray-500">
            Created: {new Date(workspace.created_at).toLocaleDateString()}
          </p>
        </div>
      ))}
    </div>
  );
};

export default WorkspaceGrid;
