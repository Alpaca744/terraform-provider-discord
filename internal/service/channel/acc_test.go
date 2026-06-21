package channel_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ---- discord_channel ----

func TestAccChannel_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig(guildID, "tf-acc-channel", "Hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_channel.test", "name", "tf-acc-channel"),
					resource.TestCheckResourceAttr("discord_channel.test", "topic", "Hello"),
					resource.TestCheckResourceAttr("discord_channel.test", "type", "0"),
					resource.TestCheckResourceAttrSet("discord_channel.test", "id"),
				),
			},
			{
				Config: testAccChannelConfig(guildID, "tf-acc-channel-updated", "Updated topic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_channel.test", "name", "tf-acc-channel-updated"),
					resource.TestCheckResourceAttr("discord_channel.test", "topic", "Updated topic"),
				),
			},
			{
				ResourceName:      "discord_channel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccChannelConfig(guildID, name, topic string) string {
	return fmt.Sprintf(`
resource "discord_channel" "test" {
  guild_id = %q
  name     = %q
  type     = 0
  topic    = %q
}`, guildID, name, topic)
}

// ---- discord_channel data source ----

func TestAccChannelDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource "discord_channel" "test" {
  guild_id = %q
  name     = "tf-acc-channel-ds"
  type     = 0
}
data "discord_channel" "test" {
  id = discord_channel.test.id
}`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_channel.test", "name", "tf-acc-channel-ds"),
				resource.TestCheckResourceAttr("data.discord_channel.test", "type", "0"),
			),
		}},
	})
}

// ---- discord_channel_permission_overwrite ----

func TestAccChannelPermissionOverwrite(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOverwriteConfig(guildID, `["VIEW_CHANNEL"]`, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("discord_channel_permission_overwrite.test", "id"),
					resource.TestCheckResourceAttr("discord_channel_permission_overwrite.test", "type", "role"),
				),
			},
			{
				Config: testAccOverwriteConfig(guildID, `["VIEW_CHANNEL", "SEND_MESSAGES"]`, `[]`),
				Check:  resource.TestCheckResourceAttrSet("discord_channel_permission_overwrite.test", "id"),
			},
			{
				ResourceName:      "discord_channel_permission_overwrite.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_channel_permission_overwrite.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["channel_id"], rs.Primary.Attributes["overwrite_id"]), nil
				},
				ImportStateVerifyIgnore: []string{"id"},
			},
		},
	})
}

func testAccOverwriteConfig(guildID, allow, deny string) string {
	return fmt.Sprintf(`
resource "discord_channel" "ow" {
  guild_id = %q
  name     = "tf-acc-overwrite-ch"
  type     = 0
}
resource "discord_role" "ow" {
  guild_id = %q
  name     = "tf-acc-overwrite-role"
}
resource "discord_channel_permission_overwrite" "test" {
  channel_id = discord_channel.ow.id
  overwrite_id = discord_role.ow.id
  type         = "role"
  allow        = %s
  deny         = %s
}`, guildID, guildID, allow, deny)
}

// ---- discord_channels data source ----

func TestAccChannelsDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "discord_channels" "test" { guild_id = %q }`, guildID),
			Check:  resource.TestCheckResourceAttrSet("data.discord_channels.test", "channels.#"),
		}},
	})
}
