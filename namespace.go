package kube_builders

import (
	"github.com/pkg/errors"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

func (kube *KubeTarget) EnsureNamespaceExists(name string, b func(builder NamespaceBuilder) NamespaceBuilder) (err error) {
	_, err = kube.iface.CoreV1().Namespaces().Get(name, meta_v1.GetOptions{})

	if kube_errors.IsNotFound(err) {
		builder := kube.CreateNamespace(name)
		if b != nil {
			builder = b(builder)
		}
		_, err = builder.Push()
	} else if err != nil {
		err = errors.Wrapf(err, "error finding namespace %s", name)
	}

	return
}

type NamespaceBuilder struct {
	kube        *KubeTarget
	name        string
	labels      map[string]string
	annotations map[string]string
}

func (kube *KubeTarget) CreateNamespace(name string) NamespaceBuilder {
	if len(name) == 0 {
		panic("invalid name for namespace passed")
	}

	return NamespaceBuilder{name: name, kube: kube}
}

func (ns NamespaceBuilder) Label(label string, value interface{}) NamespaceBuilder {
	setAtMap(&ns.labels, label, value)
	return ns
}

func (ns NamespaceBuilder) Annotation(annotation string, value interface{}) NamespaceBuilder {
	setAtMap(&ns.annotations, annotation, value)
	return ns
}

func (ns NamespaceBuilder) AsKube() (kubeNs *v1.Namespace) {
	kubeNs = new(v1.Namespace)
	kubeNs.Name = ns.name
	kubeNs.Labels = ns.labels
	kubeNs.Annotations = ns.annotations
	return
}

func (ns NamespaceBuilder) Push() (kubeNs *v1.Namespace, err error) {
	kubeNs = ns.AsKube()
	err = PushNamespace(kubeNs, ns.kube.iface)
	return
}

func PushNamespace(kubeNs *v1.Namespace, iface kubernetes.Interface) (err error) {
	_, err = iface.CoreV1().Namespaces().Create(kubeNs)
	if err != nil {
		err = errors.Wrapf(err, "creating namespace %s", kubeNs.Name)
	}
	return
}
