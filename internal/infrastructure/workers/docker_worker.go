package adapters

type DockerWorker struct {
	Name        string
	ContainerID string
	Image       string
	Command     []string
	Environment map[string]string
}

func (d *DockerWorker) GetID() string {
	return d.Name
}

func (d *DockerWorker) GetType() string {
	return "Docker"
}

func (d *DockerWorker) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"Image":       d.Image,
		"Command":     d.Command,
		"Environment": d.Environment,
	}
}
