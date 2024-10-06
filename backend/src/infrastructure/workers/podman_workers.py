import asyncio
from typing import AsyncIterable
from podman import PodmanClient
from .abstractions import JobLauncher, JobMonitor, LogStreamer

class PodmanJobLauncher(JobLauncher):
    def __init__(self, client: PodmanClient):
        self.client = client

    def launch_job(self, job_name: str, job_spec: dict) -> str:
        try:
            container = self.client.containers.create(
                job_spec['image'],
                name=job_name,
                command=job_spec.get('command'),
                detach=True,
                environment=job_spec.get('env'),
                mounts=job_spec.get('volumes'),
            )
            container.start()
            return f"Lanzando job {job_name} en Podman"
        except Exception as e:
            return f"Error al lanzar job {job_name}: {e}"

class PodmanJobMonitor(JobMonitor):
    def __init__(self, client: PodmanClient):
        self.client = client

    async def get_job_status(self, job_name: str) -> str:
        try:
            container = self.client.containers.get(job_name)
            return container.status
        except Exception:
            return "Not Found"

    async def monitor_job(self, job_name: str) -> AsyncIterable[str]:
        while True:
            try:
                container = self.client.containers.get(job_name)
                yield container.status
                if container.status in ['Exited', 'Stopped']:
                    break
            except Exception:
                yield "Not Found"
                break
            await asyncio.sleep(1)

class PodmanLogStreamer(LogStreamer):
    def __init__(self, client: PodmanClient):
        self.client = client

    async def stream_logs(self, job_name: str) -> AsyncIterable[str]:
        try:
            container = self.client.containers.get(job_name)
            for line in container.logs(stream=True, follow=True):
                yield line.decode('utf-8').strip()
        except Exception:
            yield f"Container {job_name} not found"
