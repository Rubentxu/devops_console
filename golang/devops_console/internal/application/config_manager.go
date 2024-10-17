package application

import (
	"devops_console/internal/infrastructure/workers"
	"devops_console/internal/infrastructure/workers/factories"
)

type ConfigManager struct {
	config        map[string]interface{}
	workerConfig  workers.WorkerConfig
	workerFactory factories.WorkerFactory
}

func NewConfigManager(initialConfig map[string]interface{}) *ConfigManager {
	workerConfig := workers.GetWorkerConfig(initialConfig["worker"])
	return &ConfigManager{
		config:        initialConfig,
		workerConfig:  workerConfig,
		workerFactory: factories.NewWorkerFactory(workerConfig),
	}
}

func (cm *ConfigManager) UpdateConfig(newConfig map[string]interface{}) {
	for k, v := range newConfig {
		cm.config[k] = v
	}
	cm.workerConfig = workers.GetWorkerConfig(cm.config["worker"])
	cm.workerFactory = factories.NewWorkerFactory(cm.workerConfig)
}

func (cm *ConfigManager) GetWorkerFactory() factories.WorkerFactory {
	return cm.workerFactory
}

func (cm *ConfigManager) GetCurrentConfig() map[string]interface{} {
	return cm.config
}
