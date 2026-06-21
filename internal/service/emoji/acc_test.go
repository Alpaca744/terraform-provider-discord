package emoji_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// A minimal 1x1 transparent PNG encoded as a Discord image data URI.
const testPNGDataURI = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func TestAccEmoji_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEmojiConfig(guildID, "tfaccemoji"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_emoji.test", "name", "tfaccemoji"),
					resource.TestCheckResourceAttrSet("discord_guild_emoji.test", "id"),
				),
			},
			{
				Config: testAccEmojiConfig(guildID, "tfaccemojirenamed"),
				Check:  resource.TestCheckResourceAttr("discord_guild_emoji.test", "name", "tfaccemojirenamed"),
			},
		},
	})
}

func testAccEmojiConfig(guildID, name string) string {
	return fmt.Sprintf(`
resource "discord_guild_emoji" "test" {
  guild_id = %q
  name     = %q
  image    = %q
}`, guildID, name, testPNGDataURI)
}
