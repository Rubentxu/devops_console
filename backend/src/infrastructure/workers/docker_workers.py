import asyncio
import docker
from typing import AsyncIterable
from .abstractions import JobLauncher, JobMonitor, LogStreamer

class DockerJobLauncher(JobLauncher):
    def __init__(self, client: docker.DockerClient):
        self.client = client

    def launch_job(self, job_name: str, job_spec: dict) -> str:
        try:
            container = self.client.containers.run(
                job_spec['image'],
                name=job_name,
                command=job_spec.get('command'),
                detach=True,
                environment=job_spec.get('env'),
                volumes=job_spec.get('volumes'),
            )
            return f"Lanzando job {job_name} en Docker"
        except docker.errors.APIError as e:
            return f"Error al lanzar job {job_name}: {e}"

class DockerJobMonitor(JobMonitor):
    def __init__(self, client: docker.DockerClient):
        self.client = client

    async def get_job_status(self, job_name: str) -> str:
        try:
            container = self.client.containers.get(job_name)
            return container.status
        except docker.errors.NotFound:
            return "Not Found"

    async def monitor_job(self, job_name: str) -> AsyncIterable[str]:
        while True:
            try:
                container = self.client.containers.get(job_name)
                yield container.status
                if container.status in ['exited', 'dead']:
                    break
            except docker.errors.NotFound:
                yield "Not Found"
                break
            await asyncio.sleep(1)

class DockerLogStreamer(LogStreamer):
    def __init__(self, client: docker.DockerClient):
        self.client = client

    async def stream_logs(self, job_name: str) -> AsyncIterable[str]:
        try:
            container = self.client.containers.get(job_name)
            for line in container.logs(stream=True, follow=True):
                yield line.decode('utf-8').strip()
        except docker.errors.NotFound:
            yield f"Container {job_name} not found"
