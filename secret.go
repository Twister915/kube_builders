package kube_builders

import (
	"github.com/pkg/errors"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

type SecretBuilder struct {
	kube *KubeTarget

	name      string
	namespace string
	keys      map[string]string

	labels      map[string]string
	annotations map[string]string
}

func (kube *KubeTarget) DoesSecretExist(name, namespace string) (exists bool, err error) {
	return DoesSecretExist(namespace, name, kube.iface)
}

func (kube *KubeTarget) GetSecret(name, namespace string) (data map[string][]byte, exists bool, err error) {
	data = make(map[string][]byte)
	secret, err := kube.iface.CoreV1().Secrets(namespace).Get(name, meta_v1.GetOptions{})
	if kube_errors.IsNotFound(err) {
		err = nil
		return
	} else if err != nil {
		err = errors.Wrapf(err, "getting secret")
		return
	}
	exists = true
	for key, value := range secret.Data {
		data[key] = value
	}
	return
}

func (kube *KubeTarget) NewSecret(name, namespace string) SecretBuilder {
	return SecretBuilder{kube: kube, name: name, namespace: namespace}
}

func (secret SecretBuilder) Value(key string, value interface{}) SecretBuilder {
	setAtMap(&secret.keys, key, value)
	return secret
}

func (secret SecretBuilder) Label(label string, value interface{}) SecretBuilder {
	setAtMap(&secret.labels, label, value)
	return secret
}

func (secret SecretBuilder) Annotation(annotation string, value interface{}) SecretBuilder {
	setAtMap(&secret.annotations, annotation, value)
	return secret
}

func (secret SecretBuilder) AsKube() (kubeSecret *v1.Secret) {
	kubeSecret = new(v1.Secret)
	kubeSecret.Name = secret.name
	kubeSecret.Namespace = secret.namespace
	kubeSecret.Labels = secret.labels
	kubeSecret.Annotations = secret.annotations
	if secret.keys != nil {
		kubeSecret.StringData = make(map[string]string)
		for key, value := range secret.keys {
			kubeSecret.StringData[key] = value
		}
	}
	return
}

func (secret SecretBuilder) Push() (kubeSecret *v1.Secret, err error) {
	kubeSecret = secret.AsKube()
	err = PushSecret(kubeSecret, secret.kube.iface)
	return
}

func PushSecret(kubeSecret *v1.Secret, iface kubernetes.Interface) (err error) {
	secrets := iface.CoreV1().Secrets(kubeSecret.Namespace)
	exists, err := DoesSecretExist(kubeSecret.Namespace, kubeSecret.Name, iface)
	if err != nil {
		err = errors.Wrapf(err, "could not check if secret exists")
		return
	}

	var f func(*v1.Secret) (*v1.Secret, error)
	if exists {
		f = secrets.Update
	} else {
		f = secrets.Create
	}

	_, err = f(kubeSecret)
	if err != nil {
		err = errors.Wrapf(err, "creating secret %s", kubeSecret.Name)
	}
	return
}

func DoesSecretExist(namespace, name string, iface kubernetes.Interface) (exists bool, err error) {
	_, err = iface.CoreV1().Secrets(namespace).Get(name, meta_v1.GetOptions{})
	if kube_errors.IsNotFound(err) {
		err = nil
	} else if err == nil {
		exists = true
	}
	return
}
