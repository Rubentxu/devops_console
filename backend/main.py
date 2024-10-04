from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from src.domain.task import Task, TaskCreate, TaskUpdate, TaskExecuted
from src.domain.workspace_tenant import Tenant, TenantCreate, TenantUpdate, Workspace, WorkspaceCreate, WorkspaceUpdate
from src.infrastructure.in_memory_task_repository import InMemoryTaskRepository
from src.infrastructure.in_memory_workspace_tenant_repository import InMemoryTenantRepository, InMemoryWorkspaceRepository
from src.application.task_service import TaskService
from src.application.workspace_tenant_service import TenantService, WorkspaceService
from typing import List, Dict, Any
import os
import json
from pydantic import BaseModel
import asyncio
from dotenv import load_dotenv

load_dotenv()  # Carga las variables de entorno desde .env

# Configurar FastAPI para escuchar en el puerto especificado
app = FastAPI()
# Usa la variable de entorno para el puerto
port = int(os.getenv("API_PORT", 8000))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=port)


# Configurar CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=['*'],
    allow_credentials=True,
    allow_methods=['*'],
    allow_headers=['*'],
)

# Inicializar repositorios
task_repository = InMemoryTaskRepository()
tenant_repository = InMemoryTenantRepository()
workspace_repository = InMemoryWorkspaceRepository()

# Inicializar servicios
task_service = TaskService(task_repository)
tenant_service = TenantService(tenant_repository)
workspace_service = WorkspaceService(workspace_repository)

# Feature flag para dev mode
DEV_MODE = os.getenv('DEV_MODE', 'true').lower() == 'true'



def load_dev_data():
    with open('data.json', 'r') as f:
        data = json.load(f)

    for tenant_data in data['tenants']:
        tenant_service.create_tenant(TenantCreate(**tenant_data))

    for workspace_data in data['workspaces']:
        workspace_service.create_workspace(WorkspaceCreate(**workspace_data))

    for task_data in data['tasks']:
        task_service.create_task(TaskCreate(**task_data))




if DEV_MODE:
    load_dev_data()

# Endpoints para Task
@app.post('/tasks', response_model=Task)
def create_task(task: TaskCreate):
    return task_service.create_task(task)

@app.get('/tasks', response_model=List[Task])
def read_tasks():
    return task_service.get_all_tasks()

@app.get('/tasks/{task_id}', response_model=Task)
def read_task(task_id: str):
    task = task_service.get_task_by_id(task_id)
    if task is None:
        raise HTTPException(status_code=404, detail='Task not found')
    return task

@app.put('/tasks/{task_id}', response_model=Task)
def update_task(task_id: str, task_update: TaskUpdate):
    task = task_service.update_task(task_id, task_update)
    if task is None:
        raise HTTPException(status_code=404, detail='Task not found')
    return task

@app.delete('/tasks/{task_id}', response_model=Task)
def delete_task(task_id: str):
    task = task_service.get_task_by_id(task_id)
    if task is None:
        raise HTTPException(status_code=404, detail='Task not found')
    if task_service.delete_task(task_id):
        return task
    raise HTTPException(status_code=500, detail='Failed to delete task')

@app.post('/tasks/{task_id}/executions', response_model=Task)
def add_task_execution(task_id: str, task_executed: TaskExecuted):
    task = task_service.add_task_execution(task_id, task_executed)
    if task is None:
        raise HTTPException(status_code=404, detail='Task not found')
    return task

# ... (mantén los otros endpoints existentes aquí)

@app.get('/')
def read_root():
    return {'message': 'Bienvenido a la DevOps Console API'}


class TaskExecutionRequest(BaseModel):
    form_data: Dict[str, Any]

@app.post("/tasks/{task_id}/execute")
async def execute_task(task_id: str, execution_request: TaskExecutionRequest):
    task = next((task for task in tasks if task.id == task_id), None)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    # Aquí simularemos la ejecución de la tarea
    # En una implementación real, esto podría iniciar un proceso en segundo plano
    async def simulate_execution():
        await asyncio.sleep(2)  # Simulamos algún trabajo
        return {"status": "completed", "message": f"Task {task_id} executed successfully"}

    execution_result = await simulate_execution()
    return execution_result
