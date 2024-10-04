from abc import ABC, abstractmethod
from typing import List, Optional
from .workspace_tenant import Tenant, TenantCreate, TenantUpdate, Workspace, WorkspaceCreate, WorkspaceUpdate

class TenantRepository(ABC):
    @abstractmethod
    def create(self, tenant: TenantCreate) -> Tenant:
        pass

    @abstractmethod
    def get_all(self) -> List[Tenant]:
        pass

    @abstractmethod
    def get_by_id(self, tenant_id: int) -> Optional[Tenant]:
        pass

    @abstractmethod
    def update(self, tenant_id: int, tenant_update: TenantUpdate) -> Optional[Tenant]:
        pass

    @abstractmethod
    def delete(self, tenant_id: int) -> bool:
        pass

class WorkspaceRepository(ABC):
    @abstractmethod
    def create(self, workspace: WorkspaceCreate) -> Workspace:
        pass

    @abstractmethod
    def get_all(self) -> List[Workspace]:
        pass

    @abstractmethod
    def get_by_id(self, workspace_id: int) -> Optional[Workspace]:
        pass

    @abstractmethod
    def get_by_tenant(self, tenant_id: int) -> List[Workspace]:
        pass

    @abstractmethod
    def update(self, workspace_id: int, workspace_update: WorkspaceUpdate) -> Optional[Workspace]:
        pass

    @abstractmethod
    def delete(self, workspace_id: int) -> bool:
        pass
