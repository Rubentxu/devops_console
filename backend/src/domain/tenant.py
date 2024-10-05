from pydantic import BaseModel
from datetime import datetime
from typing import Optional, List

class Tenant(BaseModel):
    id: Optional[int] = None
    name: str
    description: str
    created_at: datetime = datetime.now()

class TenantCreate(BaseModel):
    name: str
    description: str

class TenantUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None

class Workspace(BaseModel):
    id: Optional[int] = None
    name: str
    description: str
    tenant_id: int
    created_at: datetime = datetime.now()

class WorkspaceCreate(BaseModel):
    name: str
    description: str
    tenant_id: int

class WorkspaceUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    tenant_id: Optional[int] = None
