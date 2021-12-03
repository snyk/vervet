package storage

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Aws S3 Client Initialization", func() {
	ginkgo.Context("given valid config", func() {
		ginkgo.When("using proper permissioned AWS credentials", func() {
			ginkgo.It("creates a writable S3 client", func() {
				client := getS3Client()
				gomega.Expect(client).ToNot(gomega.BeNil())
			})
		})
	})
})