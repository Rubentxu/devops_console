from abc import ABC, abstractmethod
from typing import List, Optional
from .workspace import Workspace, WorkspaceCreate, WorkspaceUpdate
from .tenant import Tenant, TenantCreate, TenantUpdate

class TenantRepository(ABC):
    @abstractmethod
    def create(self, tenant: TenantCreate) -> Tenant:
        pass

    @abstractmethod
    def get_all(self) -> List[Tenant]:
        pass

    @abstractmethod
    def get_by_id(self, tenant_id: str) -> Optional[Tenant]:
        pass

    @abstractmethod
    def update(self, tenant_id: str, tenant_update: TenantUpdate) -> Optional[Tenant]:
        pass

    @abstractmethod
    def delete(self, tenant_id: str) -> bool:
        pass
