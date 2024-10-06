from pydantic import BaseModel
from datetime import datetime
from typing import Optional, List

class Tenant(BaseModel):
    id: str
    name: str
    description: str
    created_at: datetime = datetime.now()

class TenantCreate(BaseModel):
    name: str
    description: str

class TenantUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
