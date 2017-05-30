package kube_builders

import (
	"github.com/pkg/errors"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

type ServiceBuilder struct {
	kube *KubeTarget

	name      string
	namespace string

	sType    v1.ServiceType
	selector map[string]string

	labels      map[string]string
	annotations map[string]string
	ports       map[string]portSpec
}

type portSpec struct {
	target intstr.IntOrString
	port   int
}

func (kube *KubeTarget) Service(name, namespace string) ServiceBuilder {
	return ServiceBuilder{kube: kube, name: name, namespace: namespace}
}

func (svc ServiceBuilder) Type(sType v1.ServiceType) ServiceBuilder {
	svc.sType = sType
	return svc
}

func (svc ServiceBuilder) Selector(name, value string) ServiceBuilder {
	setAtMap(&svc.selector, name, value)
	return svc
}

func (svc ServiceBuilder) Label(label string, value interface{}) ServiceBuilder {
	setAtMap(&svc.labels, label, value)
	return svc
}

func (svc ServiceBuilder) Annotation(annotation string, value interface{}) ServiceBuilder {
	setAtMap(&svc.annotations, annotation, value)
	return svc
}

func (svc ServiceBuilder) PortByNumber(name string, target, port int) ServiceBuilder {
	setAtMapDirect(&svc.ports, name, portSpec{port: port, target: intstr.FromInt(target)})
	return svc
}

func (svc ServiceBuilder) PortByName(name, target string, port int) ServiceBuilder {
	setAtMapDirect(&svc.ports, name, portSpec{port: port, target: intstr.FromString(target)})
	return svc
}

func (svc ServiceBuilder) AsKube() (kubeSvc *v1.Service) {
	kubeSvc = new(v1.Service)
	kubeSvc.Name = svc.name
	kubeSvc.Namespace = svc.namespace
	kubeSvc.Annotations = svc.annotations
	kubeSvc.Labels = svc.labels
	kubeSvc.Spec.Type = svc.sType
	kubeSvc.Spec.Selector = svc.selector
	if svc.ports != nil {
		for portName, port := range svc.ports {
			kubeSvc.Spec.Ports = append(kubeSvc.Spec.Ports, v1.ServicePort{Name: portName, TargetPort: port.target, Port: int32(port.port)})
		}
	}
	return
}

func (svc ServiceBuilder) Push() (kubeSvc *v1.Service, err error) {
	kubeSvc = svc.AsKube()
	err = PushService(kubeSvc, svc.kube.iface)
	return
}

func PushService(kubeSvc *v1.Service, iface kubernetes.Interface) (err error) {
	services := iface.CoreV1().Services(kubeSvc.Namespace)
	svcFromKube, err := services.Get(kubeSvc.Name, meta_v1.GetOptions{})
	var f func(*v1.Service) (*v1.Service, error)
	if kube_errors.IsNotFound(err) {
		f = services.Create
	} else if err != nil {
		return
	} else {
		f = services.Update
		svcFromKube.Spec.Ports = kubeSvc.Spec.Ports
		svcFromKube.Spec.Type = kubeSvc.Spec.Type
		kubeSvc = svcFromKube
	}
	_, err = f(kubeSvc)
	if err != nil {
		err = errors.Wrapf(err, "creating service %s", kubeSvc.Name)
	}
	return
}
