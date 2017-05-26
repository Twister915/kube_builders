package kube_builders

import "k8s.io/client-go/pkg/api/v1"

type PodBuilder struct {
	kube *KubeTarget

	name      string
	namespace string

	labels      map[string]string
	annotations map[string]string

	containers             []v1.Container
	volumes                []v1.Volume
	imagePullSecrets       []v1.LocalObjectReference
	terminationGracePeriod *int
	nodeSelector           map[string]string
	restartPolicy          v1.RestartPolicy
	hostNetwork            bool

	hasMountedDocker bool
}

func (kube *KubeTarget) NewPod(name, namespace string) PodBuilder {
	return PodBuilder{kube: kube, name: name, namespace: namespace}
}

func (pod PodBuilder) Label(label string, value interface{}) PodBuilder {
	setAtMap(&pod.labels, label, value)
	return pod
}

func (pod PodBuilder) Annotation(annotation string, value interface{}) PodBuilder {
	setAtMap(&pod.annotations, annotation, value)
	return pod
}

func (pod PodBuilder) Container(name, image string, builder func(ContainerBuilder) ContainerBuilder) PodBuilder {
	builtContainer := builder(NewContainer(name, image))
	pod.containers = append(pod.containers, builtContainer.AsKube())
	if builtContainer.mountDocker && !pod.hasMountedDocker {
		pod = pod.Volume(dockerVolumeName, func(volume VolumeBuilder) VolumeBuilder {
			return volume.HostPath("/var/run/docker.sock")
		})
		pod.hasMountedDocker = true
		return pod
	}
	return pod
}

func (pod PodBuilder) Volume(name string, builder func(VolumeBuilder) VolumeBuilder) PodBuilder {
	pod.volumes = append(pod.volumes, builder(NewVolumeBuilder(name)).AsKube())
	return pod
}

func (pod PodBuilder) ImagePullSecret(name string) PodBuilder {
	pod.imagePullSecrets = append(pod.imagePullSecrets, v1.LocalObjectReference{Name: name})
	return pod
}

func (pod PodBuilder) TerminationGracePeriod(period int) PodBuilder {
	pod.terminationGracePeriod = new(int)
	*pod.terminationGracePeriod = period
	return pod
}

func (pod PodBuilder) NodeSelector(label string, value interface{}) PodBuilder {
	setAtMap(&pod.nodeSelector, label, value)
	return pod
}

func (pod PodBuilder) RestartPolicy(policy v1.RestartPolicy) PodBuilder {
	pod.restartPolicy = policy
	return pod
}

func (pod PodBuilder) HostNetwork(hostNetwork bool) PodBuilder {
	pod.hostNetwork = hostNetwork
	return pod
}

func (pod PodBuilder) AsKube() (kubePod *v1.Pod) {
	kubePod = new(v1.Pod)
	kubePod.Name = pod.name
	kubePod.Namespace = pod.namespace
	kubePod.Annotations = pod.annotations
	kubePod.Labels = pod.labels
	kubePod.Spec.Containers = pod.containers
	kubePod.Spec.Volumes = pod.volumes
	kubePod.Spec.ImagePullSecrets = pod.imagePullSecrets
	if pod.terminationGracePeriod != nil {
		kubePod.Spec.TerminationGracePeriodSeconds = new(int64)
		*kubePod.Spec.TerminationGracePeriodSeconds = int64(*pod.terminationGracePeriod)
	}
	kubePod.Spec.HostNetwork = pod.hostNetwork
	return
}
