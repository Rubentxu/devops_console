from abc import ABC, abstractmethod
from typing import List, Optional
from .task import Task, TaskCreate, TaskUpdate

class TaskRepository(ABC):
    @abstractmethod
    def create(self, task: TaskCreate) -> Task:
        pass

    @abstractmethod
    def get_all(self) -> List[Task]:
        pass

    @abstractmethod
    def get_by_id(self, task_id: int) -> Optional[Task]:
        pass

    @abstractmethod
    def update(self, task_id: int, task_update: TaskUpdate) -> Optional[Task]:
        pass

    @abstractmethod
    def delete(self, task_id: int) -> bool:
        pass
