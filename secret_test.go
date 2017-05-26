package kube_builders_test

import (
	. "github.com/Twister915/kube_builders"

	"encoding/base64"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
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
			Expect(kubeSecret.Data).To(HaveKeyWithValue(key, []byte(base64.StdEncoding.EncodeToString([]byte(value)))))
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

		By("existing when requested")
		secretData, exists, err := kubeTarget.GetSecret(secretName, namespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(exists).To(BeTrue())
		for key, value := range secretData {
			Expect(secretData).To(HaveKeyWithValue(key, value))
		}
	})
})
