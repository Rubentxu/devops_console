from typing import List, Optional
from src.domain.task import Task, TaskCreate, TaskUpdate
import uuid

class InMemoryTaskRepository:
    def __init__(self):
        self.tasks: List[Task] = []

    def create(self, task: TaskCreate) -> Task:
        new_task = Task(
            id=str(uuid.uuid4()),
            **task.dict()
        )
        self.tasks.append(new_task)
        return new_task

    def get_all(self) -> List[Task]:
        return self.tasks

    def get_by_id(self, task_id: str) -> Optional[Task]:
        return next((task for task in self.tasks if task.id == task_id), None)

    def update(self, task_id: str, task_update: TaskUpdate) -> Optional[Task]:
        task = self.get_by_id(task_id)
        if task:
            update_data = task_update.dict(exclude_unset=True)
            updated_task = task.copy(update=update_data)
            self.tasks = [updated_task if t.id == task_id else t for t in self.tasks]
            return updated_task
        return None

    def delete(self, task_id: str) -> bool:
        initial_length = len(self.tasks)
        self.tasks = [task for task in self.tasks if task.id != task_id]
        return len(self.tasks) < initial_length
