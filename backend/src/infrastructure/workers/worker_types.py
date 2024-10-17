from enum import Enum, auto

class WorkerType(Enum):
    KUBERNETES = auto()
    DOCKER = auto()
    PODMAN = auto()
    OPENSHIFT = auto()
