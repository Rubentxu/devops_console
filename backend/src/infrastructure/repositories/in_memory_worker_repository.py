from typing import List, Optional
from src.domain.worker.worker import Worker, WorkerCreate, WorkerUpdate
from src.domain.worker.worker_repository import WorkerRepository
import uuid

class InMemoryWorkerRepository(WorkerRepository):
    def __init__(self):
        self.workers: List[Worker] = []

    def create(self, worker: WorkerCreate) -> Worker:
        new_worker = Worker(id=str(uuid.uuid4()), **worker.dict())
        self.workers.append(new_worker)
        return new_worker

    def get_all(self) -> List[Worker]:
        return self.workers

    def get_by_id(self, worker_id: str) -> Optional[Worker]:
        return next((worker for worker in self.workers if worker.id == worker_id), None)

    def update(self, worker_id: str, worker_update: WorkerUpdate) -> Optional[Worker]:
        worker = self.get_by_id(worker_id)
        if worker:
            update_data = worker_update.dict(exclude_unset=True)
            updated_worker = worker.copy(update=update_data)
            self.workers = [updated_worker if w.id == worker_id else w for w in self.workers]
            return updated_worker
        return None

    def delete(self, worker_id: str) -> bool:
        initial_length = len(self.workers)
        self.workers = [worker for worker in self.workers if worker.id != worker_id]
        return len(self.workers) < initial_length