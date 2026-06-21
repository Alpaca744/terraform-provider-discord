package guild_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccRole_basic exercises create, update, and import of a discord_role
// against a live Discord guild. It is gated by TF_ACC and the credential
// environment variables checked in acctest.PreCheck.
func TestAccRole_basic(t *testing.T) {
	guildID := acctest.TestGuildID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig(guildID, "tf-acc-role", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_role.test", "name", "tf-acc-role"),
					resource.TestCheckResourceAttr("discord_role.test", "hoist", "true"),
					resource.TestCheckResourceAttrSet("discord_role.test", "id"),
					resource.TestCheckResourceAttrSet("discord_role.test", "position"),
				),
			},
			{
				Config: testAccRoleConfig(guildID, "tf-acc-role-renamed", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_role.test", "name", "tf-acc-role-renamed"),
					resource.TestCheckResourceAttr("discord_role.test", "hoist", "false"),
				),
			},
			{
				ResourceName:      "discord_role.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: roleImportID("discord_role.test"),
			},
		},
	})
}

// roleImportID builds the "guild_id:role_id" import string from resource state.
func roleImportID(name string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", name)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["guild_id"], rs.Primary.ID), nil
	}
}

// testAccRoleConfig renders a role configuration for the test guild.
func testAccRoleConfig(guildID, name string, hoist bool) string {
	return fmt.Sprintf(`
resource "discord_role" "test" {
  guild_id    = %[1]q
  name        = %[2]q
  hoist       = %[3]t
  mentionable = false

  permissions = [
    "VIEW_CHANNEL",
    "SEND_MESSAGES",
  ]
}
`, guildID, name, hoist)
}
