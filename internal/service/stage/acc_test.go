package stage_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccStageInstance_basic requires a Community guild (stage channels, type=13,
// are only allowed in Community guilds). Set DISCORD_TEST_COMMUNITY_GUILD_ID to run.
func TestAccStageInstance_basic(t *testing.T) {
	guildID := acctest.TestCommunityGuildID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig(guildID, "AMA Session"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_stage_instance.test", "topic", "AMA Session"),
					resource.TestCheckResourceAttrSet("discord_stage_instance.test", "id"),
				),
			},
			{
				Config: testAccStageConfig(guildID, "AMA Session Updated"),
				Check:  resource.TestCheckResourceAttr("discord_stage_instance.test", "topic", "AMA Session Updated"),
			},
		},
	})
}

func testAccStageConfig(guildID, topic string) string {
	return fmt.Sprintf(`
resource "discord_channel" "stage" {
  guild_id = %q
  name     = "tf-acc-stage"
  type     = 13
}
resource "discord_stage_instance" "test" {
  channel_id    = discord_channel.stage.id
  topic         = %q
  privacy_level = 2
}`, guildID, topic)
}
