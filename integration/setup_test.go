package integration

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/concourse/concourse/go-concourse/concourse"

	"github.com/alphagov/terraform-provider-concourse/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	concourseURL      = "http://localhost:8080"
	concourseTeam     = "main"
	concourseUsername = "admin"
	concoursePassword = "password"
)

func NewConcourseClient() (concourse.Client, error) {
	c, err := client.NewConcourseClient(
		concourseURL,
		concourseTeam,
		concourseUsername, concoursePassword,
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func SetupSuite() {
	Expect(os.Setenv("TF_ACC", "true")).NotTo(HaveOccurred())

	Expect(os.Setenv("FLY_URL", concourseURL)).NotTo(HaveOccurred())
	Expect(os.Setenv("FLY_TEAM", concourseTeam)).NotTo(HaveOccurred())
	Expect(os.Setenv("FLY_USERNAME", concourseUsername)).NotTo(HaveOccurred())
	Expect(os.Setenv("FLY_PASSWORD", concoursePassword)).NotTo(HaveOccurred())

	buildCmd := exec.Command("docker-compose", "build")
	session, err := gexec.Start(buildCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, 300).Should(gexec.Exit(0))
}

func SetupTest() {
	upCmd := exec.Command("docker-compose", "up", "-d", "--force-recreate")
	session, err := gexec.Start(upCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, 120).Should(gexec.Exit(0))

	Eventually(func() error {
		fmt.Println("Waiting for Concourse to be ready")

		client, err := NewConcourseClient()

		if err != nil {
			return err
		}

		_, err = client.ListWorkers()

		if err != nil {
			return err
		}

		fmt.Println("Concourse is ready")
		return nil
	}, "15s", "3s").ShouldNot(HaveOccurred())
}

func TeardownTest() {
	downCmd := exec.Command("docker-compose", "down")
	session, err := gexec.Start(downCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, 60).Should(gexec.Exit(0))
}

func TeardownSuite() {
	gexec.KillAndWait()
}

type GinkgoTerraformTestingT struct {
	GinkgoTInterface

	CurrentDescription GinkgoTestDescription
}

func NewGinkoTerraformTestingT() GinkgoTerraformTestingT {
	return GinkgoTerraformTestingT{GinkgoT(), CurrentGinkgoTestDescription()}
}

func (t GinkgoTerraformTestingT) Helper() {}

func (t GinkgoTerraformTestingT) Name() string {
	return t.CurrentDescription.FullTestText
}

func (t GinkgoTerraformTestingT) Verbose() bool {
	return true
}
