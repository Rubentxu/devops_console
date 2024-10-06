import React, { useState } from "react";
import { useWorkspaceStore } from "../store/workspaceStore";

export const WorkspaceForm: React.FC = () => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const { createWorkspace } = useWorkspaceStore();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createWorkspace({ name, description });
    setName("");
    setDescription("");
  };

  return (
    <form onSubmit={handleSubmit} className="mb-6">
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Workspace Name"
        className="mr-2 p-2 border rounded"
        required
      />
      <input
        type="text"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        placeholder="Description"
        className="mr-2 p-2 border rounded"
      />
      <button type="submit" className="p-2 bg-blue-500 text-white rounded">
        Create Workspace
      </button>
    </form>
  );
};

export default WorkspaceForm;
