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
	macadamTest        *MacadamTestIntegration
	machineResponses   []ListReporter
	err                error
	tempDir            string
	CENTOS_QCOW2_IMAGE string
	keypath            string
	cloudinitPath      string
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Macadam suite")
}

func noVMcheck(opts ...string) {
	listCMD := []string{"list", "--format", "json"}
	if len(opts) > 0 {
		listCMD = append(listCMD, "--provider", opts[0])
	}

	session := macadamTest.Macadam(listCMD)
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	err := json.Unmarshal(session.Out.Contents(), &machineResponses)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(machineResponses)).Should(Equal(0))
}

func vmNumberCheck(expect_number int, opts ...string) {
	listCMD := []string{"list", "--format", "json"}
	if len(opts) > 0 {
		listCMD = append(listCMD, "--provider", opts[0])
	}

	session := macadamTest.Macadam(listCMD)
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	err := json.Unmarshal(session.Out.Contents(), &machineResponses)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(machineResponses)).Should(Equal(expect_number))
}

func runCMDsuccess(cmd []string) string {
	GinkgoWriter.Printf("Run cmd: %v\n", cmd)
	session := macadamTest.Macadam(cmd)
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	return session.OutputToString()
}

func runCMD(cmd []string) (int, string) {
	GinkgoWriter.Printf("Run cmd: %v\n", cmd)
	session := macadamTest.Macadam(cmd)
	session.WaitWithDefaultTimeout()
	return session.ExitCode(), session.OutputToString()
}

func removeAllVM(opts ...string) {
	listCMD := []string{"list", "--format", "json"}
	baseCMD := []string{}
	if len(opts) > 0 {
		provider := opts[0]
		listCMD = append(listCMD, "--provider", provider)
		baseCMD = append(baseCMD, "--provider", provider)
	}

	session := macadamTest.Macadam(listCMD)
	session.WaitWithDefaultTimeout()
	Expect(session).Should(gexec.Exit(0))
	err = json.Unmarshal(session.Out.Contents(), &machineResponses)
	Expect(err).NotTo(HaveOccurred())

	for _, m := range machineResponses {
		stopCMD := append([]string{"stop"}, baseCMD...)
		stopCMD = append(stopCMD, m.Name)
		rmCMD := append([]string{"rm", "-f"}, baseCMD...)
		rmCMD = append(rmCMD, m.Name)

		// stop the CentOS VM
		session = macadamTest.Macadam(stopCMD)
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("stopped successfully"))

		// rm the CentOS VM and verify that "list" does not return any vm
		session = macadamTest.Macadam(rmCMD)
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
	}

	session = macadamTest.Macadam(listCMD)
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
	CENTOS_QCOW2_IMAGE, err = centosProvider.Fetch(tempDir)
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
