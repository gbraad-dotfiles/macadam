package e2e

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Macadam init setup test", Label("multiple"), func() {
	BeforeEach(func() {
		noVMcheck()
	})

	AfterEach(func() {
		removeAllVM()
	})

	It("create multiple CentOS VM", Label("mul-centos"), func() {
		// init a CentOS VM with name vm1
		session := macadamTest.Macadam([]string{"init", "--name", "vm1", image})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		// check the list command returns one item
		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(1))

		// start the CentOS VM1
		session = macadamTest.Macadam([]string{"start", "vm1"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("started successfully"))

		// ssh into the VM1 and prints user
		session = macadamTest.Macadam([]string{"ssh", "vm1", "whoami"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(Equal("core"))

		//init another centos VM with name vm2
		session = macadamTest.Macadam([]string{"init", "--name", "vm2", image})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		// check the list command returns two items
		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(2))

		// start the CentOS VM2
		session = macadamTest.Macadam([]string{"start", "vm2"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("started successfully"))

		// ssh into the VM2 and prints user
		session = macadamTest.Macadam([]string{"ssh", "vm2", "whoami"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(Equal("core"))

		//check again VM1 still running
		session = macadamTest.Macadam([]string{"ssh", "vm1", "whoami"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(Equal("core"))
	})

})
