package adapters

type KubernetesWorker struct {
	Name        string
	JobName     string
	Namespace   string
	Image       string
	Command     []string
	Environment map[string]string
}

func (k *KubernetesWorker) GetID() string {
	return k.Name
}

func (k *KubernetesWorker) GetType() string {
	return "Kubernetes"
}

func (k *KubernetesWorker) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"Namespace":   k.Namespace,
		"Image":       k.Image,
		"Command":     k.Command,
		"Environment": k.Environment,
	}
}
