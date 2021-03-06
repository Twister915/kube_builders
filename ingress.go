package kube_builders

import (
	"github.com/pkg/errors"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type IngressBuilder struct {
	kube *KubeTarget

	name      string
	namespace string
	host      string

	paths     map[string]ingressServiceTarget
	tlsSecret string

	labels      map[string]string
	annotations map[string]string
}

type ingressServiceTarget struct {
	service string
	port    int
}

func (kube *KubeTarget) Ingress(name, namespace, domain string) IngressBuilder {
	return IngressBuilder{kube: kube, name: name, namespace: namespace, host: domain}
}

func (ing IngressBuilder) Path(path, service string, port int) IngressBuilder {
	setAtMapDirect(&ing.paths, path, ingressServiceTarget{service: service, port: port})
	return ing
}

func (ing IngressBuilder) TLS(secret string) IngressBuilder {
	ing.tlsSecret = secret
	return ing
}

func (ing IngressBuilder) TLSAcme() IngressBuilder {
	return ing.Annotation("kubernetes.io/tls-acme", "true")
}

func (ing IngressBuilder) Label(label string, value interface{}) IngressBuilder {
	setAtMap(&ing.labels, label, value)
	return ing
}

func (ing IngressBuilder) Annotation(annotation string, value interface{}) IngressBuilder {
	setAtMap(&ing.annotations, annotation, value)
	return ing
}

func (ing IngressBuilder) AsKube() (kubeIng *v1beta1.Ingress) {
	kubeIng = new(v1beta1.Ingress)

	kubeIng.Name = ing.name
	kubeIng.Namespace = ing.namespace
	for path, service := range ing.paths {
		kubeIng.Spec.Rules = append(kubeIng.Spec.Rules,
			v1beta1.IngressRule{Host: ing.host,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{Path: path,
								Backend: v1beta1.IngressBackend{ServiceName: service.service,
									ServicePort: intstr.FromInt(service.port)}},
						},
					},
				}})
	}
	kubeIng.Annotations = ing.annotations
	kubeIng.Labels = ing.labels
	if len(ing.tlsSecret) > 0 {
		kubeIng.Spec.TLS = []v1beta1.IngressTLS{{Hosts: []string{ing.host}, SecretName: ing.tlsSecret}}
	}
	return
}

func (ing IngressBuilder) Push() (kubeIng *v1beta1.Ingress, err error) {
	kubeIng = ing.AsKube()
	err = PushIngress(kubeIng, ing.kube.iface)
	return
}

func PushIngress(kubeIng *v1beta1.Ingress, iface kubernetes.Interface) (err error) {
	ingresses := iface.ExtensionsV1beta1().Ingresses(kubeIng.Namespace)
	foundIng, err := ingresses.Get(kubeIng.Name, meta_v1.GetOptions{})
	var f func(*v1beta1.Ingress) (*v1beta1.Ingress, error)
	if kube_errors.IsNotFound(err) {
		f = ingresses.Create
	} else if err != nil {
		return
	} else {
		foundIng.Spec = kubeIng.Spec
		kubeIng = foundIng
		f = ingresses.Update
	}
	_, err = f(kubeIng)
	if err != nil {
		err = errors.Wrapf(err, "failed to create ingress")
	}
	return
}
