from src.infrastructure.config.app_config import create_app
from src.infrastructure.utils.environment import load_environment_variables
from src.infrastructure.utils.logger import setup_logger
# Repositories
from src.infrastructure.repositories.in_memory_task_repository import InMemoryTaskRepository
from src.infrastructure.repositories.in_memory_tenant_repository import InMemoryTenantRepository
from src.infrastructure.repositories.in_memory_workspace_repository import InMemoryWorkspaceRepository
from src.infrastructure.repositories.in_memory_worker_repository import InMemoryWorkerRepository
from src.infrastructure.repositories.in_memory_job_repository import InMemoryJobRepository
# Services
from src.application.task_service import TaskService
from src.application.worker_service import WorkerService
from src.application.job_service import JobService
from src.application.workspace_service import WorkspaceService
from src.application.tenant_service import TenantService
# Routes
from src.infrastructure.routes.task_routes import create_task_router
from src.infrastructure.routes.workspace_routes import create_workspace_router
from src.infrastructure.routes.tenant_routes import create_tenant_router
from src.infrastructure.routes.worker_routes import create_worker_router
from src.infrastructure.routes.job_routes import create_job_router
from src.infrastructure.workers.factories.worker_factory import WorkerType
# Domain
from src.domain.tenant.tenant import Tenant, TenantCreate, TenantUpdate
from src.domain.workspace.workspace import Workspace, WorkspaceCreate, WorkspaceUpdate
from src.domain.task.task import Task, TaskCreate, TaskUpdate
from src.domain.worker.worker import WorkerCreate
from src.domain.worker.worker_config import get_worker_config

from src.infrastructure.config.app_config import create_app

# Libraries
from colorama import init as colorama_init
import json
import uvicorn
from fastapi import FastAPI




def initialize_repositories():
    return (
        InMemoryTaskRepository(),
        InMemoryTenantRepository(),
        InMemoryWorkspaceRepository(),
        InMemoryWorkerRepository(),
        InMemoryJobRepository()
    )

def initialize_services(task_repo, tenant_repo, workspace_repo, worker_repo, job_repo):
    return (
        TaskService(task_repo),  # Inicializamos TaskService sin configuraciones de worker
        TenantService(tenant_repo),
        WorkspaceService(workspace_repo),
        WorkerService(worker_repo),
        JobService(job_repo)
    )

def load_dev_data(task_service, tenant_service, workspace_service, worker_service, job_service, logger):
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

    worker_configs = {}
    for worker_data in data['workers']:
        config = get_worker_config(worker_data)
        worker_type = WorkerType(worker_data['type'])
        worker_configs[worker_type] = config

        worker_create = WorkerCreate(
            name=worker_data['name'],
            type=worker_data['type'],
            config=config.dict()
        )
        worker_service.create_worker(worker_create)

    # Actualizar las configuraciones de worker en el TaskService
    task_service.update_worker_configs(worker_configs)

    logger.info('Dev data loaded successfully')

def main():
    colorama_init(autoreset=True)
    port, dev_mode = load_environment_variables()
    logger = setup_logger(dev_mode)
    app = create_app()

    task_repo, tenant_repo, workspace_repo, worker_repo, job_repo = initialize_repositories()
    task_service, tenant_service, workspace_service, worker_service, job_service = initialize_services(
        task_repo, tenant_repo, workspace_repo, worker_repo, job_repo
    )

    if dev_mode.lower() == 'true':
        load_dev_data(task_service, tenant_service, workspace_service, worker_service, job_service, logger)

    task_router = create_task_router(task_service)
    workspace_router = create_workspace_router(workspace_service)
    tenant_router = create_tenant_router(tenant_service)
    worker_router = create_worker_router(worker_service)
    job_router = create_job_router(job_service)

    app.include_router(task_router)
    app.include_router(workspace_router)
    app.include_router(tenant_router)
    app.include_router(worker_router)
    app.include_router(job_router)

    @app.get('/')
    def read_root():
        return {'message': 'Bienvenido a la DevOps Console API'}

    logger.info(f"Starting server at port {port} and Dev mode: {dev_mode}")
    uvicorn.run(app, host="0.0.0.0", port=port)

if __name__ == "__main__":
    main()
