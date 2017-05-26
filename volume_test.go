package kubernetes_test

import (
	. "github.com/Twister915/kube_builders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volume", func() {
	const (
		name     = "test"
		hostPath = "/var/run/docker.sock"
	)

	It("can be created", func() {
		vol := NewVolumeBuilder(name).HostPath(hostPath).AsKube()
		Expect(vol.HostPath).ToNot(BeNil())
		Expect(vol.Name).To(Equal(name))
		Expect(vol.HostPath.Path).To(Equal(hostPath))
	})
})
