from fastapi import APIRouter, HTTPException
from src.domain.job.job import Job, JobCreate, JobUpdate
from src.application.job_service import JobService
from typing import List

def create_job_router(job_service: JobService):
    router = APIRouter()

    @router.post('/jobs', response_model=Job)
    def create_job(job: JobCreate):
        return job_service.create_job(job)

    @router.get('/jobs', response_model=List[Job])
    def read_jobs():
        return job_service.get_all_jobs()

    @router.get('/jobs/{job_id}', response_model=Job)
    def read_job(job_id: str):
        job = job_service.get_job_by_id(job_id)
        if job is None:
            raise HTTPException(status_code=404, detail='Job not found')
        return job

    @router.put('/jobs/{job_id}', response_model=Job)
    def update_job(job_id: str, job_update: JobUpdate):
        job = job_service.update_job(job_id, job_update)
        if job is None:
            raise HTTPException(status_code=404, detail='Job not found')
        return job

    @router.delete('/jobs/{job_id}', response_model=Job)
    def delete_job(job_id: str):
        job = job_service.get_job_by_id(job_id)
        if job is None:
            raise HTTPException(status_code=404, detail='Job not found')
        if job_service.delete_job(job_id):
            return job
        raise HTTPException(status_code=500, detail='Failed to delete job')

    @router.get('/workers/{worker_id}/jobs', response_model=List[Job])
    def read_jobs_by_worker(worker_id: str):
        return job_service.get_jobs_by_worker_id(worker_id)

    return router