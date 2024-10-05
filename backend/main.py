from fastapi import FastAPI
from src.infrastructure.config.app_config import create_app
from src.infrastructure.utils.environment import load_environment_variables
from src.infrastructure.utils.logger import setup_logger
from src.infrastructure.repositories.in_memory_task_repository import InMemoryTaskRepository
from src.infrastructure.repositories.in_memory_workspace_tenant_repository import InMemoryTenantRepository, InMemoryWorkspaceRepository
from src.application.task_service import TaskService, TaskCreate, TaskUpdate, TaskExecuted
from src.application.workspace_service import WorkspaceService, WorkspaceCreate
from src.application.tenant_service import TenantService, TenantCreate
from src.infrastructure.routes.task_routes import create_task_router
from src.infrastructure.routes.workspace_routes import create_workspace_router
from colorama import init as colorama_init
import json
import uvicorn

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

def main():
    colorama_init(autoreset=True)
    port, dev_mode = load_environment_variables()
    logger = setup_logger(dev_mode)
    app = create_app()

    task_repo, tenant_repo, workspace_repo = initialize_repositories()
    task_service, tenant_service, workspace_service = initialize_services(task_repo, tenant_repo, workspace_repo)

    if dev_mode.lower() == 'true':
        load_dev_data(task_service, tenant_service, workspace_service, logger)

    task_router = create_task_router(task_service)
    workspace_router = create_workspace_router(workspace_service)

    app.include_router(task_router)
    app.include_router(workspace_router)

    @app.get('/')
    def read_root():
        return {'message': 'Bienvenido a la DevOps Console API'}

    logger.info(f"Starting server at port {port} and Dev mode: {dev_mode}")
    uvicorn.run(app, host="0.0.0.0", port=port)

if __name__ == "__main__":
    main()
