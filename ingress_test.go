package kubernetes_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Ingress Builder", func() {
	const (

		name      = "test-ing"
		namespace = "test"

		serviceName = "test"
		servicePort = 80

		domain = "test.spectonic.com"
		path   = "/"
	)

	var fakeKubernetes kubernetes.Interface
	var kubeTarget *KubeTarget
	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
	})

	It("generates a valid ingress", func() {
		ingress := kubeTarget.Ingress(name, namespace, domain).Path(path, serviceName, servicePort).AsKube()
		By("generating the right metadata")
		Expect(ingress.Namespace).To(Equal(namespace))
		Expect(ingress.Name).To(Equal(name))

		By("not generating TLS")
		Expect(ingress.Spec.TLS).To(BeEmpty())

		By("correctly generating the rule")
		Expect(ingress.Spec.Rules).To(HaveLen(1))
		rule := ingress.Spec.Rules[0]
		Expect(rule.HTTP).ToNot(BeNil())
		Expect(rule.HTTP.Paths).To(HaveLen(1))
		pathFromGen := rule.HTTP.Paths[0]
		Expect(pathFromGen.Path).To(Equal(path))
		Expect(pathFromGen.Backend.ServiceName).To(Equal(serviceName))
		Expect(pathFromGen.Backend.ServicePort.IntValue()).To(Equal(servicePort))
	})

	It("pushes to kubernetes", func() {
		ingress, err := kubeTarget.Ingress(name, namespace, domain).Path(path, serviceName, servicePort).Push()
		Expect(err).ToNot(HaveOccurred())
		Expect(ingress.Name).To(Equal(name))
	})
})
