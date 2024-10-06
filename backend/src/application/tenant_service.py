from src.domain.tenant.tenant import Tenant, TenantCreate, TenantUpdate
from src.domain.tenant.tenant_repository import TenantRepository
from typing import List, Optional

class TenantService:
    def __init__(self, tenant_repository: TenantRepository):
        self.tenant_repository = tenant_repository

    def create_tenant(self, tenant: TenantCreate) -> Tenant:
        return self.tenant_repository.create(tenant)

    def get_all_tenants(self) -> List[Tenant]:
        return self.tenant_repository.get_all()

    def get_tenant_by_id(self, tenant_id: str) -> Optional[Tenant]:
        return self.tenant_repository.get_by_id(tenant_id)

    def update_tenant(self, tenant_id: str, tenant_update: TenantUpdate) -> Optional[Tenant]:
        return self.tenant_repository.update(tenant_id, tenant_update)

    def delete_tenant(self, tenant_id: str) -> bool:
        return self.tenant_repository.delete(tenant_id)
