from enum import Enum
from src.infrastructure.workers.kubernetes_workers import KubernetesJobLauncher, KubernetesJobMonitor, KubernetesLogStreamer
from src.infrastructure.workers.docker_workers import DockerJobLauncher, DockerJobMonitor, DockerLogStreamer
# Importa también las clases de Podman si las tienes

class WorkerType(Enum):
    KUBERNETES = "kubernetes"
    DOCKER = "docker"
    PODMAN = "podman"

class WorkerFactory:
    @staticmethod
    def get_job_launcher(worker_type: WorkerType):
        if worker_type == WorkerType.KUBERNETES:
            return KubernetesJobLauncher()
        elif worker_type == WorkerType.DOCKER:
            return DockerJobLauncher()
        elif worker_type == WorkerType.PODMAN:
            return PodmanJobLauncher()
        else:
            raise ValueError("Tipo de worker no soportado")

    @staticmethod
    def get_job_monitor(worker_type: WorkerType):
        if worker_type == WorkerType.KUBERNETES:
            return KubernetesJobMonitor()
        elif worker_type == WorkerType.DOCKER:
            return DockerJobMonitor()
        elif worker_type == WorkerType.PODMAN:
            return PodmanJobMonitor()
        else:
            raise ValueError("Tipo de worker no soportado")

    @staticmethod
    def get_log_streamer(worker_type: WorkerType):
        if worker_type == WorkerType.KUBERNETES:
            return KubernetesLogStreamer()
        elif worker_type == WorkerType.DOCKER:
            return DockerLogStreamer()
        elif worker_type == WorkerType.PODMAN:
            return PodmanLogStreamer()
        else:
            raise ValueError("Tipo de worker no soportado")
