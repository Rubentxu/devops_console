from fastapi import APIRouter, HTTPException
from src.domain.worker.worker import Worker, WorkerCreate, WorkerUpdate
from src.application.worker_service import WorkerService
from typing import List

def create_worker_router(worker_service: WorkerService):
    router = APIRouter()

    @router.post('/workers', response_model=Worker)
    def create_worker(worker: WorkerCreate):
        return worker_service.create_worker(worker)

    @router.get('/workers', response_model=List[Worker])
    def read_workers():
        return worker_service.get_all_workers()

    @router.get('/workers/{worker_id}', response_model=Worker)
    def read_worker(worker_id: str):
        worker = worker_service.get_worker_by_id(worker_id)
        if worker is None:
            raise HTTPException(status_code=404, detail='Worker not found')
        return worker

    @router.put('/workers/{worker_id}', response_model=Worker)
    def update_worker(worker_id: str, worker_update: WorkerUpdate):
        worker = worker_service.update_worker(worker_id, worker_update)
        if worker is None:
            raise HTTPException(status_code=404, detail='Worker not found')
        return worker

    @router.delete('/workers/{worker_id}', response_model=Worker)
    def delete_worker(worker_id: str):
        worker = worker_service.get_worker_by_id(worker_id)
        if worker is None:
            raise HTTPException(status_code=404, detail='Worker not found')
        if worker_service.delete_worker(worker_id):
            return worker
        raise HTTPException(status_code=500, detail='Failed to delete worker')

    return router
