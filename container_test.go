package kube_builders_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api/v1"
)

var _ = Describe("Container Builder", func() {
	const (
		name  = "test"
		image = "docker.spectonic.com/test/test"

		secret        = "test-secret"
		secretKey     = "access-key"
		secretEnvName = "SECRET_NAME"

		envName  = "ENV_NAME"
		envValue = "some value 123"

		portNumber = 80
		portName   = "http"
	)
	container := NewContainer(name, image).
		Env(envName, envValue).
		Secret(secretEnvName, secret, secretKey).
		Port(portNumber, portName)

	kubeContainer := container.AsKube()
	withDocker := container.MountDocker(true).AsKube()

	It("sets name and image correctly", func() {
		Expect(kubeContainer.Name).To(Equal(name))
		Expect(kubeContainer.Image).To(Equal(image))
	})

	It("creates enviornment variables for secret & env", func() {
		Expect(kubeContainer.Env).To(HaveLen(2))
		Expect(kubeContainer.Env).To(ContainElement(v1.EnvVar{Name: envName, Value: envValue}))
		Expect(kubeContainer.Env).To(ContainElement(v1.EnvVar{Name: secretEnvName, ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				Key: secretKey, LocalObjectReference: v1.LocalObjectReference{Name: secret}},
		}}))
	})

	It("adds an http port", func() {
		Expect(kubeContainer.Ports).To(HaveLen(1))
		Expect(kubeContainer.Ports).To(ContainElement(v1.ContainerPort{ContainerPort: portNumber, Name: "http"}))
	})

	Describe("Container with Docker", func() {
		It("mounts docker as a volume", func() {
			Expect(withDocker.VolumeMounts).To(ContainElement(v1.VolumeMount{Name: "docker.sock", MountPath: "/var/run/docker.sock", ReadOnly: true}))
		})
	})

	It("can set config map env var", func() {
		By("creating it in the builder")
		container = container.ConfigMapRef("TEST_CONFIG_MAP", "cf", "key")
		kubeContainer = container.AsKube()
		By("existing in the env array for the kube container")
		var cfSelector v1.ConfigMapKeySelector
		cfSelector.Name = "cf"
		cfSelector.Key = "key"
		Expect(kubeContainer.Env).To(ContainElement(v1.EnvVar{Name: "TEST_CONFIG_MAP", ValueFrom: &v1.EnvVarSource{ConfigMapKeyRef: &cfSelector}}))
	})

	It("can set resource env var", func() {
		By("creating it in the builder")
		container = container.ResourceRef("TEST_RESOURCE", "rsc")
		kubeContainer = container.AsKube()

		By("existing in the env array for the kube container")
		var rscSelector v1.ResourceFieldSelector
		rscSelector.Resource = "rsc"
		Expect(kubeContainer.Env).To(ContainElement(v1.EnvVar{Name: "TEST_RESOURCE", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &rscSelector}}))
	})

	It("can set field env var", func() {
		By("creating it in the builder")
		container = container.FieldRef("TEST_FIELD", "field")
		kubeContainer = container.AsKube()

		By("existing in the env array for the kube container")
		var fieldSelector v1.ObjectFieldSelector
		fieldSelector.FieldPath = "field"
		Expect(kubeContainer.Env).To(ContainElement(v1.EnvVar{Name: "TEST_FIELD", ValueFrom: &v1.EnvVarSource{FieldRef: &fieldSelector}}))
	})
})
