from typing import List, Optional
from src.domain.job.job import Job, JobCreate, JobUpdate
from src.domain.job.job_repository import JobRepository
import uuid

class InMemoryJobRepository(JobRepository):
    def __init__(self):
        self.jobs: List[Job] = []

    def create(self, job: JobCreate) -> Job:
        new_job = Job(id=str(uuid.uuid4()), **job.dict())
        self.jobs.append(new_job)
        return new_job

    def get_all(self) -> List[Job]:
        return self.jobs

    def get_by_id(self, job_id: str) -> Optional[Job]:
        return next((job for job in self.jobs if job.id == job_id), None)

    def update(self, job_id: str, job_update: JobUpdate) -> Optional[Job]:
        job = self.get_by_id(job_id)
        if job:
            update_data = job_update.dict(exclude_unset=True)
            updated_job = job.copy(update=update_data)
            self.jobs = [updated_job if j.id == job_id else j for j in self.jobs]
            return updated_job
        return None

    def delete(self, job_id: str) -> bool:
        initial_length = len(self.jobs)
        self.jobs = [job for job in self.jobs if job.id != job_id]
        return len(self.jobs) < initial_length

    def get_jobs_by_worker_id(self, worker_id: str) -> List[Job]:
        return [job for job in self.jobs if job.worker_id == worker_id]
