from typing import Dict, Any
from kubernetes import client as k8s_client
from docker import DockerClient
# Comentamos la importación de Podman ya que parece que no está instalado
# from podman import PodmanClient
import openshift_client as oc
from ..kubernetes_workers import KubernetesJobLauncher, KubernetesJobMonitor, KubernetesLogStreamer
from ..docker_workers import DockerJobLauncher, DockerJobMonitor, DockerLogStreamer
from ..podman_workers import PodmanJobLauncher, PodmanJobMonitor, PodmanLogStreamer
from ..openshift_workers import OpenShiftJobLauncher, OpenShiftJobMonitor, OpenShiftLogStreamer
from ..worker_config import OpenShiftWorkerConfig, KubernetesWorkerConfig, DockerWorkerConfig, PodmanWorkerConfig
from ..worker_types import WorkerType
from src.infrastructure.utils.logger import setup_logger


# Configurar el logger
logger = setup_logger('true')

class WorkerFactory:
    def __init__(self):
        self.docker_client = None
        self.podman_client = None
        self.kubernetes_client = None
        self.openshift_client = None
        self.configs = {}

    def initialize_kubernetes(self):
        k8s_client.config.load_kube_config()
        self.kubernetes_client = {
            'batch_v1': k8s_client.BatchV1Api(),
            'core_v1': k8s_client.CoreV1Api()
        }

    def initialize_docker(self):
        self.docker_client = DockerClient.from_env()

    def initialize_podman(self):
        # Comentamos esto ya que Podman no está instalado
        # self.podman_client = PodmanClient()
        pass

    def initialize_openshift(self):
        # oc.config.load_kube_config()
        self.openshift_client = oc

    def set_config(self, worker_type: WorkerType, config: Dict[str, Any]):
        if worker_type == WorkerType.KUBERNETES:
            self.configs[worker_type] = KubernetesWorkerConfig
        elif worker_type == WorkerType.DOCKER:
            self.configs[worker_type] = DockerWorkerConfig
        elif worker_type == WorkerType.PODMAN:
            self.configs[worker_type] = PodmanWorkerConfig
        elif worker_type == WorkerType.OPENSHIFT:
            self.configs[worker_type] = OpenShiftWorkerConfig(**config)
        else:
            raise ValueError("Tipo de worker no soportado")

    def get_job_launcher(self, worker_type: WorkerType):
        if worker_type == WorkerType.KUBERNETES:
            if not self.kubernetes_client:
                self.initialize_kubernetes()
            return KubernetesJobLauncher(self.kubernetes_client['batch_v1'], self.configs[worker_type])
        elif worker_type == WorkerType.DOCKER:
            if not self.docker_client:
                self.initialize_docker()
            return DockerJobLauncher(self.docker_client)
        elif worker_type == WorkerType.PODMAN:
            if not self.podman_client:
                self.initialize_podman()
            return PodmanJobLauncher(self.podman_client)
        elif worker_type == WorkerType.OPENSHIFT:
            if not self.openshift_client:
                self.initialize_openshift()
            worker_type = self.configs[worker_type]
            logger.info(f"Worker type: {worker_type}")
            return OpenShiftJobLauncher(worker_type)
        else:
            raise ValueError("Tipo de worker no soportado")

    def get_job_monitor(self, worker_type: WorkerType):
        if worker_type == WorkerType.KUBERNETES:
            if not self.kubernetes_client:
                self.initialize_kubernetes()
            return KubernetesJobMonitor(self.kubernetes_client['batch_v1'])
        elif worker_type == WorkerType.DOCKER:
            if not self.docker_client:
                self.initialize_docker()
            return DockerJobMonitor(self.docker_client)
        elif worker_type == WorkerType.PODMAN:
            if not self.podman_client:
                self.initialize_podman()
            return PodmanJobMonitor(self.podman_client)
        elif worker_type == WorkerType.OPENSHIFT:
            if not self.openshift_client:
                self.initialize_openshift()
            return OpenShiftJobMonitor(self.openshift_client)
        else:
            raise ValueError("Tipo de worker no soportado")

    def get_log_streamer(self, worker_type: WorkerType):
        if worker_type == WorkerType.KUBERNETES:
            if not self.kubernetes_client:
                self.initialize_kubernetes()
            return KubernetesLogStreamer(self.kubernetes_client['core_v1'])
        elif worker_type == WorkerType.DOCKER:
            if not self.docker_client:
                self.initialize_docker()
            return DockerLogStreamer(self.docker_client)
        elif worker_type == WorkerType.PODMAN:
            if not self.podman_client:
                self.initialize_podman()
            return PodmanLogStreamer(self.podman_client)
        elif worker_type == WorkerType.OPENSHIFT:
            if not self.openshift_client:
                self.initialize_openshift()
            return OpenShiftLogStreamer(self.openshift_client)
        else:
            raise ValueError("Tipo de worker no soportado")
