import asyncio
from typing import AsyncIterable, Any
import openshift_client as oc
from .abstractions import JobLauncher, JobMonitor, LogStreamer
from .worker_config import OpenShiftWorkerConfig

class OpenShiftJobLauncher(JobLauncher):
    def __init__(self, config: OpenShiftWorkerConfig):
        self.config = config

    def launch_job(self, job_name: str, job_spec: dict) -> str:
        with oc.project(self.config.project):
            job_template = self.config.job_template.copy()
            job_template['metadata']['name'] = job_name
            job_template['spec'] = job_spec
            try:
                oc.selector('jobs').create(job_template)
                return f"Lanzando job {job_name} en OpenShift"
            except Exception as e:
                return f"Error al lanzar job {job_name}: {e}"

class OpenShiftJobMonitor(JobMonitor):
    async def get_job_status(self, job_name: str) -> str:
        with oc.project(self.config.project):
            try:
                job = oc.selector(f'job/{job_name}').object()
                if job.model.status.active:
                    return "Running"
                elif job.model.status.succeeded:
                    return "Succeeded"
                elif job.model.status.failed:
                    return "Failed"
                else:
                    return "Unknown"
            except Exception as e:
                return f"Error: {e}"

    async def monitor_job(self, job_name: str) -> AsyncIterable[str]:
        with oc.project(self.config.project):
            while True:
                status = await self.get_job_status(job_name)
                yield status
                if status in ["Succeeded", "Failed"]:
                    break
                await asyncio.sleep(1)

class OpenShiftLogStreamer(LogStreamer):
    async def stream_logs(self, job_name: str) -> AsyncIterable[str]:
        with oc.project(self.config.project):
            pod = await self._get_pod_for_job(job_name)
            async for line in self._stream_pod_logs(pod):
                yield line

    async def _get_pod_for_job(self, job_name: str):
        while True:
            pods = oc.selector(f'job-name={job_name}')
            if pods.count_existing() > 0:
                return pods.objects()[0]
            await asyncio.sleep(1)

    async def _stream_pod_logs(self, pod) -> AsyncIterable[str]:
        for line in pod.logs(follow=True):
            yield line.strip()