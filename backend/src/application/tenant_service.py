from src.domain.workspace import Tenant, TenantCreate, TenantUpdate, Workspace, WorkspaceCreate, WorkspaceUpdate
from src.domain.workspace_repository import TenantRepository, WorkspaceRepository
from typing import List, Optional

class TenantService:
    def __init__(self, tenant_repository: TenantRepository):
        self.tenant_repository = tenant_repository

    def create_tenant(self, tenant: TenantCreate) -> Tenant:
        return self.tenant_repository.create(tenant)

    def get_all_tenants(self) -> List[Tenant]:
        return self.tenant_repository.get_all()

    def get_tenant_by_id(self, tenant_id: int) -> Optional[Tenant]:
        return self.tenant_repository.get_by_id(tenant_id)

    def update_tenant(self, tenant_id: int, tenant_update: TenantUpdate) -> Optional[Tenant]:
        return self.tenant_repository.update(tenant_id, tenant_update)

    def delete_tenant(self, tenant_id: int) -> bool:
        return self.tenant_repository.delete(tenant_id)

class WorkspaceService:
    def __init__(self, workspace_repository: WorkspaceRepository):
        self.workspace_repository = workspace_repository

    def create_workspace(self, workspace: WorkspaceCreate) -> Workspace:
        return self.workspace_repository.create(workspace)

    def get_all_workspaces(self) -> List[Workspace]:
        return self.workspace_repository.get_all()

    def get_workspace_by_id(self, workspace_id: int) -> Optional[Workspace]:
        return self.workspace_repository.get_by_id(workspace_id)

    def get_workspaces_by_tenant(self, tenant_id: int) -> List[Workspace]:
        return self.workspace_repository.get_by_tenant(tenant_id)

    def update_workspace(self, workspace_id: int, workspace_update: WorkspaceUpdate) -> Optional[Workspace]:
        return self.workspace_repository.update(workspace_id, workspace_update)

    def delete_workspace(self, workspace_id: int) -> bool:
        return self.workspace_repository.delete(workspace_id)
