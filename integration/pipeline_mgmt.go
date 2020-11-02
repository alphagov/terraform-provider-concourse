package integration

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alphagov/terraform-provider-concourse/pkg/provider"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	// "github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var _ = Describe("Pipeline management", func() {
	BeforeEach(SetupTest)
	AfterEach(TeardownTest)

	const (
		pipelineConfig = `#
resources:
  - name: every-midnight
    type: time
    source:
      location: Europe/London
      start: 12:00AM
      stop: 12:15AM

jobs:
  - name: check-the-time
    serial: true
    plan:
    - get: every-midnight
      trigger: true
`

		pipelineConfigJSON = `{"jobs":[{"name":"check-the-time","plan":[{"get":"every-midnight","trigger":true}],"serial":true}],"resources":[{"name":"every-midnight","source":{"location":"Europe/London","start":"12:00AM","stop":"12:15AM"},"type":"time"}]}`

		updatedPipelineConfig = `#
resources:
  - name: every-midnight
    type: time
    source:
      location: Europe/London
      start: 12:00AM
      stop: 12:15AM

jobs:
  - name: check-the-time-on-demand
    serial: true
    plan:
    - get: every-midnight
`
		updatedPipelineConfigJSON = `{"jobs":[{"name":"check-the-time-on-demand","plan":[{"get":"every-midnight"}],"serial":true}],"resources":[{"name":"every-midnight","source":{"location":"Europe/London","start":"12:00AM","stop":"12:15AM"},"type":"time"}]}`
	)

	It("should manage the lifecycle of a pipeline", func() {
		providers := map[string]terraform.ResourceProvider{
			"concourse": provider.Provider(),
		}

		client, err := NewConcourseClient()

		Expect(err).NotTo(HaveOccurred())

		resource.Test(NewGinkoTerraformTestingT(), resource.TestCase{
			IsUnitTest: false,

			Providers: providers,

			Steps: []resource.TestStep{
				resource.TestStep{
					// Add a pipeline

					Config: fmt.Sprintf(`data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }

                   resource "concourse_pipeline" "a_pipeline" {
                      team_name     = "${data.concourse_team.main_team.team_name}"
                      pipeline_name = "pipeline-a"

                      is_exposed = false
                      is_paused  = false

                      pipeline_config_format = "yaml"
                      pipeline_config        = <<PIPELINE
%s
                      PIPELINE
                   }`, pipelineConfig),

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Adding a pipeline")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "pipeline_name", "pipeline-a"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "team_name", "main"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_exposed", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_paused", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "json", pipelineConfigJSON),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(1))

							Expect(pipelines[0].Name).To(Equal("pipeline-a"))
							Expect(pipelines[0].TeamName).To(Equal("main"))
							Expect(pipelines[0].Paused).To(Equal(false))
							Expect(pipelines[0].Public).To(Equal(false))

							return nil
						},
					),
				},

				resource.TestStep{
					// Pause and expose the pipeline

					Config: fmt.Sprintf(`data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }

                   resource "concourse_pipeline" "a_pipeline" {
                      team_name     = "${data.concourse_team.main_team.team_name}"
                      pipeline_name = "pipeline-a"

                      is_exposed = true
                      is_paused  = true

                      pipeline_config_format = "yaml"
                      pipeline_config        = <<PIPELINE
%s
                      PIPELINE
                   }`, pipelineConfig),

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Pausing and exposing the pipeline")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "pipeline_name", "pipeline-a"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "team_name", "main"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_exposed", "true"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_paused", "true"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "json", pipelineConfigJSON),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(1))

							Expect(pipelines[0].Name).To(Equal("pipeline-a"))
							Expect(pipelines[0].TeamName).To(Equal("main"))
							Expect(pipelines[0].Paused).To(Equal(true))
							Expect(pipelines[0].Public).To(Equal(true))

							return nil
						},
					),
				},

				resource.TestStep{
					// Unpause and hide the pipeline

					Config: fmt.Sprintf(`data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }

                   resource "concourse_pipeline" "a_pipeline" {
                      team_name     = "${data.concourse_team.main_team.team_name}"
                      pipeline_name = "pipeline-a"

                      is_exposed = false
                      is_paused  = false

                      pipeline_config_format = "yaml"
                      pipeline_config        = <<PIPELINE
%s
                      PIPELINE
                   }`, pipelineConfig),

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Unpausing and hiding the pipeline")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "pipeline_name", "pipeline-a"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "team_name", "main"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_exposed", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_paused", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "json", pipelineConfigJSON),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(1))

							Expect(pipelines[0].Name).To(Equal("pipeline-a"))
							Expect(pipelines[0].TeamName).To(Equal("main"))
							Expect(pipelines[0].Paused).To(Equal(false))
							Expect(pipelines[0].Public).To(Equal(false))

							return nil
						},
					),
				},

				resource.TestStep{
					// Update the pipeline configuration

					Config: fmt.Sprintf(`data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }

                   resource "concourse_pipeline" "a_pipeline" {
                      team_name     = "${data.concourse_team.main_team.team_name}"
                      pipeline_name = "pipeline-a"

                      is_exposed = false
                      is_paused  = false

                      pipeline_config_format = "yaml"
                      pipeline_config        = <<PIPELINE
%s
                      PIPELINE
                   }`, updatedPipelineConfig),

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Updating the pipeline configuration")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "pipeline_name", "pipeline-a"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "team_name", "main"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_exposed", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_paused", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "json", updatedPipelineConfigJSON),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(1))

							Expect(pipelines[0].Name).To(Equal("pipeline-a"))
							Expect(pipelines[0].TeamName).To(Equal("main"))
							Expect(pipelines[0].Paused).To(Equal(false))
							Expect(pipelines[0].Public).To(Equal(false))

							return nil
						},
					),
				},

				resource.TestStep{
					// Move a pipeline from one team to another

					Config: fmt.Sprintf(`data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }

                   resource "concourse_pipeline" "a_pipeline" {
                      team_name     = "${concourse_team.other_team.team_name}"
                      pipeline_name = "pipeline-a"

                      is_exposed = false
                      is_paused  = false

                      pipeline_config_format = "yaml"
                      pipeline_config        = <<PIPELINE
%s
                      PIPELINE
                   }`, updatedPipelineConfig),

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Moving the pipeline from one team to another")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "pipeline_name", "pipeline-a"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "team_name", "other"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_exposed", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_paused", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "json", updatedPipelineConfigJSON),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(1))

							Expect(pipelines[0].Name).To(Equal("pipeline-a"))
							Expect(pipelines[0].TeamName).To(Equal("other"))
							Expect(pipelines[0].Paused).To(Equal(false))
							Expect(pipelines[0].Public).To(Equal(false))

							return nil
						},
					),
				},

				resource.TestStep{
					// Rename the pipeline

					Config: fmt.Sprintf(`data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }

                   resource "concourse_pipeline" "a_pipeline" {
                      team_name     = "${concourse_team.other_team.team_name}"
                      pipeline_name = "pipeline-a-renamed"

                      is_exposed = false
                      is_paused  = false

                      pipeline_config_format = "yaml"
                      pipeline_config        = <<PIPELINE
%s
                      PIPELINE
                   }`, updatedPipelineConfig),

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Renaming the pipeline")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "pipeline_name", "pipeline-a-renamed"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "team_name", "other"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_exposed", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "is_paused", "false"),
						resource.TestCheckResourceAttr("concourse_pipeline.a_pipeline", "json", updatedPipelineConfigJSON),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(1))

							Expect(pipelines[0].Name).To(Equal("pipeline-a-renamed"))
							Expect(pipelines[0].TeamName).To(Equal("other"))
							Expect(pipelines[0].Paused).To(Equal(false))
							Expect(pipelines[0].Public).To(Equal(false))

							return nil
						},
					),
				},

				resource.TestStep{
					// Delete the pipeline

					Config: `data "concourse_team" "main_team" {
					 team_name = "main"
                   }

                   resource "concourse_team" "other_team" {
					 team_name = "other"
					 owners = ["user:github:tlwr"]
                   }`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Deleting the pipeline")

							fmt.Printf("%+v\n", s)
							return nil
						},

						func(s *terraform.State) error {
							Expect(s.RootModule().Resources).To(HaveLen(2))

							Expect(
								s.RootModule().Resources["data.concourse_team.main_team"],
							).NotTo(BeNil())

							Expect(
								s.RootModule().Resources["concourse_team.other_team"],
							).NotTo(BeNil())

							return nil
						},

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return err
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("other"))

							pipelines, err := client.ListPipelines()
							Expect(err).NotTo(HaveOccurred())
							Expect(pipelines).To(HaveLen(0))

							return nil
						},
					),
				},
			},

			CheckDestroy: resource.ComposeTestCheckFunc(
				func(s *terraform.State) error {
					teams, err := client.ListTeams()

					if err != nil {
						return err
					}

					Expect(teams).To(HaveLen(1))

					Expect(teams[0].Name).To(Equal("main"))

					pipelines, err := client.ListPipelines()
					Expect(err).NotTo(HaveOccurred())
					Expect(pipelines).To(HaveLen(0))

					return nil
				},
			),
		})
	})
})
