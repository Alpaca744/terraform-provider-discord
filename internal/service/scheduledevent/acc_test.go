package scheduledevent_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScheduledEvent_basic(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledEventConfig(guildID, "tf-acc-event", "Gaming night"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_scheduled_event.test", "name", "tf-acc-event"),
					resource.TestCheckResourceAttr("discord_guild_scheduled_event.test", "description", "Gaming night"),
					resource.TestCheckResourceAttrSet("discord_guild_scheduled_event.test", "id"),
				),
			},
			{
				Config: testAccScheduledEventConfig(guildID, "tf-acc-event", "Updated description"),
				Check:  resource.TestCheckResourceAttr("discord_guild_scheduled_event.test", "description", "Updated description"),
			},
			{
				ResourceName:      "discord_guild_scheduled_event.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_guild_scheduled_event.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["guild_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccScheduledEventConfig(guildID, name, description string) string {
	return fmt.Sprintf(`
resource "discord_guild_scheduled_event" "test" {
  guild_id             = %q
  name                 = %q
  description          = %q
  privacy_level        = 2
  entity_type          = 3
  location             = "Online"
  scheduled_start_time = "2026-12-01T18:00:00Z"
  scheduled_end_time   = "2026-12-01T20:00:00Z"
}`, guildID, name, description)
}
