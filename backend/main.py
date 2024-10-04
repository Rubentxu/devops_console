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
import logging
from colorama import Fore, Style, init

def initialize_colorama():
    init(autoreset=True)

def load_environment_variables():
    load_dotenv()
    return int(os.getenv("API_PORT", 8080)), os.getenv('DEV_MODE', 'False')

def setup_logger(dev_mode):
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.DEBUG)
    ch = logging.StreamHandler()
    ch.setLevel(logging.DEBUG if dev_mode.lower() == 'true' else logging.INFO)
    ch.setFormatter(ColoredFormatter())
    logger.addHandler(ch)
    return logger

class ColoredFormatter(logging.Formatter):
    FORMATS = {
        logging.DEBUG: Fore.CYAN + "%(asctime)s - %(name)s - %(levelname)s - %(message)s" + Style.RESET_ALL,
        logging.INFO: Fore.GREEN + "%(asctime)s - %(name)s - %(levelname)s - %(message)s" + Style.RESET_ALL,
        logging.WARNING: Fore.YELLOW + "%(asctime)s - %(name)s - %(levelname)s - %(message)s" + Style.RESET_ALL,
        logging.ERROR: Fore.RED + "%(asctime)s - %(name)s - %(levelname)s - %(message)s" + Style.RESET_ALL,
        logging.CRITICAL: Fore.MAGENTA + "%(asctime)s - %(name)s - %(levelname)s - %(message)s" + Style.RESET_ALL
    }

    def format(self, record):
        log_fmt = self.FORMATS.get(record.levelno)
        formatter = logging.Formatter(log_fmt)
        return formatter.format(record)

def create_app():
    app = FastAPI()
    app.add_middleware(
        CORSMiddleware,
        allow_origins=['http://localhost:5173'],
        allow_credentials=True,
        allow_methods=['*'],
        allow_headers=['*'],
    )
    return app

def initialize_repositories():
    return InMemoryTaskRepository(), InMemoryTenantRepository(), InMemoryWorkspaceRepository()

def initialize_services(task_repo, tenant_repo, workspace_repo):
    return TaskService(task_repo), TenantService(tenant_repo), WorkspaceService(workspace_repo)

def load_dev_data(task_service, tenant_service, workspace_service, logger):
    with open('data.json', 'r') as f:
        logger.info('Loading dev data')
        data = json.load(f)
        logger.debug(data)

    for tenant_data in data['tenants']:
        tenant_service.create_tenant(TenantCreate(**tenant_data))

    for workspace_data in data['workspaces']:
        workspace_service.create_workspace(WorkspaceCreate(**workspace_data))

    for task_data in data['tasks']:
        task_service.create_task(TaskCreate(**task_data))

def setup_routes(app, task_service):
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

    @app.get('/')
    def read_root():
        return {'message': 'Bienvenido a la DevOps Console API'}

    class TaskExecutionRequest(BaseModel):
        form_data: Dict[str, Any]

    @app.post("/tasks/{task_id}/execute")
    async def execute_task(task_id: str, execution_request: TaskExecutionRequest):
        task = task_service.get_task_by_id(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")

        async def simulate_execution():
            await asyncio.sleep(2)
            return {"status": "completed", "message": f"Task {task_id} executed successfully"}

        execution_result = await simulate_execution()
        return execution_result

def main():
    initialize_colorama()
    port, dev_mode = load_environment_variables()
    logger = setup_logger(dev_mode)
    app = create_app()
    task_repo, tenant_repo, workspace_repo = initialize_repositories()
    task_service, tenant_service, workspace_service = initialize_services(task_repo, tenant_repo, workspace_repo)

    if dev_mode.lower() == 'true':
        load_dev_data(task_service, tenant_service, workspace_service, logger)

    setup_routes(app, task_service)

    logger.info(f"Starting server at port {port} and Dev mode: {dev_mode}")
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=port)

if __name__ == "__main__":
    main()
