package integration

import (
	"fmt"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alphagov/terraform-provider-concourse/pkg/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/concourse/concourse/atc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var _ = Describe("Team management", func() {
	BeforeEach(SetupTest)
	AfterEach(TeardownTest)

	It("should manage the lifecycle of a team", func() {
		providers := map[string]*schema.Provider{
			"concourse": provider.Provider(),
		}

		client, err := NewConcourseClient()

		Expect(err).NotTo(HaveOccurred())

		resource.Test(NewGinkoTerraformTestingT(), resource.TestCase{
			IsUnitTest: false,

			Providers: providers,

			Steps: []resource.TestStep{
				resource.TestStep{
					// Add a user as an owner

					Config: `resource "concourse_team" "a_team" {
									   team_name = "team-a"
									   owners = ["user:github:tlwr"]
									}`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Adding a user as an owner")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_team.a_team", "team_name", "team-a"),

						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.#", "1"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.0", "user:github:tlwr"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "members.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "viewers.#", "0"),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("team-a"))

							expectedTeamAuth := atc.TeamAuth{
								"owner": {"users": {"github:tlwr"}},
							}

							Expect(teams[1].Auth).To(Equal(expectedTeamAuth))

							return nil
						},
					),
				},

				resource.TestStep{
					// Add another user as another owner

					Config: `resource "concourse_team" "a_team" {
									   team_name = "team-a"
									   owners = [
									     "user:github:tlwr",
									     "user:github:terraform-provider-concourse",
									   ]
									}`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Adding a user as an owner")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_team.a_team", "team_name", "team-a"),

						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.#", "2"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.0", "user:github:terraform-provider-concourse"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.1", "user:github:tlwr"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "members.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "viewers.#", "0"),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("team-a"))

							expectedTeamAuth := atc.TeamAuth{
								"owner": {"users": {
									"github:terraform-provider-concourse",
									"github:tlwr",
								}},
							}

							sort.Strings(teams[1].Auth["owner"]["users"])
							Expect(teams[1].Auth).To(Equal(expectedTeamAuth))

							return nil
						},
					),
				},

				resource.TestStep{
					// Change a user from an owner to a pipeline-operator

					Config: `resource "concourse_team" "a_team" {
									   team_name = "team-a"

										 owners = ["user:github:terraform-provider-concourse"]
										 pipeline_operators = ["user:github:tlwr"]
									}`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Changing a user from an owner to a pipeline-operator")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_team.a_team", "team_name", "team-a"),

						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.#", "1"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.0", "user:github:terraform-provider-concourse"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "members.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.#", "1"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.0", "user:github:tlwr"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "viewers.#", "0"),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("team-a"))

							expectedTeamAuth := atc.TeamAuth{
								"pipeline-operator": {
								  "users": {
								    "github:tlwr",
								  },
								},
								"owner": {
								  "users": {
								    "github:terraform-provider-concourse",
								  },
								},
							}

							Expect(teams[1].Auth).To(Equal(expectedTeamAuth))

							return nil
						},
					),
				},

				resource.TestStep{
					// Removing a user, adding a group

					Config: `resource "concourse_team" "a_team" {
									   team_name = "team-a"

										 owners = [
										   "user:github:terraform-provider-concourse",
										   "group:github:alphagov:paas-team",
										 ]
									}`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Removing a user, adding a group")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_team.a_team", "team_name", "team-a"),

						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.#", "2"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.0", "group:github:alphagov:paas-team"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.1", "user:github:terraform-provider-concourse"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "members.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "viewers.#", "0"),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("team-a"))

							expectedTeamAuth := atc.TeamAuth{
								"owner": {
								  "users": {"github:terraform-provider-concourse"},
								  "groups": {"github:alphagov:paas-team"},
								},
							}

							Expect(teams[1].Auth).To(Equal(expectedTeamAuth))

							return nil
						},
					),
				},

				resource.TestStep{
					// New team

					Config: `resource "concourse_team" "new_team" {
									   team_name = "team-new"

										 pipeline_operators = ["user:github:tlwr"]
									}`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("New team")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_team.new_team", "team_name", "team-new"),

						resource.TestCheckResourceAttr("concourse_team.new_team", "owners.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.new_team", "members.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.new_team", "pipeline_operators.#", "1"),
						resource.TestCheckResourceAttr("concourse_team.new_team", "pipeline_operators.0", "user:github:tlwr"),
						resource.TestCheckResourceAttr("concourse_team.new_team", "viewers.#", "0"),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("team-new"))

							expectedTeamAuth := atc.TeamAuth{
								"pipeline-operator": {"users": {"github:tlwr"}},
							}

							Expect(teams[1].Auth).To(Equal(expectedTeamAuth))

							return nil
						},
					),
				},
				resource.TestStep{
					// Rename the team

					Config: `resource "concourse_team" "a_team" {
									   team_name = "team-a-renamed"

										 pipeline_operators = ["user:github:tlwr"]
									}`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Renaming the team")

							fmt.Printf("%+v\n", s)
							return nil
						},

						resource.TestCheckResourceAttr("concourse_team.a_team", "team_name", "team-a-renamed"),

						resource.TestCheckResourceAttr("concourse_team.a_team", "owners.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "members.#", "0"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.#", "1"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "pipeline_operators.0", "user:github:tlwr"),
						resource.TestCheckResourceAttr("concourse_team.a_team", "viewers.#", "0"),

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(2))

							Expect(teams[0].Name).To(Equal("main"))
							Expect(teams[1].Name).To(Equal("team-a-renamed"))

							expectedTeamAuth := atc.TeamAuth{
								"pipeline-operator": {"users": {"github:tlwr"}},
							}

							Expect(teams[1].Auth).To(Equal(expectedTeamAuth))

							return nil
						},
					),
				},

				resource.TestStep{
					// Delete the team

					Config: `# Cannot be empty`,

					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							By("Deleting the team")

							fmt.Printf("%+v\n", s)
							return nil
						},

						func(s *terraform.State) error {
							Expect(s.RootModule().Resources).To(HaveLen(0))
							return nil
						},

						func(s *terraform.State) error {
							teams, err := client.ListTeams()

							if err != nil {
								return nil
							}

							Expect(teams).To(HaveLen(1))

							Expect(teams[0].Name).To(Equal("main"))

							return nil
						},
					),
				},
			},

			CheckDestroy: resource.ComposeTestCheckFunc(
				func(s *terraform.State) error {
					teams, err := client.ListTeams()

					if err != nil {
						return nil
					}

					Expect(teams).To(HaveLen(1))

					Expect(teams[0].Name).To(Equal("main"))

					return nil
				},
			),
		})
	})
})
