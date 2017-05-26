package kube_builders

import "k8s.io/client-go/pkg/api/v1"

type VolumeBuilder struct {
	name string

	hostPath *string
}

func NewVolumeBuilder(name string) VolumeBuilder {
	return VolumeBuilder{name: name}
}

func (v VolumeBuilder) HostPath(path string) VolumeBuilder {
	v.hostPath = new(string)
	*v.hostPath = path
	return v
}

func (v VolumeBuilder) AsKube() (kubeVolume v1.Volume) {
	kubeVolume.Name = v.name
	if v.hostPath != nil {
		kubeVolume.HostPath = new(v1.HostPathVolumeSource)
		kubeVolume.HostPath.Path = *v.hostPath
		return
	}

	panic("unknown volume type (none defined?)")
}