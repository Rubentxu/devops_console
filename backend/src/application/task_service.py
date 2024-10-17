from src.domain.task.task_repository import TaskRepository
import asyncio
from datetime import datetime, timedelta
from collections import deque
from src.domain.task.task import Task, TaskCreate, TaskUpdate, TaskExecuted, TaskStatus
from typing import List, Optional, AsyncIterable, AsyncIterator, Dict
from src.infrastructure.workers.factories.worker_factory import WorkerFactory, WorkerType
from contextlib import asynccontextmanager
from src.infrastructure.workers.worker_config import get_worker_config, WorkerConfig

import logging

# Configuración básica del logger
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class TaskService:
    def __init__(self, task_repository: TaskRepository, worker_configs: Dict[WorkerType, WorkerConfig]):
        self.task_repository = task_repository
        self.worker_configs = worker_configs
        self.worker_factory = WorkerFactory()
        self.task_queue = deque(maxlen=100)
        self.task_statistics = {
            "total_tasks": 0,
            "successful_tasks": 0,
            "failed_tasks": 0,
            "total_duration": timedelta()
        }
        self.paused_tasks = {}

    def get_worker(self, worker_type: WorkerType):
        logger.info(f"Getting worker for type: {worker_type}")
        logger.info(f"Worker configs: {self.worker_configs}")
        config = self.worker_configs.get(worker_type)
        logger.info(f"Worker config: {config}")
        if not config:
            logger.error(f"No configuration found for worker type: {worker_type}")
            logger.info(f"Available worker types: {self.worker_configs.keys()}")
            raise ValueError(f"No configuration found for worker type: {worker_type}")

        job_launcher = self.worker_factory.get_job_launcher(worker_type)
        job_monitor = self.worker_factory.get_job_monitor(worker_type)
        log_streamer = self.worker_factory.get_log_streamer(worker_type)

        return job_launcher, job_monitor, log_streamer, config

    def update_worker_configs(self, new_configs: Dict[WorkerType, WorkerConfig]):
        self.worker_configs.update(new_configs)

    def create_task(self, task: TaskCreate) -> Task:
        return self.task_repository.create(task)

    def get_all_tasks(self) -> List[Task]:
        return self.task_repository.get_all()

    def get_task_by_id(self, task_id: str) -> Optional[Task]:
        return self.task_repository.get_by_id(task_id)

    def update_task(self, task_id: str, task_update: TaskUpdate) -> Optional[Task]:
        return self.task_repository.update(task_id, task_update)

    def delete_task(self, task_id: str) -> bool:
        return self.task_repository.delete(task_id)

    def add_task_execution(self, task_id: str, task_executed: TaskExecuted) -> Optional[Task]:
        task = self.get_task_by_id(task_id)
        if task:
            task.tasks_executed.append(task_executed)
            return self.update_task(task_id, TaskUpdate(tasks_executed=task.tasks_executed))
        return None

    def add_task_to_queue(self, task_id: str):
        self.task_queue.append(task_id)

    def get_next_task_from_queue(self):
        return self.task_queue.popleft() if self.task_queue else None

    def update_task_statistics(self, success: bool, duration: timedelta):
        self.task_statistics["total_tasks"] += 1
        if success:
            self.task_statistics["successful_tasks"] += 1
        else:
            self.task_statistics["failed_tasks"] += 1
        self.task_statistics["total_duration"] += duration

    def get_task_statistics(self):
        avg_duration = self.task_statistics["total_duration"] / self.task_statistics["total_tasks"] if self.task_statistics["total_tasks"] > 0 else timedelta()
        success_rate = self.task_statistics["successful_tasks"] / self.task_statistics["total_tasks"] if self.task_statistics["total_tasks"] > 0 else 0
        return {
            "average_duration": avg_duration,
            "success_rate": success_rate,
            "total_tasks": self.task_statistics["total_tasks"],
            "successful_tasks": self.task_statistics["successful_tasks"],
            "failed_tasks": self.task_statistics["failed_tasks"]
        }

    async def execute_task_with_timeout(self, task_id: str, timeout_seconds: int):
        start_time = datetime.now()
        result = None
        try:
            result = await asyncio.wait_for(self.execute_task(task_id), timeout=timeout_seconds)
            return result
        except asyncio.TimeoutError:
            result = {"type": "error", "message": f"La tarea {task_id} excedió el tiempo límite de {timeout_seconds} segundos"}
            return result
        finally:
            duration = datetime.now() - start_time
            if result:
                self.update_task_statistics(success=(result.get("type") != "error"), duration=duration)

    async def pause_task(self, task_id: str):
        self.paused_tasks[task_id] = {"paused_at": datetime.now()}
        return {"type": "status", "message": f"Tarea {task_id} pausada"}

    async def resume_task(self, task_id: str):
        if task_id in self.paused_tasks:
            paused_info = self.paused_tasks.pop(task_id)
            return {"type": "status", "message": f"Tarea {task_id} reanudada"}
        return {"type": "error", "message": f"Tarea {task_id} no está pausada"}

    async def execute_task(self, task_id: str) -> str:
        task = self.get_task_by_id(task_id)
        if task:
            worker_type = self.determine_worker_type(task)
            logger.info(f"Executing task {task.name} with worker type {worker_type}")
            job_launcher, _, _, config = self.get_worker(worker_type)
            return job_launcher.launch_job(task.name, config.get_launch_config())
        return "Tarea no encontrada"

    async def get_task_status(self, task_id: str) -> str:
        task = self.get_task_by_id(task_id)
        if task:
            worker_type = self.determine_worker_type(task)
            _, job_monitor, _, _ = self.get_worker(worker_type)
            return await job_monitor.get_job_status(task.name)
        return "Tarea no encontrada"

    async def monitor_task(self, task_id: str) -> AsyncIterable[str]:
        task = self.get_task_by_id(task_id)
        if task:
            worker_type = self.determine_worker_type(task)
            _, job_monitor, _, _ = self.get_worker(worker_type)
            async for status in job_monitor.monitor_job(task.name):
                yield status
        else:
            yield "Tarea no encontrada"

    async def stream_task_logs(self, task_id: str) -> AsyncIterable[str]:
        task = self.get_task_by_id(task_id)
        if task:
            worker_type = self.determine_worker_type(task)
            _, _, log_streamer, _ = self.get_worker(worker_type)
            async for log in log_streamer.stream_logs(task.name):
                yield log
        else:
            yield "Tarea no encontrada"

    async def execute_and_monitor_task(self, task_id: str) -> AsyncIterable[dict]:
        task = self.get_task_by_id(task_id)
        if not task:
            yield {"type": "error", "message": "Tarea no encontrada"}
            return

        worker_type = self.determine_worker_type(task)
        job_launcher, job_monitor, log_streamer, config = self.get_worker(worker_type)

        try:
            launch_result = job_launcher.launch_job(task.name, config.get_launch_config())
            yield {"type": "launch", "message": launch_result}

            monitor_task = asyncio.create_task(self._monitor_task_status(task.name, job_monitor))
            log_stream = self._stream_task_logs(task.name, log_streamer)

            while not monitor_task.done():
                done, pending = await asyncio.wait(
                    [monitor_task, asyncio.create_task(anext(log_stream, None))],
                    return_when=asyncio.FIRST_COMPLETED
                )

                if monitor_task in done:
                    status = await monitor_task
                    yield {"type": "status", "status": status}
                    if status in [TaskStatus.COMPLETED, TaskStatus.FAILED]:
                        break

                for completed_task in done - {monitor_task}:
                    try:
                        log = await completed_task
                        if log is None:  # End of log stream
                            break
                        yield {"type": "log", "message": log}
                    except Exception as e:
                        yield {"type": "error", "message": f"Error en la transmisión de logs: {str(e)}"}
                        break

        except Exception as e:
            yield {"type": "error", "message": f"Error durante la ejecución de la tarea: {str(e)}"}

        finally:
            yield {"type": "close", "message": "Tarea finalizada"}

    async def _monitor_task_status(self, task_name: str, job_monitor) -> str:
        try:
            async for status in job_monitor.monitor_job(task_name):
                if status in [TaskStatus.COMPLETED, TaskStatus.FAILED]:
                    return status
            return TaskStatus.PENDING
        except GeneratorExit:
            return TaskStatus.PENDING

    async def _stream_task_logs(self, task_name: str, log_streamer) -> AsyncIterator[str]:
        try:
            async for log in log_streamer.stream_logs(task_name):
                yield log
        except GeneratorExit:
            pass

    def determine_worker_type(self, task: Task) -> WorkerType:
        # Si worker_type ya está definido, lo usamos directamente
        if task.worker_type:
            return task.worker_type

        # Si no está definido, usamos el algoritmo para determinarlo
        technology_map = {
            'kubernetes': WorkerType.KUBERNETES,
            'openshift': WorkerType.OPENSHIFT,
            'docker': WorkerType.DOCKER,
            'podman': WorkerType.PODMAN
        }

        if task.technology:
            technology = task.technology.lower()
            for key, worker_type in technology_map.items():
                if key in technology:
                    return worker_type

        if task.tags:
            tags = [tag.lower() for tag in task.tags]
            for tag in tags:
                for key, worker_type in technology_map.items():
                    if key in tag:
                        return worker_type

        if task.task_type:
            task_type = task.task_type.lower()
            if 'deployment' in task_type:
                return WorkerType.KUBERNETES
            elif 'build' in task_type:
                return WorkerType.DOCKER

        return WorkerType.KUBERNETES
