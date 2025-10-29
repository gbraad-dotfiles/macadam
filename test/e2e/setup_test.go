package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/crc-org/macadam/test/osprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var tempDir string
var machineResponses []ListReporter
var err error
var image string
var keypath string
var cloudinitPath string

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

var _ = Describe("Macadam init setup test", Label("init"), func() {
	BeforeEach(func() {
		session := macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err := json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(0))
	})

	AfterEach(func() {
		report := CurrentSpecReport()
		labels := report.Labels()
		stopCmd := "stop"
		rmCmd := "rm -f"
		for _, l := range labels {
			if l == "name" {
				stopCmd = "stop myVM"
				rmCmd = "rm myVM -f"
				break
			}
		}

		// stop the CentOS VM
		session := macadamTest.Macadam(strings.Fields(stopCmd))
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("stopped successfully"))

		// rm the CentOS VM and verify that "list" does not return any vm
		session = macadamTest.Macadam(strings.Fields(rmCmd))
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(0))
	})

	It("init CentOS VM with cpu, disk and memory setup", Label("cpu"), func() {
		// init a CentOS VM with cpu and disk-size setup
		session := macadamTest.Macadam([]string{"init", "--cpus", "3", "--disk-size", "30", "--memory", "2048", image})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		// check the list command returns one item
		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(1))

		// start the CentOS VM
		session = macadamTest.Macadam([]string{"start"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("started successfully"))

		// ssh into the VM and prints user
		session = macadamTest.Macadam([]string{"ssh", "nproc"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(strings.TrimSpace(session.OutputToString())).Should(Equal("3"))

		session = macadamTest.Macadam([]string{"ssh", "lsblk"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("30G"))

		session = macadamTest.Macadam([]string{"ssh", "free", "-h"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		// Verify memory is close to 2048MB (allow for system overhead)
		output := session.OutputToString()
		Expect(output).Should(Or(ContainSubstring("1.7G"), ContainSubstring("1.8G"), ContainSubstring("1.9G"), ContainSubstring("2.0G")))
	})

	It("init CentOS VM with username and sshkey setup", Label("sshkey"), func() {
		// init a CentOS VM with cpu and disk-size setup
		session := macadamTest.Macadam([]string{"init", "--username", "test", "--ssh-identity-path", keypath, image})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		// check the list command returns one item
		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(1))

		// start the CentOS VM
		session = macadamTest.Macadam([]string{"start"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("started successfully"))

		// ssh into the VM and prints user
		session = macadamTest.Macadam([]string{"ssh", "--username", "test", "whoami"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(strings.TrimSpace(session.OutputToString())).Should(Equal("test"))
	})

	It("init CentOS VM with cloud-init setup", Label("cloudinit"), func() {
		// init a CentOS VM with cpu and disk-size setup
		session := macadamTest.Macadam([]string{"init", "--cloud-init", cloudinitPath, "--username", "macadamtest", "--ssh-identity-path", keypath, image})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		// check the list command returns one item
		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(1))

		// start the CentOS VM
		session = macadamTest.Macadam([]string{"start"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("started successfully"))

		// ssh into the VM and prints user
		session = macadamTest.Macadam([]string{"ssh", "whoami"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(strings.TrimSpace(session.OutputToString())).Should(Equal("macadamtest"))

		// wait until cloud-init finish
		Eventually(func() string {
			session = macadamTest.Macadam([]string{"ssh", "systemctl", "status", "cloud-final"})
			session.WaitWithDefaultTimeout()

			if session.ExitCode() != 0 {
				return ""
			}
			return session.OutputToString()
		}, 10*time.Minute, 30*time.Second).Should(ContainSubstring("Active: active (exited)"))

		fmt.Println("cloud-init has finished")

		// ssh into the VM and check installed app
		session = macadamTest.Macadam([]string{"ssh", "git", "--version"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("git version"))

		// ssh into the VM and check file created
		session = macadamTest.Macadam([]string{"ssh", "ls", "/home/macadamtest/hello.txt"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
	})

	It("init CentOS VM with name", Label("name"), func() {
		// init a CentOS VM with name setup
		session := macadamTest.Macadam([]string{"init", "--name", "myVM", image})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))

		// check the list command returns one item
		session = macadamTest.Macadam([]string{"list", "--format", "json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		err = json.Unmarshal(session.Out.Contents(), &machineResponses)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(machineResponses)).Should(Equal(1))

		// start the CentOS VM with set name
		session = macadamTest.Macadam([]string{"start", "myVM"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(session.OutputToString()).Should(ContainSubstring("started successfully"))

		// ssh into the VM and prints user
		session = macadamTest.Macadam([]string{"ssh", "myVM", "whoami"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(gexec.Exit(0))
		Expect(strings.TrimSpace(session.OutputToString())).Should(Equal("core"))
	})
})
