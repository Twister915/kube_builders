package kubernetes

import "k8s.io/client-go/pkg/api/v1"

type ContainerBuilder struct {
	name  string
	image string

	envString map[string]string
	envSecret map[string]*v1.EnvVarSource
	ports     map[string]uint16

	mountDocker bool
}

type SecretRef struct {
	Name string
	Key  string
}

func NewContainer(name, image string) ContainerBuilder {
	return ContainerBuilder{name: name, image: image}
}

func (container ContainerBuilder) Env(name string, value interface{}) ContainerBuilder {
	setAtMap(&container.envString, name, value)
	return container
}

func (container ContainerBuilder) Secret(name, secretName, secretKey string) ContainerBuilder {
	var selector v1.SecretKeySelector
	selector.Name = secretName
	selector.Key = secretKey

	setAtMapDirect(&container.envSecret, name, &v1.EnvVarSource{SecretKeyRef: &selector})
	return container
}

func (container ContainerBuilder) MountDocker(value bool) ContainerBuilder {
	container.mountDocker = value
	return container
}

func (container ContainerBuilder) Port(num int, name string) ContainerBuilder {
	setAtMapDirect(&container.ports, name, uint16(num))
	return container
}

func (container ContainerBuilder) AsKube() (kubeContainer v1.Container) {
	kubeContainer.Name = container.name
	kubeContainer.Image = container.image

	envTarget := &kubeContainer.Env
	if container.envString != nil {
		for name, value := range container.envString {
			*envTarget = append(*envTarget, v1.EnvVar{Name: name, Value: value})
		}
	}

	if container.envSecret != nil {
		for name, value := range container.envSecret {
			*envTarget = append(*envTarget, v1.EnvVar{Name: name, ValueFrom: value})
		}
	}

	if container.ports != nil {
		portsTarget := &kubeContainer.Ports
		for name, port := range container.ports {
			*portsTarget = append(*portsTarget, v1.ContainerPort{ContainerPort: int32(port), Name: name})
		}
	}

	if container.mountDocker {
		kubeContainer.VolumeMounts = append(kubeContainer.VolumeMounts, v1.VolumeMount{
			Name:      dockerVolumeName,
			MountPath: "/var/run/docker.sock",
			ReadOnly:  true,
		})
	}

	return
}
