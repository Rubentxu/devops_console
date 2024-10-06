from abc import ABC, abstractmethod
from typing import Dict, Any
from pydantic import BaseModel

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
    command: str = None
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
    command: str = None
    environment: Dict[str, str] = {}
    mounts: Dict[str, Dict[str, str]] = {}

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
        return KubernetesConfig(**config)
    elif worker_type == "docker":
        return DockerConfig(**config)
    elif worker_type == "podman":
        return PodmanConfig(**config)
    else:
        raise ValueError(f"Unsupported worker type: {worker_type}")