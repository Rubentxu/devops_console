from src.domain.task import Task, TaskCreate, TaskUpdate, TaskExecuted
from src.infrastructure.in_memory_task_repository import InMemoryTaskRepository
from typing import List, Optional

class TaskService:
    def __init__(self, task_repository: InMemoryTaskRepository):
        self.task_repository = task_repository

    def create_task(self, task: TaskCreate) -> Task:
        return self.task_repository.create(task)

    def get_all_tasks(self) -> List[Task]:
        return self.task_repository.get_all()

    def get_task_by_id(self, task_id: str) -> Optional[Task]:
        return self.task_repository.get_by_id(task_id)

    def update_task(self, task_id: str, task_update: TaskUpdate) -> Optional[Task]:
        return self.task_repository.update(task_id, task_update)

    def delete_task(self, task_id: str) -> bool:
        return self.task_repository.delete(task_id)

    def add_task_execution(self, task_id: str, task_executed: TaskExecuted) -> Optional[Task]:
        task = self.get_task_by_id(task_id)
        if task:
            task.tasks_executed.append(task_executed)
            return self.update_task(task_id, TaskUpdate(tasks_executed=task.tasks_executed))
        return None
