package automod_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAutoModerationRule_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAutomodConfig(guildID, "tf-acc-automod", "badword"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_auto_moderation_rule.test", "name", "tf-acc-automod"),
					resource.TestCheckResourceAttr("discord_auto_moderation_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("discord_auto_moderation_rule.test", "id"),
				),
			},
			{
				Config: testAccAutomodConfig(guildID, "tf-acc-automod-updated", "badword2"),
				Check:  resource.TestCheckResourceAttr("discord_auto_moderation_rule.test", "name", "tf-acc-automod-updated"),
			},
			{
				ResourceName:      "discord_auto_moderation_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_auto_moderation_rule.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["guild_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccAutomodConfig(guildID, name, keyword string) string {
	return fmt.Sprintf(`
resource "discord_auto_moderation_rule" "test" {
  guild_id     = %q
  name         = %q
  event_type   = 1
  trigger_type = 1
  enabled      = true

  trigger_metadata = {
    keyword_filter = [%q]
  }

  actions = [
    { type = 1 }
  ]
}`, guildID, name, keyword)
}
