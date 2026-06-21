package invite_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccInvite_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInviteConfig(guildID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("discord_invite.test", "code"),
					resource.TestCheckResourceAttrSet("discord_invite.test", "url"),
					resource.TestCheckResourceAttr("discord_invite.test", "max_age", "86400"),
					resource.TestCheckResourceAttr("discord_invite.test", "max_uses", "0"),
					resource.TestCheckResourceAttr("discord_invite.test", "temporary", "false"),
				),
			},
			{
				ResourceName:                         "discord_invite.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_invite.test"]
					return rs.Primary.Attributes["code"], nil
				},
				ImportStateVerifyIgnore: []string{"unique", "max_age", "max_uses", "temporary"},
			},
		},
	})
}

func testAccInviteConfig(guildID string) string {
	return fmt.Sprintf(`
resource "discord_channel" "inv" {
  guild_id = %q
  name     = "tf-acc-invite-ch"
  type     = 0
}
resource "discord_invite" "test" {
  channel_id = discord_channel.inv.id
  max_age    = 86400
  max_uses   = 0
  temporary  = false
  unique     = true
}`, guildID)
}

func TestAccInviteDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource "discord_channel" "inv_ds" {
  guild_id = %q
  name     = "tf-acc-invite-ds-ch"
  type     = 0
}
resource "discord_invite" "inv_ds" {
  channel_id = discord_channel.inv_ds.id
  max_age    = 3600
  unique     = true
}
data "discord_invite" "test" {
  code = discord_invite.inv_ds.code
}`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.discord_invite.test", "channel_id"),
				resource.TestCheckResourceAttrSet("data.discord_invite.test", "guild_id"),
			),
		}},
	})
}
