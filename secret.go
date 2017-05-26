package kubernetes

import (
	"encoding/base64"
	"github.com/pkg/errors"
	kube_errors "k8s.io/client-go/pkg/api/errors"
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
	_, err = kube.iface.CoreV1().Secrets(namespace).Get(name)
	if kube_errors.IsNotFound(err) {
		err = nil
	} else if err == nil {
		exists = true
	}
	return
}

func (kube *KubeTarget) GetSecret(name, namespace string) (data map[string][]byte, exists bool, err error) {
	data = make(map[string][]byte)
	secret, err := kube.iface.CoreV1().Secrets(namespace).Get(name)
	if kube_errors.IsNotFound(err) {
		err = nil
		return
	} else if err != nil {
		err = errors.Wrapf(err, "getting secret")
		return
	}
	exists = true
	for key, value := range secret.Data {
		realValue, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			err = errors.Wrapf(err, "could not decode secret data at %s", key)
		}
		data[key] = realValue
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
		kubeSecret.Data = make(map[string][]byte)
		for key, value := range secret.keys {
			d := base64.StdEncoding.EncodeToString([]byte(value))
			kubeSecret.Data[key] = []byte(d)
		}
	}
	return
}

func (secret SecretBuilder) Push() (kubeSecret *v1.Secret, err error) {
	kubeSecret = secret.AsKube()

	secrets := secret.kube.iface.CoreV1().Secrets(secret.namespace)
	exists, err := secret.kube.DoesSecretExist(secret.namespace, secret.name)
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
		err = errors.Wrapf(err, "creating secret %s", secret.name)
	}
	return
}
