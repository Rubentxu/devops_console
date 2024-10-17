# Importaciones de configuración y utilidades
from src.infrastructure.config.app_config import create_app
from src.infrastructure.utils.environment import load_environment_variables
from src.infrastructure.utils.logger import setup_logger

# Importaciones de repositorios
from src.infrastructure.repositories.in_memory_task_repository import InMemoryTaskRepository
from src.infrastructure.repositories.in_memory_tenant_repository import InMemoryTenantRepository
from src.infrastructure.repositories.in_memory_workspace_repository import InMemoryWorkspaceRepository
from src.infrastructure.repositories.in_memory_worker_repository import InMemoryWorkerRepository
from src.infrastructure.repositories.in_memory_job_repository import InMemoryJobRepository

# Importaciones de servicios
from src.application.task_service import TaskService
from src.application.worker_service import WorkerService
from src.application.job_service import JobService
from src.application.workspace_service import WorkspaceService
from src.application.tenant_service import TenantService

# Importaciones de rutas
from src.infrastructure.routes.task_routes import create_task_router
from src.infrastructure.routes.workspace_routes import create_workspace_router
from src.infrastructure.routes.tenant_routes import create_tenant_router
from src.infrastructure.routes.worker_routes import create_worker_router
from src.infrastructure.routes.job_routes import create_job_router

# Importaciones de dominio
from src.infrastructure.workers.factories.worker_factory import WorkerType
from src.domain.tenant.tenant import Tenant, TenantCreate, TenantUpdate
from src.domain.workspace.workspace import Workspace, WorkspaceCreate, WorkspaceUpdate
from src.domain.task.task import Task, TaskCreate, TaskUpdate
from src.domain.worker.worker import WorkerCreate
from src.infrastructure.workers.worker_config import get_worker_config, WorkerConfig

# Bibliotecas externas
from colorama import init as colorama_init
import json
import uvicorn
from fastapi import FastAPI
from typing import Dict

def initialize_repositories():
    """
    Inicializa y retorna instancias de todos los repositorios en memoria.

    Returns:
        tuple: Instancias de repositorios (Task, Tenant, Workspace, Worker, Job)
    """
    return (
        InMemoryTaskRepository(),
        InMemoryTenantRepository(),
        InMemoryWorkspaceRepository(),
        InMemoryWorkerRepository(),
        InMemoryJobRepository()
    )

def load_worker_configs(config_file: str) -> Dict[WorkerType, WorkerConfig]:
    """
    Carga las configuraciones de los workers desde un archivo JSON.

    Args:
        config_file (str): Ruta al archivo de configuración JSON.

    Returns:
        Dict[WorkerType, WorkerConfig]: Diccionario de configuraciones de worker.
    """
    with open(config_file, 'r') as f:
        data = json.load(f)

    worker_configs = {}
    for worker_data in data['workers']:
        config = get_worker_config(worker_data)
        worker_type_str = worker_data['type'].upper()
        try:
            worker_type = WorkerType[worker_type_str]
            worker_configs[worker_type] = config
        except KeyError:
            print(f"Warning: Unknown worker type '{worker_data['type']}'. Skipping this worker configuration.")

    return worker_configs

def initialize_services(task_repo, tenant_repo, workspace_repo, worker_repo, job_repo, worker_configs):
    """
    Inicializa y retorna instancias de todos los servicios.

    Args:
        task_repo, tenant_repo, workspace_repo, worker_repo, job_repo: Instancias de repositorios.
        worker_configs (Dict[WorkerType, WorkerConfig]): Configuraciones de los workers.

    Returns:
        tuple: Instancias de servicios (Task, Tenant, Workspace, Worker, Job)
    """
    return (
        TaskService(task_repo, worker_configs),
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

    # Crear tenants
    for tenant_data in data['tenants']:
        tenant_service.create_tenant(TenantCreate(**tenant_data))

    # Crear workspaces
    for workspace_data in data['workspaces']:
        workspace_service.create_workspace(WorkspaceCreate(**workspace_data))

    # Crear tasks
    for task_data in data['tasks']:
        # Convertir worker_type de string a WorkerType si existe
        if 'worker_type' in task_data:
            worker_type_str = task_data['worker_type'].upper()
            try:
                task_data['worker_type'] = WorkerType[worker_type_str]
            except KeyError:
                logger.warning(f"Unknown worker type '{task_data['worker_type']}' for task. Setting to None.")
                task_data['worker_type'] = None
        task_service.create_task(TaskCreate(**task_data))

    # Crear workers
    for worker_data in data['workers']:
        worker_type_str = worker_data['type'].upper()
        try:
            worker_type = WorkerType[worker_type_str]
            worker_create = WorkerCreate(
                name=worker_data['name'],
                type=worker_type,
                config=worker_data['config']
            )
            worker_service.create_worker(worker_create)
        except KeyError:
            logger.warning(f"Unknown worker type '{worker_data['type']}'. Skipping this worker.")

    logger.info('Dev data loaded successfully')

def main():
    """
    Función principal que configura y ejecuta la aplicación.
    """
    # Inicializar colorama para salida de color en la consola
    colorama_init(autoreset=True)

    # Cargar variables de entorno
    port, dev_mode = load_environment_variables()

    # Configurar el logger
    logger = setup_logger(dev_mode)

    # Crear la aplicación FastAPI
    app = create_app()

    # Cargar configuraciones de workers
    worker_configs = load_worker_configs('worker_config.json')

    # Inicializar repositorios y servicios
    repos = initialize_repositories()
    services = initialize_services(*repos, worker_configs)
    task_service, tenant_service, workspace_service, worker_service, job_service = services

    # Cargar datos de desarrollo si estamos en modo dev
    if dev_mode.lower() == 'true':
        load_dev_data(task_service, tenant_service, workspace_service, worker_service, job_service, logger)

    # Crear routers
    task_router = create_task_router(task_service)
    workspace_router = create_workspace_router(workspace_service)
    tenant_router = create_tenant_router(tenant_service)
    worker_router = create_worker_router(worker_service)
    job_router = create_job_router(job_service)

    # Incluir routers en la aplicación
    app.include_router(task_router)
    app.include_router(workspace_router)
    app.include_router(tenant_router)
    app.include_router(worker_router)
    app.include_router(job_router)

    # Definir ruta raíz
    @app.get('/')
    def read_root():
        return {'message': 'Bienvenido a la DevOps Console API'}

    # Iniciar el servidor
    logger.info(f"Starting server at port {port} and Dev mode: {dev_mode}")
    uvicorn.run(app, host="0.0.0.0", port=port)

if __name__ == "__main__":
    main()
