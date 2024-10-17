from abc import ABC, abstractmethod
from typing import Dict, Any
from pydantic import BaseModel
from .worker_types import WorkerType

import logging

# Configuración básica del logger
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class WorkerConfig:
    def __init__(self, config: Dict[str, Any]):
        self.config = config

    def get_launch_config(self) -> Dict[str, Any]:
        return self.config

class KubernetesWorkerConfig(WorkerConfig):
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.namespace = config.get('namespace', 'default')
        self.job_template = config.get('job_template', {})

class DockerWorkerConfig(WorkerConfig):
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.image = config['image']
        self.command = config['command']
        self.environment = config.get('environment', {})
        self.volumes = config.get('volumes', {})

class PodmanWorkerConfig(WorkerConfig):
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.image = config['image']
        self.command = config['command']
        self.environment = config.get('environment', {})
        self.mounts = config.get('mounts', {})

class OpenShiftWorkerConfig(WorkerConfig):
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.project = config.get('project', 'default')
        self.job_template = config.get('job_template', {})

def get_worker_config(worker_data: Dict[str, Any]) -> WorkerConfig:
    worker_type = worker_data['type'].upper()
    config = worker_data['config']

    if worker_type == 'KUBERNETES':
        return KubernetesWorkerConfig(config)
    elif worker_type == 'DOCKER':
        return DockerWorkerConfig(config)
    elif worker_type == 'PODMAN':
        return PodmanWorkerConfig(config)
    elif worker_type == 'OPENSHIFT':
        return OpenShiftWorkerConfig(config)
    else:
        raise ValueError(f"Unsupported worker type: {worker_type}")

