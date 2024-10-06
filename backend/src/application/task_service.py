from src.domain.task.task_repository import TaskRepository
import asyncio
from datetime import datetime, timedelta
from collections import deque
from src.domain.task.task import Task, TaskCreate, TaskUpdate, TaskExecuted
from typing import List, Optional, AsyncIterable, AsyncIterator, Tuple, Dict
from src.infrastructure.workers.factories.worker_factory import WorkerFactory, WorkerType
from contextlib import asynccontextmanager
from src.infrastructure.workers.worker_config import get_worker_config, WorkerConfig

class TaskQueue:
    def __init__(self, max_size=100):
        self.queue = deque(maxlen=max_size)

    def add_task(self, task_id):
        self.queue.append(task_id)

    def get_next_task(self):
        return self.queue.popleft() if self.queue else None

class TaskStatistics:
    def __init__(self):
        self.total_tasks = 0
        self.successful_tasks = 0
        self.failed_tasks = 0
        self.total_duration = timedelta()

    def add_task_result(self, success, duration):
        self.total_tasks += 1
        if success:
            self.successful_tasks += 1
        else:
            self.failed_tasks += 1
        self.total_duration += duration

    def get_average_duration(self):
        return self.total_duration / self.total_tasks if self.total_tasks > 0 else timedelta()

    def get_success_rate(self):
        return self.successful_tasks / self.total_tasks if self.total_tasks > 0 else 0

class TaskManager:
    def __init__(self):
        self.task_queue = TaskQueue()
        self.task_statistics = TaskStatistics()
        self.paused_tasks = {}

    async def execute_task_with_timeout(self, task_id, timeout_seconds):
        start_time = datetime.now()
        try:
            result = await asyncio.wait_for(self.execute_task(task_id), timeout=timeout_seconds)
            return result
        except asyncio.TimeoutError:
            return {"type": "error", "message": f"La tarea {task_id} excedió el tiempo límite de {timeout_seconds} segundos"}
        finally:
            duration = datetime.now() - start_time
            self.task_statistics.add_task_result(success=(result.get("type") != "error"), duration=duration)

    async def pause_task(self, task_id):
        # Implementación depende del sistema de workers
        self.paused_tasks[task_id] = {"paused_at": datetime.now()}
        return {"type": "status", "message": f"Tarea {task_id} pausada"}

    async def resume_task(self, task_id):
        if task_id in self.paused_tasks:
            # Implementación depende del sistema de workers
            paused_info = self.paused_tasks.pop(task_id)
            return {"type": "status", "message": f"Tarea {task_id} reanudada"}
        return {"type": "error", "message": f"Tarea {task_id} no está pausada"}

    def get_task_statistics(self):
        return {
            "average_duration": self.task_statistics.get_average_duration(),
            "success_rate": self.task_statistics.get_success_rate(),
            "total_tasks": self.task_statistics.total_tasks,
            "successful_tasks": self.task_statistics.successful_tasks,
            "failed_tasks": self.task_statistics.failed_tasks
        }


class TaskService:
    def __init__(self, task_repository: TaskRepository, worker_configs: Dict[WorkerType, WorkerConfig] = None):
        self.task_repository = task_repository
        self.worker_configs = worker_configs or {}
        self.worker_factory = WorkerFactory()

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

    async def execute_task(self, task_id: str) -> str:
        task = self.get_task_by_id(task_id)
        if task:
            return self.job_launcher.launch_job(task.name)
        return "Tarea no encontrada"

    async def get_task_status(self, task_id: str) -> str:
        task = self.get_task_by_id(task_id)
        if task:
            return await self.job_monitor.get_job_status(task.name)
        return "Tarea no encontrada"

    async def monitor_task(self, task_id: str) -> AsyncIterable[str]:
        task = self.get_task_by_id(task_id)
        if task:
            async for status in await self.job_monitor.monitor_job(task.name):
                yield status
        else:
            yield "Tarea no encontrada"

    async def stream_task_logs(self, task_id: str) -> AsyncIterable[str]:
        task = self.get_task_by_id(task_id)
        if task:
            async for log in await self.log_streamer.stream_logs(task.name):
                yield log
        else:
            yield "Tarea no encontrada"

    @asynccontextmanager
    async def managed_stream(self, stream):
        try:
            yield stream
        finally:
            await stream.aclose()

    async def execute_and_monitor_task(self, task_id: str) -> AsyncIterable[dict]:
        task = self.get_task_by_id(task_id)
        if not task:
            yield {"type": "error", "message": "Tarea no encontrada"}
            return

        try:
            launch_result = self.job_launcher.launch_job(task.name)
            yield {"type": "launch", "message": launch_result}

            monitor_task = asyncio.create_task(self._monitor_task_status(task.name))
            log_stream = self._stream_task_logs(task.name)

            while not monitor_task.done():
                done, pending = await asyncio.wait(
                    [monitor_task, asyncio.create_task(anext(log_stream, None))],
                    return_when=asyncio.FIRST_COMPLETED
                )

                if monitor_task in done:
                    status = await monitor_task
                    yield {"type": "status", "status": status}
                    if status in ["Succeeded", "Failed", "Error"]:
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

    async def _monitor_task_status(self, task_name: str) -> str:
        try:
            async for status in self.job_monitor.monitor_job(task_name):
                if status in ["Succeeded", "Failed", "Error"]:
                    return status
            return "Unknown"
        except GeneratorExit:
            return "Cancelled"

    async def _stream_task_logs(self, task_name: str) -> AsyncIterator[str]:
        try:
            async for log in self.log_streamer.stream_logs(task_name):
                yield log
        except GeneratorExit:
            pass
