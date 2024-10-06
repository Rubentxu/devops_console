from fastapi import APIRouter, HTTPException, Depends
from typing import List
from src.domain.tenant.tenant import Tenant, TenantCreate, TenantUpdate
from src.application.tenant_service import TenantService

def create_tenant_router(tenant_service: TenantService):
    router = APIRouter()

    @router.post("/tenants/", response_model=Tenant)
    async def create_tenant(tenant: TenantCreate):
        return tenant_service.create_tenant(tenant)

    @router.get("/tenants/", response_model=List[Tenant])
    async def read_tenants():
        return tenant_service.get_all_tenants()

    @router.get("/tenants/{tenant_id}", response_model=Tenant)
    async def read_tenant(tenant_id: str):
        tenant = tenant_service.get_tenant_by_id(tenant_id)
        if tenant is None:
            raise HTTPException(status_code=404, detail="Tenant not found")
        return tenant

    @router.put("/tenants/{tenant_id}", response_model=Tenant)
    async def update_tenant(tenant_id: str, tenant_update: TenantUpdate):
        updated_tenant = tenant_service.update_tenant(tenant_id, tenant_update)
        if updated_tenant is None:
            raise HTTPException(status_code=404, detail="Tenant not found")
        return updated_tenant

    @router.delete("/tenants/{tenant_id}", response_model=bool)
    async def delete_tenant(tenant_id: str):
        deleted = tenant_service.delete_tenant(tenant_id)
        if not deleted:
            raise HTTPException(status_code=404, detail="Tenant not found")
        return True

    return router
