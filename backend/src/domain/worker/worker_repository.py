from abc import ABC, abstractmethod
from typing import List, Optional
from .worker import Worker, WorkerCreate, WorkerUpdate

class WorkerRepository(ABC):
    @abstractmethod
    def create(self, worker: WorkerCreate) -> Worker:
        pass

    @abstractmethod
    def get_all(self) -> List[Worker]:
        pass

    @abstractmethod
    def get_by_id(self, worker_id: str) -> Optional[Worker]:
        pass

    @abstractmethod
    def update(self, worker_id: str, worker_update: WorkerUpdate) -> Optional[Worker]:
        pass

    @abstractmethod
    def delete(self, worker_id: str) -> bool:
        pass
