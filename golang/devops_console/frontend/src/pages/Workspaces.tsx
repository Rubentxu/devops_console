import React, { useEffect } from "react";
import { useWorkspaceStore } from "../store/workspaceStore";
import WorkspaceGrid from "../components/WorkspaceGrid.tsx";
import WorkspaceForm from "../components/WorkspaceForm.tsx";

const Workspaces: React.FC = () => {
  const { workspaces, fetchWorkspaces } = useWorkspaceStore();

  useEffect(() => {
    fetchWorkspaces();
  }, [fetchWorkspaces]);

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Workspaces</h1>
      <WorkspaceForm />
      <WorkspaceGrid workspaces={workspaces} />
    </div>
  );
};

export default Workspaces;
