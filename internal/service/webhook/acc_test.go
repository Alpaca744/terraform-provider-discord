package webhook_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWebhook_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig(guildID, "tf-acc-webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_webhook.test", "name", "tf-acc-webhook"),
					resource.TestCheckResourceAttrSet("discord_webhook.test", "id"),
					resource.TestCheckResourceAttrSet("discord_webhook.test", "url"),
				),
			},
			{
				Config: testAccWebhookConfig(guildID, "tf-acc-webhook-renamed"),
				Check:  resource.TestCheckResourceAttr("discord_webhook.test", "name", "tf-acc-webhook-renamed"),
			},
			{
				ResourceName:            "discord_webhook.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "url"},
			},
		},
	})
}

func testAccWebhookConfig(guildID, name string) string {
	return fmt.Sprintf(`
resource "discord_channel" "wh" {
  guild_id = %q
  name     = "tf-acc-webhook-ch"
  type     = 0
}
resource "discord_webhook" "test" {
  channel_id = discord_channel.wh.id
  name       = %q
}`, guildID, name)
}

func TestAccWebhookDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource "discord_channel" "wh_ds" {
  guild_id = %q
  name     = "tf-acc-wh-ds-ch"
  type     = 0
}
resource "discord_webhook" "wh_ds" {
  channel_id = discord_channel.wh_ds.id
  name       = "tf-acc-wh-ds"
}
data "discord_webhook" "test" {
  id = discord_webhook.wh_ds.id
}`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_webhook.test", "name", "tf-acc-wh-ds"),
				resource.TestCheckResourceAttrSet("data.discord_webhook.test", "channel_id"),
			),
		}},
	})
}
