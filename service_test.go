package kubernetes_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Service", func() {
	const (
		serviceNamespace = "test"
		serviceName      = "service"
		selectorKey      = "app"
		selectorValue    = "the-app"
		portName         = "http"
		port             = 80
	)

	var (
		fakeKubernetes kubernetes.Interface
		kubeTarget     *KubeTarget
	)

	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
	})

	It("deploys a service", func() {
		By("creating a service")
		svc := kubeTarget.Service(serviceName, serviceNamespace).Selector(selectorKey, selectorValue).PortByNumber(portName, port)
		service := svc.AsKube()
		Expect(service.Name).To(Equal(serviceName))
		Expect(service.Namespace).To(Equal(serviceNamespace))
		Expect(service.Spec.Selector).To(HaveLen(1))
		Expect(service.Spec.Selector).To(HaveKeyWithValue(selectorKey, selectorValue))
		Expect(service.Spec.Ports).To(HaveLen(1))
		p := service.Spec.Ports[0]
		Expect(p.Name).To(Equal(portName))
		Expect(p.TargetPort.IntValue()).To(BeEquivalentTo(port))

		By("deploying it to kubernetes")
		_, err := svc.Push()
		Expect(err).ToNot(HaveOccurred())
	})
})
