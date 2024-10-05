from typing import List, Optional
from src.domain.workspace import Workspace, WorkspaceCreate, WorkspaceUpdate
from src.domain.workspace_repository import WorkspaceRepository

class InMemoryWorkspaceRepository(WorkspaceRepository):
    def __init__(self):
        self.workspaces: List[Workspace] = []
        self.next_id = 1

    def create(self, workspace: WorkspaceCreate) -> Workspace:
        new_workspace = Workspace(id=self.next_id, **workspace.dict())
        self.workspaces.append(new_workspace)
        self.next_id += 1
        return new_workspace

    def get_all(self) -> List[Workspace]:
        return self.workspaces

    def get_by_id(self, workspace_id: int) -> Optional[Workspace]:
        return next((workspace for workspace in self.workspaces if workspace.id == workspace_id), None)

    def update(self, workspace_id: int, workspace_update: WorkspaceUpdate) -> Optional[Workspace]:
        workspace = self.get_by_id(workspace_id)
        if workspace:
            update_data = workspace_update.dict(exclude_unset=True)
            updated_workspace = workspace.copy(update=update_data)
            self.workspaces = [updated_workspace if w.id == workspace_id else w for w in self.workspaces]
            return updated_workspace
        return None

    def delete(self, workspace_id: int) -> bool:
        initial_length = len(self.workspaces)
        self.workspaces = [workspace for workspace in self.workspaces if workspace.id != workspace_id]
        return len(self.workspaces) < initial_length
