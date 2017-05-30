package kube_builders

import (
	"github.com/pkg/errors"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type DeploymentBuilder struct {
	kube *KubeTarget

	name      string
	namespace string

	replicas, history int

	pod         v1.Pod
	labels      map[string]string
	annotations map[string]string
}

func (pod PodBuilder) Deployment(name string) (deployment DeploymentBuilder) {
	deployment.kube = pod.kube
	deployment.pod = *pod.AsKube()
	deployment.name = name
	deployment.namespace = pod.namespace
	return
}

func (deployment DeploymentBuilder) Replicas(num int) DeploymentBuilder {
	deployment.replicas = num
	return deployment
}

func (deployment DeploymentBuilder) Label(label string, value interface{}) DeploymentBuilder {
	setAtMap(&deployment.labels, label, value)
	return deployment
}

func (deployment DeploymentBuilder) Annotation(annotation string, value interface{}) DeploymentBuilder {
	setAtMap(&deployment.annotations, annotation, value)
	return deployment
}

func (deployment DeploymentBuilder) History(count int) DeploymentBuilder {
	deployment.history = count
	return deployment
}

func (deployment DeploymentBuilder) AsKube() (kubeDeployment *v1beta1.Deployment) {
	kubeDeployment = new(v1beta1.Deployment)
	kubeDeployment.Name = deployment.name
	kubeDeployment.Namespace = deployment.namespace
	kubeDeployment.Annotations = deployment.annotations
	kubeDeployment.Labels = deployment.labels

	if deployment.replicas > 0 {
		kubeDeployment.Spec.Replicas = new(int32)
		*kubeDeployment.Spec.Replicas = int32(deployment.replicas)
	}

	if deployment.history > 0 {
		kubeDeployment.Spec.RevisionHistoryLimit = new(int32)
		*kubeDeployment.Spec.RevisionHistoryLimit = int32(deployment.history)
	}

	kubeDeployment.Spec.Template.Spec = deployment.pod.Spec
	kubeDeployment.Spec.Template.ObjectMeta.Labels = deployment.pod.Labels
	kubeDeployment.Spec.Template.ObjectMeta.Annotations = deployment.pod.Annotations
	return
}

func (deployment DeploymentBuilder) Push() (kubeDeployment *v1beta1.Deployment, err error) {
	kubeDeployment = deployment.AsKube()
	deployments := deployment.kube.iface.ExtensionsV1beta1().Deployments(deployment.namespace)

	_, err = deployments.Get(deployment.name, meta_v1.GetOptions{})
	if kube_errors.IsNotFound(err) {
		_, err = deployments.Create(kubeDeployment)
		if err != nil {
			err = errors.Wrapf(err, "failed to create deployment")
		}
	} else if err != nil {
		err = errors.Wrapf(err, "failed to get current deployments")
	} else {
		_, err = deployments.Update(kubeDeployment)
		if err != nil {
			err = errors.Wrapf(err, "failed to update deployment")
		}
	}

	return
}
