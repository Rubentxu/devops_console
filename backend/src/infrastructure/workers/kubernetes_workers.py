import asyncio
from typing import AsyncIterable
from kubernetes import client, watch
from .abstractions import JobLauncher, JobMonitor, LogStreamer
from .worker_config import KubernetesWorkerConfig

class KubernetesJobLauncher(JobLauncher):
    def __init__(self, batch_v1: client.BatchV1Api, config: KubernetesWorkerConfig):
        self.batch_v1 = batch_v1
        self.config = config

    def launch_job(self, job_name: str, job_spec: dict) -> str:
        namespace = self.config.namespace
        job_template = self.config.job_template.copy()
        job_template['metadata']['name'] = job_name
        job_template['spec'] = job_spec
        try:
            self.batch_v1.create_namespaced_job(namespace, job_template)
            return f"Lanzando job {job_name} en Kubernetes"
        except client.ApiException as e:
            return f"Error al lanzar job {job_name}: {e}"

class KubernetesJobMonitor(JobMonitor):
    def __init__(self, batch_v1: client.BatchV1Api):
        self.batch_v1 = batch_v1

    async def get_job_status(self, job_name: str) -> str:
        namespace = "default"
        try:
            job = self.batch_v1.read_namespaced_job_status(job_name, namespace)
            if job.status.active:
                return "Running"
            elif job.status.succeeded:
                return "Succeeded"
            elif job.status.failed:
                return "Failed"
            else:
                return "Unknown"
        except client.ApiException as e:
            return f"Error: {e}"

    async def monitor_job(self, job_name: str) -> AsyncIterable[str]:
        namespace = "default"
        w = watch.Watch()
        for event in w.stream(self.batch_v1.list_namespaced_job, namespace, field_selector=f"metadata.name={job_name}"):
            job = event['object']
            if job.status.active:
                yield "Running"
            elif job.status.succeeded:
                yield "Succeeded"
                break
            elif job.status.failed:
                yield "Failed"
                break
            await asyncio.sleep(1)

class KubernetesLogStreamer(LogStreamer):
    def __init__(self, core_v1: client.CoreV1Api):
        self.core_v1 = core_v1

    async def stream_logs(self, job_name: str) -> AsyncIterable[str]:
        namespace = "default"
        pod = None
        while not pod:
            pods = self.core_v1.list_namespaced_pod(namespace, label_selector=f"job-name={job_name}")
            if pods.items:
                pod = pods.items[0]
            else:
                await asyncio.sleep(1)

        logs = self.core_v1.read_namespaced_pod_log(pod.metadata.name, namespace, follow=True, _preload_content=False)
        for line in logs:
            yield line.decode('utf-8').strip()
