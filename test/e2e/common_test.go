package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crc-org/macadam/test/osprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	macadamTest *MacadamTestIntegration
	machineResponses []ListReporter
	err error
	tempDir string
	image string
	keypath string
	cloudinitPath string
)



func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Macadam suite")
}

func noVMcheck() {
	session := macadamTest.Macadam([]string{"list", "--format", "json"})
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	err := json.Unmarshal(session.Out.Contents(), &machineResponses)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(machineResponses)).Should(Equal(0))
}

func removeAllVM() {
	session := macadamTest.Macadam([]string{"list", "--format", "json"})
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	err = json.Unmarshal(session.Out.Contents(), &machineResponses)
	Expect(err).NotTo(HaveOccurred())

	for _, m := range machineResponses {
		stopCmd := "stop " + string(m.Name)
		rmCmd := "rm -f " + string(m.Name)

		// stop the CentOS VM
		session = macadamTest.Macadam(strings.Fields(stopCmd))
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("stopped successfully"))

		// rm the CentOS VM and verify that "list" does not return any vm
		session = macadamTest.Macadam(strings.Fields(rmCmd))
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
	}

	session = macadamTest.Macadam([]string{"list", "--format", "json"})
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	err = json.Unmarshal(session.Out.Contents(), &machineResponses)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(machineResponses)).Should(Equal(0))
}

var _ = BeforeSuite(func() {
	tempDir, err = os.MkdirTemp("", "test-")
	Expect(err).NotTo(HaveOccurred())

	// download CentOS image
	centosProvider := osprovider.NewCentosProvider()
	image, err = centosProvider.Fetch(tempDir)
	Expect(err).NotTo(HaveOccurred())

	keypath = filepath.Join(tempDir, "id_rsa")
	cloudinitPath = filepath.Join(tempDir, "user-data")
	//generate ssh key
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-f", keypath, "-N", "")
	err = cmd.Run()
	Expect(err).ShouldNot(HaveOccurred())
	//copy user-data
	wd, err := os.Getwd()
	Expect(err).ShouldNot(HaveOccurred())
	cloudinit := wd + "/../testdata/user-data"
	content, err := os.ReadFile(cloudinit)
	Expect(err).ShouldNot(HaveOccurred())
	//replace ssh pub key to user-data file
	pubkeypath := keypath + ".pub"
	pubkey, err := os.ReadFile(pubkeypath)
	Expect(err).ShouldNot(HaveOccurred())
	newContent := strings.ReplaceAll(string(content), "[sshkey]", string(pubkey))
	err = os.WriteFile(cloudinitPath, []byte(newContent), 0644)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	os.RemoveAll(tempDir)
})

var _ = BeforeEach(func() {
		macadamTest = MacadamTestCreate()
	})