package kube_builders_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Pod Builder", func() {
	const (
		namespace = "test"
		name      = "test"

		containerName     = "web"
		containerImage    = "docker.spectonic.com/test/123"
		containerPort     = 80
		containerPortName = "http"
	)

	var (
		fakeKubernetes kubernetes.Interface
		kubeTarget     *KubeTarget
	)

	tinyPod := func() PodBuilder {
		return kubeTarget.NewPod(name, namespace).Container(containerName, containerImage, func(ctr ContainerBuilder) ContainerBuilder {
			return ctr.Port(containerPort, containerPortName)
		})
	}

	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
	})

	It("creates a pod", func() {
		pod := tinyPod().AsKube()
		By("copying pod metadata")
		Expect(pod.Name).To(Equal(name))
		Expect(pod.Namespace).To(Equal(namespace))

		By("having a container")
		Expect(pod.Spec.Containers).To(HaveLen(1))
		c := pod.Spec.Containers[0]

		By("creating a valid container")
		Expect(c.Name).To(Equal(containerName))
		Expect(c.Image).To(Equal(containerImage))
		Expect(c.Ports).To(HaveLen(1))

		By("creating the right port")
		p := c.Ports[0]
		Expect(p.Name).To(Equal(containerPortName))
		Expect(p.ContainerPort).To(BeEquivalentTo(containerPort))
	})
})
