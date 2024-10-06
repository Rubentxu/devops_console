from fastapi import APIRouter, HTTPException, WebSocket, WebSocketDisconnect
from src.domain.task.task import Task, TaskCreate, TaskUpdate, TaskExecuted
from src.application.task_service import TaskService
from typing import List, Dict, Any
import asyncio
import random
from pydantic import BaseModel


class TaskExecutionRequest(BaseModel):
    form_data: Dict[str, Any]

active_connections: Dict[str, WebSocket] = {}

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

    @router.post("/tasks/{task_id}/execute")
    async def execute_task(task_id: str):
        result = await task_service.execute_task(task_id)
        if result == "Tarea no encontrada":
            raise HTTPException(status_code=404, detail="Task not found")
        return {"result": result, "websocket_url": f"/ws/task/{task_id}"}

    @router.websocket("/ws/task/{task_id}")
    async def websocket_endpoint(websocket: WebSocket, task_id: str):
        await websocket.accept()
        try:
            async for message in task_service.execute_and_monitor_task(task_id):
                if await websocket.send_json(message):
                    break
        except WebSocketDisconnect:
            pass
        except Exception as e:
            logger.error(f"Error in WebSocket connection: {str(e)}")
        finally:
            try:
                await websocket.close()
            except RuntimeError:
                # La conexión ya está cerrada, ignoramos este error
                pass


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
