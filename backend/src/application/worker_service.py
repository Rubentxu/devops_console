from src.domain.worker.worker_repository import WorkerRepository
from src.domain.worker.worker import Worker, WorkerCreate, WorkerUpdate
from typing import List, Optional

class WorkerService:
    def __init__(self, worker_repository: WorkerRepository):
        self.worker_repository = worker_repository

    def create_worker(self, worker: WorkerCreate) -> Worker:
        return self.worker_repository.create(worker)

    def get_all_workers(self) -> List[Worker]:
        return self.worker_repository.get_all()

    def get_worker_by_id(self, worker_id: str) -> Optional[Worker]:
        return self.worker_repository.get_by_id(worker_id)

    def update_worker(self, worker_id: str, worker_update: WorkerUpdate) -> Optional[Worker]:
        return self.worker_repository.update(worker_id, worker_update)

    def delete_worker(self, worker_id: str) -> bool:
        return self.worker_repository.delete(worker_id)
