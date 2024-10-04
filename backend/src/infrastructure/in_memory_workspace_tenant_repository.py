from typing import List, Optional
from src.domain.workspace_tenant import Tenant, TenantCreate, TenantUpdate, Workspace, WorkspaceCreate, WorkspaceUpdate
from src.domain.workspace_tenant_repository import TenantRepository, WorkspaceRepository

class InMemoryTenantRepository(TenantRepository):
    def __init__(self):
        self.tenants: List[Tenant] = []
        self.next_id = 1

    def create(self, tenant: TenantCreate) -> Tenant:
        new_tenant = Tenant(id=self.next_id, **tenant.dict())
        self.tenants.append(new_tenant)
        self.next_id += 1
        return new_tenant

    def get_all(self) -> List[Tenant]:
        return self.tenants

    def get_by_id(self, tenant_id: int) -> Optional[Tenant]:
        return next((tenant for tenant in self.tenants if tenant.id == tenant_id), None)

    def update(self, tenant_id: int, tenant_update: TenantUpdate) -> Optional[Tenant]:
        tenant = self.get_by_id(tenant_id)
        if tenant:
            update_data = tenant_update.dict(exclude_unset=True)
            updated_tenant = tenant.copy(update=update_data)
            self.tenants = [updated_tenant if t.id == tenant_id else t for t in self.tenants]
            return updated_tenant
        return None

    def delete(self, tenant_id: int) -> bool:
        initial_length = len(self.tenants)
        self.tenants = [tenant for tenant in self.tenants if tenant.id != tenant_id]
        return len(self.tenants) < initial_length

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

    def get_by_tenant(self, tenant_id: int) -> List[Workspace]:
        return [workspace for workspace in self.workspaces if workspace.tenant_id == tenant_id]

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
