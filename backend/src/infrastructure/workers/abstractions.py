from abc import ABC, abstractmethod
from typing import AsyncIterable

class JobLauncher(ABC):
    @abstractmethod
    def launch_job(self, job_name: str, job_spec: dict) -> str:
        pass

class JobMonitor(ABC):
    @abstractmethod
    async def get_job_status(self, job_name: str) -> str:
        pass

    @abstractmethod
    async def monitor_job(self, job_name: str) -> AsyncIterable[str]:
        pass

class LogStreamer(ABC):
    @abstractmethod
    async def stream_logs(self, job_name: str) -> AsyncIterable[str]:
        pass
