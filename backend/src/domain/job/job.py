from pydantic import BaseModel
from typing import Optional, Dict
from datetime import datetime
import uuid
from enum import Enum
from typing import Any
from pydantic import Field

class JobStatus(str, Enum):
    PENDING = 'Pending'
    RUNNING = 'Running'
    COMPLETED = 'Completed'
    FAILED = 'Failed'

class Job(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    worker_id: str
    name: str
    status: JobStatus = JobStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.utcnow)
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    result: Optional[str] = None
    metadata: Dict[str, Any] = {}

class JobCreate(BaseModel):
    worker_id: str
    name: str
    metadata: Dict[str, Any] = {}

class JobUpdate(BaseModel):
    status: Optional[JobStatus] = None
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    result: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None
