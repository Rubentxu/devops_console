from src.domain.job.job_repository import JobRepository
from src.domain.job.job import Job, JobCreate, JobUpdate
from typing import List, Optional

class JobService:
    def __init__(self, job_repository: JobRepository):
        self.job_repository = job_repository

    def create_job(self, job: JobCreate) -> Job:
        return self.job_repository.create(job)

    def get_all_jobs(self) -> List[Job]:
        return self.job_repository.get_all()

    def get_job_by_id(self, job_id: str) -> Optional[Job]:
        return self.job_repository.get_by_id(job_id)

    def update_job(self, job_id: str, job_update: JobUpdate) -> Optional[Job]:
        return self.job_repository.update(job_id, job_update)

    def delete_job(self, job_id: str) -> bool:
        return self.job_repository.delete(job_id)

    def get_jobs_by_worker_id(self, worker_id: str) -> List[Job]:
        return self.job_repository.get_jobs_by_worker_id(worker_id)