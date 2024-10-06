from abc import ABC, abstractmethod
from typing import List, Optional
from .job import Job, JobCreate, JobUpdate

class JobRepository(ABC):
    @abstractmethod
    def create(self, job: JobCreate) -> Job:
        pass

    @abstractmethod
    def get_all(self) -> List[Job]:
        pass

    @abstractmethod
    def get_by_id(self, job_id: str) -> Optional[Job]:
        pass

    @abstractmethod
    def update(self, job_id: str, job_update: JobUpdate) -> Optional[Job]:
        pass

    @abstractmethod
    def delete(self, job_id: str) -> bool:
        pass

    @abstractmethod
    def get_jobs_by_worker_id(self, worker_id: str) -> List[Job]:
        pass