package soundboard_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccSoundboardSound_basic tests soundboard sound lifecycle.
// Requires DISCORD_TEST_SOUND_DATA_URI to be set to a valid MP3 or OGG data URI
// (e.g. "data:audio/mpeg;base64,<base64-encoded-audio>").
// Discord validates audio content server-side, so a real audio file is required.
func TestAccSoundboardSound_basic(t *testing.T) {
	soundDataURI := os.Getenv("DISCORD_TEST_SOUND_DATA_URI")
	if soundDataURI == "" {
		t.Skip("DISCORD_TEST_SOUND_DATA_URI not set — skipping soundboard test (requires real MP3/OGG data URI)")
	}
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSoundboardConfig(guildID, "tfaccsound", 1.0, soundDataURI),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_soundboard_sound.test", "name", "tfaccsound"),
					resource.TestCheckResourceAttr("discord_soundboard_sound.test", "volume", "1"),
					resource.TestCheckResourceAttrSet("discord_soundboard_sound.test", "sound_id"),
				),
			},
			{
				Config: testAccSoundboardConfig(guildID, "tfaccsoundrenamed", 0.5, soundDataURI),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_soundboard_sound.test", "name", "tfaccsoundrenamed"),
					resource.TestCheckResourceAttr("discord_soundboard_sound.test", "volume", "0.5"),
				),
			},
			{
				ResourceName:      "discord_soundboard_sound.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_soundboard_sound.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["guild_id"], rs.Primary.Attributes["sound_id"]), nil
				},
				ImportStateVerifyIgnore: []string{"sound"},
			},
		},
	})
}

func testAccSoundboardConfig(guildID, name string, volume float64, soundDataURI string) string {
	return fmt.Sprintf(`
resource "discord_soundboard_sound" "test" {
  guild_id = %q
  name     = %q
  sound    = %q
  volume   = %g
}`, guildID, name, soundDataURI, volume)
}
