package kube_builders_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

var _ = Describe("Daemonset Builder", func() {
	const (
		namespace           = "test-ns"

		containerName     = "web"
		containerImage    = "test"
		containerPort     = 80
		containerPortName = "http"

		name = "test-ds"
	)

	var fakeKubernetes kubernetes.Interface
	var kubeTarget *KubeTarget
	var pod PodBuilder

	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
		pod = kubeTarget.
			NewPod("", namespace).
			Container(containerName, containerImage, func(container ContainerBuilder) ContainerBuilder {
				return container.Port(containerPort, containerPortName)
			})
	})

	It("sets name and namespace correctly", func() {
		ds := pod.DaemonSet(name).AsKube()
		Expect(ds.Name).To(Equal(name))
		Expect(ds.Namespace).To(Equal(namespace))
	})

	It("generates the correct pod", func() {
		ds := pod.DaemonSet(name).AsKube()
		podSpec := ds.Spec.Template

		By("having containers")
		containers := podSpec.Spec.Containers
		Expect(containers).To(HaveLen(1))

		container := containers[0]
		By("having ports")
		Expect(container.Ports).To(HaveLen(1))
		Expect(container.Ports).To(ContainElement(v1.ContainerPort{Name: containerPortName, ContainerPort: containerPort}))
	})

	It("pushes to kubernetes", func() {
		By("pushing")
		_, err := pod.DaemonSet(name).Push()
		Expect(err).ToNot(HaveOccurred())

		By("being on kubernetes")
		_, err = fakeKubernetes.ExtensionsV1beta1().DaemonSets(namespace).Get(name)
		Expect(err).ToNot(HaveOccurred())
	})
})
