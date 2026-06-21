package sticker_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Minimal 1x1 transparent PNG (same pixel used by emoji tests) encoded as raw
// base64 (no data-URI prefix) for sticker upload.
const testStickerPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func TestAccGuildSticker_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStickerConfig(guildID, "tfaccsticker", "a test sticker", "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_sticker.test", "name", "tfaccsticker"),
					resource.TestCheckResourceAttr("discord_guild_sticker.test", "description", "a test sticker"),
					resource.TestCheckResourceAttr("discord_guild_sticker.test", "tags", "test"),
					resource.TestCheckResourceAttrSet("discord_guild_sticker.test", "id"),
				),
			},
			{
				Config: testAccStickerConfig(guildID, "tfaccstickerrenamed", "updated description", "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_sticker.test", "name", "tfaccstickerrenamed"),
					resource.TestCheckResourceAttr("discord_guild_sticker.test", "description", "updated description"),
					resource.TestCheckResourceAttr("discord_guild_sticker.test", "tags", "test2"),
				),
			},
			{
				ResourceName:      "discord_guild_sticker.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_guild_sticker.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["guild_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"file_content_base64", "format_type", "format"},
			},
		},
	})
}

func testAccStickerConfig(guildID, name, description, tags string) string {
	return fmt.Sprintf(`
resource "discord_guild_sticker" "test" {
  guild_id            = %q
  name                = %q
  description         = %q
  tags                = %q
  file_content_base64 = %q
  format              = "png"
}`, guildID, name, description, tags, testStickerPNGBase64)
}
