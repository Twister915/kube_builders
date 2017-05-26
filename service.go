package kube_builders

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"
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
	port int
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

func (svc ServiceBuilder) AsKube() (service *v1.Service) {
	service = new(v1.Service)
	service.Name = svc.name
	service.Namespace = svc.namespace
	service.Annotations = svc.annotations
	service.Labels = svc.labels
	service.Spec.Type = svc.sType
	service.Spec.Selector = svc.selector
	if svc.ports != nil {
		for portName, port := range svc.ports {
			service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{Name: portName, TargetPort: port.target, Port: int32(port.port)})
		}
	}
	return
}

func (svc ServiceBuilder) Push() (service *v1.Service, err error) {
	service = svc.AsKube()
	_, err = svc.kube.iface.CoreV1().Services(svc.namespace).Create(service)
	if err != nil {
		err = errors.Wrapf(err, "creating service %s", svc.name)
	}
	return
}
