from src.domain.workspace import Workspace, WorkspaceCreate, WorkspaceUpdate
from src.domain.workspace_repository import WorkspaceRepository
from typing import List, Optional

class WorkspaceService:
    def __init__(self, workspace_repository: WorkspaceRepository):
        self.workspace_repository = workspace_repository

    def create_workspace(self, workspace: WorkspaceCreate) -> Workspace:
        return self.workspace_repository.create(workspace)

    def get_all_workspaces(self) -> List[Workspace]:
        return self.workspace_repository.get_all()

    def get_workspace_by_id(self, workspace_id: int) -> Optional[Workspace]:
        return self.workspace_repository.get_by_id(workspace_id)

    def update_workspace(self, workspace_id: int, workspace_update: WorkspaceUpdate) -> Optional[Workspace]:
        return self.workspace_repository.update(workspace_id, workspace_update)

    def delete_workspace(self, workspace_id: int) -> bool:
        return self.workspace_repository.delete(workspace_id)
