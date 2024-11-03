export interface Workspace {
  id: string;
  name: string;
  description: string;
  created_at: string;
}

export interface WorkspaceCreate {
  name: string;
  description: string;
}
