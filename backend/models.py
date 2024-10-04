from pydantic import BaseModel
from enum import Enum
from typing import Optional

class TaskType(str, Enum):
    PIPELINE = 'pipeline'
    DEPLOYMENT = 'deployment'
    MONITORING = 'monitoring'

class Task(BaseModel):
    id: Optional[int] = None
    title: str
    description: str
    type: TaskType

class TaskCreate(BaseModel):
    title: str
    description: str
    type: TaskType

class TaskUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    type: Optional[TaskType] = None
