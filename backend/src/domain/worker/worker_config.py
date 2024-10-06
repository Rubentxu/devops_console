from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional
from pydantic import BaseModel, validator

from src.infrastructure.workers.factories.worker_factory import WorkerType

class WorkerConfig(BaseModel, ABC):
    type: str

    @abstractmethod
    def get_launch_config(self) -> Dict[str, Any]:
        pass

class KubernetesConfig(WorkerConfig):
    type: str = "kubernetes"
    namespace: str = "default"
    job_template: Dict[str, Any]

    def get_launch_config(self) -> Dict[str, Any]:
        return {
            "namespace": self.namespace,
            "job_template": self.job_template
        }

class DockerConfig(WorkerConfig):
    type: str = "docker"
    image: str
    command: str = ''
    environment: Dict[str, str] = {}
    volumes: Dict[str, Dict[str, str]] = {}

    def get_launch_config(self) -> Dict[str, Any]:
        return {
            "image": self.image,
            "command": self.command,
            "environment": self.environment,
            "volumes": self.volumes
        }

class PodmanConfig(WorkerConfig):
    type: str = "podman"
    image: str
    command: str = ''
    environment: Dict[str, str] = {}
    mounts: Dict[str, Dict[str, Any]] = {}

    @validator('mounts')
    def validate_mounts(cls, v):
        for mount in v.values():
            if 'read_only' in mount and not isinstance(mount['read_only'], bool):
                raise ValueError("'read_only' must be a boolean")
        return v

    def get_launch_config(self) -> Dict[str, Any]:
        return {
            "image": self.image,
            "command": self.command,
            "environment": self.environment,
            "mounts": self.mounts
        }


def get_worker_config(config: Dict[str, Any]) -> WorkerConfig:
    worker_type = config.get("type")
    if worker_type == "kubernetes":
        kubernetes_config = config.get("config", {})
        return KubernetesConfig(**kubernetes_config)
    elif worker_type == "docker":
        docker_config = config.get("config", {})
        return DockerConfig(**docker_config)
    elif worker_type == "podman":
        podman_config = config.get("config", {})
        return PodmanConfig(**podman_config)
    else:
        raise ValueError(f"Unsupported worker type: {worker_type}")
