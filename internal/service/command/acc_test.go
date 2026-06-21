package command_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccApplicationCommand_basic(t *testing.T) {
	appID := acctest.TestApplicationID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalCommandConfig(appID, "tf-acc-cmd", "A test command"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_application_command.test", "name", "tf-acc-cmd"),
					resource.TestCheckResourceAttr("discord_application_command.test", "description", "A test command"),
					resource.TestCheckResourceAttrSet("discord_application_command.test", "id"),
				),
			},
			{
				Config: testAccGlobalCommandConfig(appID, "tf-acc-cmd", "Updated description"),
				Check:  resource.TestCheckResourceAttr("discord_application_command.test", "description", "Updated description"),
			},
			{
				ResourceName:      "discord_application_command.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_application_command.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["application_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccGlobalCommandConfig(appID, name, description string) string {
	return fmt.Sprintf(`
resource "discord_application_command" "test" {
  application_id = %q
  name           = %q
  description    = %q
  type           = 1
}`, appID, name, description)
}

func TestAccGuildApplicationCommand_basic(t *testing.T) {
	appID := acctest.TestApplicationID()
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuildCommandConfig(appID, guildID, "tf-acc-guild-cmd", "A guild command"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_application_command.test", "name", "tf-acc-guild-cmd"),
					resource.TestCheckResourceAttrSet("discord_guild_application_command.test", "id"),
				),
			},
			{
				Config: testAccGuildCommandConfig(appID, guildID, "tf-acc-guild-cmd", "Updated guild command"),
				Check:  resource.TestCheckResourceAttr("discord_guild_application_command.test", "description", "Updated guild command"),
			},
			{
				ResourceName:      "discord_guild_application_command.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_guild_application_command.test"]
					return fmt.Sprintf("%s:%s:%s",
						rs.Primary.Attributes["application_id"],
						rs.Primary.Attributes["guild_id"],
						rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccGuildCommandConfig(appID, guildID, name, description string) string {
	return fmt.Sprintf(`
resource "discord_guild_application_command" "test" {
  application_id = %q
  guild_id       = %q
  name           = %q
  description    = %q
  type           = 1
}`, appID, guildID, name, description)
}
