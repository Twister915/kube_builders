package kubernetes_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Namespace", func() {
	const (
		namespace           = "test"

		annotation      = "test.spectonic.com/kube.annotation"
		annotationValue = "yes"
	)

	var fakeKubernetes kubernetes.Interface
	var kubeTarget *KubeTarget
	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
	})

	It("generates the correct namespace", func() {
		ns := kubeTarget.CreateNamespace(namespace).AsKube()
		Expect(ns.Name).To(Equal(namespace))
	})

	It("adds annotations", func() {
		namespace := kubeTarget.CreateNamespace(namespace).Annotation(annotation, annotationValue).AsKube()
		Expect(namespace.Annotations).To(HaveKeyWithValue(annotation, annotationValue))
	})

	It("pushes to kubernetes", func() {
		ns, err := kubeTarget.CreateNamespace(namespace).Annotation(annotation, annotationValue).Push()
		Expect(err).ToNot(HaveOccurred())
		Expect(ns.Name).To(Equal(namespace))
		Expect(ns.Annotations).To(HaveKeyWithValue(annotation, annotationValue))
	})

	It("creates a namespace when it does not exist", func() {
		By("pushing first time")
		var called bool
		err := kubeTarget.EnsureNamespaceExists(namespace, func(ns NamespaceBuilder) NamespaceBuilder {
			called = true
			return ns.Annotation(annotation, annotationValue)
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(called).To(BeTrue())

		By("not re-pushing when the namespace already exists")
		called = false
		err = kubeTarget.EnsureNamespaceExists(namespace, func(ns NamespaceBuilder) NamespaceBuilder {
			called = true
			return ns.Annotation(annotation, annotationValue)
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(called).To(BeFalse())
	})
})
