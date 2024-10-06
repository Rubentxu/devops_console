from pydantic import BaseModel
from datetime import datetime
from typing import Optional

class Workspace(BaseModel):
    id: str
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
