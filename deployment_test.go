package kubernetes_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Deployment Builder", func() {
	const (
		namespace = "test"
		name      = "test"

		containerName     = "web"
		containerImage    = "docker.spectonic.com/test/123"
		containerPort     = 80
		containerPortName = "http"

		secretName = "secret"
		envValue   = "HI"

		annotationName  = "hi"
		annotationValue = "yo"

		labelName  = "test"
		labelValue = "123"

		replicas = 2
		history  = 3
	)

	var (
		secretKeys     = []string{"test", "test2", "magic"}
		secretEnvs     = []string{"TEST", "TEST_2", "MAGIC"}
		envVars        = []string{"SOMETHING", "SOMETHING_ELSE"}
		fakeKubernetes kubernetes.Interface
		kubeTarget     *KubeTarget
	)

	tinyDeploy := func() DeploymentBuilder {
		return kubeTarget.NewPod("", namespace).Container(containerName, containerImage, func(ctr ContainerBuilder) ContainerBuilder {
			return ctr.Port(containerPort, containerPortName)
		}).Deployment(name).History(history).Replicas(replicas).Label(labelName, labelValue).Annotation(annotationName, annotationValue)
	}

	bigDeploy := func() DeploymentBuilder {
		return kubeTarget.NewPod("", namespace).Container(containerName, containerImage, func(ctr ContainerBuilder) ContainerBuilder {
			ctr = ctr.Port(containerPort, containerPortName)
			for _, envVar := range envVars {
				ctr = ctr.Env(envVar, envValue)
			}
			for i, secret := range secretKeys {
				ctr = ctr.Secret(secretEnvs[i], secretName, secret)
			}
			return ctr
		}).Deployment(name).History(history).Replicas(replicas).Label(labelName, labelValue).Annotation(annotationName, annotationValue)
	}

	BeforeEach(func() {
		fakeKubernetes = fake.NewSimpleClientset()
		kubeTarget = NewKubeTarget(fakeKubernetes)
	})

	It("creates a deployment", func() {
		deployment := tinyDeploy().AsKube()

		By("setting the metadata correctly")
		Expect(deployment.Name).To(Equal(name))
		Expect(deployment.Namespace).To(Equal(namespace))
		Expect(deployment.Annotations).To(HaveKeyWithValue(annotationName, annotationValue))
		Expect(deployment.Labels).To(HaveKeyWithValue(labelName, labelValue))
		Expect(deployment.Spec.Replicas).ToNot(BeNil())
		Expect(*deployment.Spec.Replicas).To(BeEquivalentTo(replicas))
		Expect(deployment.Spec.RevisionHistoryLimit).ToNot(BeNil())
		Expect(*deployment.Spec.RevisionHistoryLimit).To(BeEquivalentTo(history))
		containers := deployment.Spec.Template.Spec.Containers
		By("creating containers")
		Expect(containers).To(HaveLen(1))
		c := containers[0]
		Expect(c.Name).To(Equal(containerName))
		Expect(c.Image).To(Equal(containerImage))
		By("creating container ports")
		Expect(c.Ports).To(HaveLen(1))
		By("creating no container volumes")
		Expect(c.VolumeMounts).To(HaveLen(0))
	})

	It("can create a deployment with secrets/env vars", func() {
		deployment := bigDeploy().AsKube()

		c := deployment.Spec.Template.Spec.Containers[0]
		Expect(c.Ports).To(HaveLen(1))
		Expect(c.Env).To(HaveLen(len(secretEnvs) + len(envVars)))
	})

	It("can push to kubernetes", func() {
		_, err := bigDeploy().Push()
		Expect(err).ToNot(HaveOccurred())
	})
})
