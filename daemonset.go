package kubernetes

import (
	"github.com/pkg/errors"
	kube_errors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type DaemonSetBuilder struct {
	kube *KubeTarget

	name      string
	namespace string

	pod         v1.Pod
	labels      map[string]string
	annotations map[string]string
}

func (pod PodBuilder) DaemonSet(name string) DaemonSetBuilder {
	return DaemonSetBuilder{name: name, namespace: pod.namespace, pod: *pod.AsKube(), kube: pod.kube}
}

func (ds DaemonSetBuilder) Label(label string, value interface{}) DaemonSetBuilder {
	setAtMap(&ds.labels, label, value)
	return ds
}

func (ds DaemonSetBuilder) Annotation(annotation string, value interface{}) DaemonSetBuilder {
	setAtMap(&ds.annotations, annotation, value)
	return ds
}

func (ds DaemonSetBuilder) AsKube() (kubeDs *v1beta1.DaemonSet) {
	kubeDs = new(v1beta1.DaemonSet)
	kubeDs.Name = ds.name
	kubeDs.Namespace = ds.namespace

	kubeDs.Spec.Template.Spec = ds.pod.Spec
	kubeDs.Spec.Template.Annotations = ds.pod.Annotations
	kubeDs.Spec.Template.Labels = ds.pod.Labels

	kubeDs.Labels = ds.labels
	kubeDs.Annotations = ds.annotations
	return
}

func (ds DaemonSetBuilder) Push() (kubeDs *v1beta1.DaemonSet, err error) {
	kubeDs = ds.AsKube()
	dses := ds.kube.iface.ExtensionsV1beta1().DaemonSets(ds.namespace)

	_, err = dses.Get(ds.name)
	if kube_errors.IsNotFound(err) {
		_, err = dses.Create(kubeDs)
		if err != nil {
			err = errors.Wrapf(err, "failed to create daemon set")
		}
	} else if err != nil {
		err = errors.Wrapf(err, "failed to get current daemon set")
	} else {
		_, err = dses.Update(kubeDs)
		if err != nil {
			err = errors.Wrapf(err, "failed to update daemon set")
		}
	}

	return
}
