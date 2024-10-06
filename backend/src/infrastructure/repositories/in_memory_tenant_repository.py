from typing import List, Optional
import uuid
from src.domain.tenant.tenant import Tenant, TenantCreate, TenantUpdate
from src.domain.tenant.tenant_repository import Tenant, TenantCreate, TenantUpdate, TenantRepository
from src.domain.workspace.workspace_repository import WorkspaceRepository, Workspace, WorkspaceCreate, WorkspaceUpdate

class InMemoryTenantRepository(TenantRepository):
    def __init__(self):
        self.tenants: List[Tenant] = []
        self.next_id = str(uuid.uuid4())

    def create(self, tenant: TenantCreate) -> Tenant:
        new_tenant = Tenant(id=self.next_id, **tenant.dict())
        self.tenants.append(new_tenant)
        self.next_id = str(uuid.uuid4())
        return new_tenant

    def get_all(self) -> List[Tenant]:
        return self.tenants

    def get_by_id(self, tenant_id: str) -> Optional[Tenant]:
        return next((tenant for tenant in self.tenants if tenant.id == tenant_id), None)

    def update(self, tenant_id: str, tenant_update: TenantUpdate) -> Optional[Tenant]:
        tenant = self.get_by_id(tenant_id)
        if tenant:
            update_data = tenant_update.dict(exclude_unset=True)
            updated_tenant = tenant.copy(update=update_data)
            self.tenants = [updated_tenant if t.id == tenant_id else t for t in self.tenants]
            return updated_tenant
        return None

    def delete(self, tenant_id: str) -> bool:
        initial_length = len(self.tenants)
        self.tenants = [tenant for tenant in self.tenants if tenant.id != tenant_id]
        return len(self.tenants) < initial_length
