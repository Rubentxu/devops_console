from pydantic import BaseModel
from typing import Optional, Dict, Any
from pydantic import Field
import uuid

from src.infrastructure.workers.factories.worker_factory import WorkerType

class Worker(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    name: str
    type: str
    config: Dict[str, Any]

class WorkerCreate(BaseModel):
    name: str
    type: WorkerType
    config: Dict[str, Any]

class WorkerUpdate(BaseModel):
    name: Optional[str] = None
    type: Optional[str] = None
    config: Optional[Dict[str, Any]] = None
