from pydantic import BaseModel, Field
from typing import Optional, List, Dict
from datetime import datetime
from enum import Enum
import uuid

from src.infrastructure.workers.worker_types import WorkerType

class TaskStatus(str, Enum):
    PENDING = 'Pending'
    IN_PROGRESS = 'InProgress'
    COMPLETED = 'Completed'
    FAILED = 'Failed'
    SCHEDULED = 'Scheduled'
    PENDING_VALIDATION = 'PendingValidation'

class Form(BaseModel):
    id: str
    name: str
    fields: Dict[str, str]

class Approval(BaseModel):
    id: str
    user_id: str
    approved: bool
    approval_date: Optional[datetime]

class TaskExecuted(BaseModel):
    id: str
    run_at: datetime
    workspace_id: str
    done: bool
    status: TaskStatus

class Task(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    create_at: datetime = Field(default_factory=datetime.utcnow)
    workspace_id: str
    name: str
    task_type: str
    technology: str
    worker_type: Optional[WorkerType] = None
    description: Optional[str] = None
    extended_info: Optional[str] = None
    tags: List[str] = []
    forms: List[Form] = []
    approvals: List[Approval] = []
    metadata: Dict[str, str] = {}
    tasks_executed: List[TaskExecuted] = []

class TaskCreate(BaseModel):
    workspace_id: str
    name: str
    task_type: str
    technology: str
    description: Optional[str] = None
    extended_info: Optional[str] = None
    tags: List[str] = []
    forms: List[Form] = []
    approvals: List[Approval] = []
    metadata: Dict[str, str] = {}

class TaskUpdate(BaseModel):
    title: Optional[str] = None
    task_type: Optional[str] = None
    technology: Optional[str] = None
    description: Optional[str] = None
    extended_info: Optional[str] = None
    tags: Optional[List[str]] = None
    forms: Optional[List[Form]] = None
    approvals: Optional[List[Approval]] = None
    metadata: Optional[Dict[str, str]] = None
    tasks_executed: Optional[List[TaskExecuted]] = None
