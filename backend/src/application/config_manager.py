from typing import Dict, Any
from src.infrastructure.workers.worker_config import get_worker_config, WorkerConfig
from src.infrastructure.workers.factories.worker_factory import WorkerFactory

class ConfigManager:
    def __init__(self, initial_config: Dict[str, Any]):
        self.config = initial_config
        self.worker_config = get_worker_config(self.config['worker'])
        self.worker_factory = WorkerFactory(self.worker_config)

    def update_config(self, new_config: Dict[str, Any]):
        self.config.update(new_config)
        self.worker_config = get_worker_config(self.config['worker'])
        self.worker_factory = WorkerFactory(self.worker_config)

    def get_worker_factory(self) -> WorkerFactory:
        return self.worker_factory

    def get_current_config(self) -> Dict[str, Any]:
        return self.config