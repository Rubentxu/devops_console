from abc import ABC, abstractmethod
from typing import List, Optional
from .workspace import Workspace, WorkspaceCreate, WorkspaceUpdate

class WorkspaceRepository(ABC):
    @abstractmethod
    def create(self, workspace: WorkspaceCreate) -> Workspace:
        pass

    @abstractmethod
    def get_all(self) -> List[Workspace]:
        pass

    @abstractmethod
    def get_by_id(self, workspace_id: str) -> Optional[Workspace]:
        pass

    @abstractmethod
    def update(self, workspace_id: str, workspace_update: WorkspaceUpdate) -> Optional[Workspace]:
        pass

    @abstractmethod
    def delete(self, workspace_id: str) -> bool:
        pass
