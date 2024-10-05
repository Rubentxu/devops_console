from fastapi import APIRouter, HTTPException, WebSocket, WebSocketDisconnect
from src.domain.task import Task, TaskCreate, TaskUpdate, TaskExecuted
from src.application.task_service import TaskService
from typing import List, Dict, Any
import asyncio
import random
from pydantic import BaseModel

def create_task_router(task_service: TaskService):
    router = APIRouter()

    @router.post('/tasks', response_model=Task)
    def create_task(task: TaskCreate):
        return task_service.create_task(task)

    @router.get('/tasks', response_model=List[Task])
    def read_tasks():
        return task_service.get_all_tasks()

    @router.get('/tasks/{task_id}', response_model=Task)
    def read_task(task_id: str):
        task = task_service.get_task_by_id(task_id)
        if task is None:
            raise HTTPException(status_code=404, detail='Task not found')
        return task

    @router.put('/tasks/{task_id}', response_model=Task)
    def update_task(task_id: str, task_update: TaskUpdate):
        task = task_service.update_task(task_id, task_update)
        if task is None:
            raise HTTPException(status_code=404, detail='Task not found')
        return task

    @router.delete('/tasks/{task_id}', response_model=Task)
    def delete_task(task_id: str):
        task = task_service.get_task_by_id(task_id)
        if task is None:
            raise HTTPException(status_code=404, detail='Task not found')
        if task_service.delete_task(task_id):
            return task
        raise HTTPException(status_code=500, detail='Failed to delete task')

    @router.post('/tasks/{task_id}/executions', response_model=Task)
    def add_task_execution(task_id: str, task_executed: TaskExecuted):
        task = task_service.add_task_execution(task_id, task_executed)
        if task is None:
            raise HTTPException(status_code=404, detail='Task not found')
        return task

    class TaskExecutionRequest(BaseModel):
        form_data: Dict[str, Any]

    active_connections: Dict[str, WebSocket] = {}

    @router.post("/tasks/{task_id}/execute")
    async def execute_task(task_id: str, execution_request: TaskExecutionRequest):
        task = task_service.get_task_by_id(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")

        async def simulate_execution():
            await asyncio.sleep(2)
            return {"status": "started", "task_id": task_id, "websocket_url": f"/ws/task/{task_id}"}

        execution_result = await simulate_execution()
        return execution_result

    @router.websocket("/ws/task/{task_id}")
    async def websocket_endpoint(websocket: WebSocket, task_id: str):
        await websocket.accept()
        active_connections[task_id] = websocket
        try:
            await simulate_task_execution(task_id, websocket)
        except WebSocketDisconnect:
            del active_connections[task_id]
        finally:
            if task_id in active_connections:
                del active_connections[task_id]

    async def simulate_task_execution(task_id: str, websocket: WebSocket):
        steps = ["Initializing", "Processing", "Finalizing", "Evaluation", "Calculating", "Validating", "Finishing"]
        try:
            for step in steps:
                await websocket.send_json({"type": "log", "message": f"[INFO] {step} task {task_id}"})
                await asyncio.sleep(0.3)
                for _ in range(30):
                    log_type = random.choice(["INFO", "WARNING", "ERROR", "DEBUG"])
                    message = f"[{log_type}] {random.choice(['Process A', 'Process B', 'Process C', 'Process D', 'Process E'])} - {random.randint(1000, 9999)}"
                    await websocket.send_json({"type": "log", "message": message})
                    await asyncio.sleep(0.2)

            if random.random() < 0.9:
                await websocket.send_json({"type": "status", "status": "completed"})
            else:
                await websocket.send_json({"type": "error", "message": "Task failed due to an unexpected error"})
        finally:
            await websocket.close()

    return router
