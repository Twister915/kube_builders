package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const dockerVolumeName = "docker.sock"
const webPort = 80
const dockerPullSecret = "docker"

func NewKubernetes(ca, cert, key []byte, username, password, address string) (t *KubeTarget, err error) {
	cfg := &rest.Config{Host: address, Username: username, Password: password}
	cfg.CAData = ca
	cfg.CertData = cert
	cfg.KeyData = key
	k, err := kube.NewForConfig(cfg)
	if err != nil {
		return
	}
	t = NewKubeTarget(k)
	return
}

func NewKubeTarget(iface kubernetes.Interface) *KubeTarget {
	return &KubeTarget{iface: iface}
}

type KubeTarget struct {
	iface kubernetes.Interface
}

