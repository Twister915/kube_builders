package kube_builders_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"encoding/base64"
)

var _ = Describe("Secret Generator", func() {
	const (
		secretName = "secret"
		namespace  = "secret-spy"
	)

	var (
		secretData     = map[string]string{"something": "fancy", "something-else": "fancyer"}
		fakeKubernetes kubernetes.Interface
		kubeTarget     *KubeTarget
	)

	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
	})

	It("generates a secret", func() {
		secret := kubeTarget.NewSecret(secretName, namespace)
		for key, value := range secretData {
			secret = secret.Value(key, value)
		}
		kubeSecret := secret.AsKube()

		Expect(kubeSecret.Name).To(Equal(secretName))
		Expect(kubeSecret.Namespace).To(Equal(namespace))
		for key, value := range secretData {
			Expect(kubeSecret.StringData).To(HaveKeyWithValue(key, value))
		}
	})

	It("pushes to kubernetes", func() {
		By("pushing")
		secret := kubeTarget.NewSecret(secretName, namespace)
		for key, value := range secretData {
			secret = secret.Value(key, value)
		}
		_, err := secret.Push()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can get secrets from kubernetes", func() {
		//create a secret
		By("creating a fake secret")
		secret := kubeTarget.NewSecret(secretName, namespace)
		for key, value := range secretData {
			secret = secret.Value(key, value)
		}
		kubeSecret := secret.AsKube()
		_, err := fakeKubernetes.CoreV1().Secrets(kubeSecret.Namespace).Create(kubeSecret)
		Expect(err).ToNot(HaveOccurred())
		By("making it look like kubernetes gave it to us")
		kubeSecret.Data = make(map[string][]byte)
		for key, value := range kubeSecret.StringData {
			kubeSecret.Data[key] = []byte(base64.StdEncoding.EncodeToString([]byte(value)))
		}

		By("getting it from Kubernetes")
		_, exists, err := kubeTarget.GetSecret(secretName, namespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(exists).To(BeTrue())
		By("having the right data")
		for key, value := range secretData {
			Expect(secretData).To(HaveKeyWithValue(key, value))
		}
	})
})
