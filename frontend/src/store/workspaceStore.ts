import { create } from "zustand";
import { config } from "../config";
import { Workspace, WorkspaceCreate } from "../types/workspaceTypes";

interface WorkspaceStore {
  workspaces: Workspace[];
  currentWorkspace: Workspace | null;
  fetchWorkspaces: () => Promise<void>;
  setCurrentWorkspace: (workspace: Workspace | null) => void;
  createWorkspace: (workspaceData: WorkspaceCreate) => Promise<void>;
}

export const useWorkspaceStore = create<WorkspaceStore>((set) => ({
  workspaces: [],
  currentWorkspace: null,

  fetchWorkspaces: async () => {
    try {
      const response = await fetch(`${config.apiUrl}/workspaces`);
      if (!response.ok) {
        throw new Error("Failed to fetch workspaces");
      }
      const workspaces = await response.json();
      set({ workspaces });
    } catch (error) {
      console.error("Error fetching workspaces:", error);
    }
  },

  setCurrentWorkspace: (workspace) => set({ currentWorkspace: workspace }),

  createWorkspace: async (workspaceData) => {
    try {
      const response = await fetch(`${config.apiUrl}/workspaces`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(workspaceData),
      });
      if (!response.ok) {
        throw new Error("Failed to create workspace");
      }
      const newWorkspace = await response.json();
      set((state) => ({ workspaces: [...state.workspaces, newWorkspace] }));
    } catch (error) {
      console.error("Error creating workspace:", error);
    }
  },
}));
