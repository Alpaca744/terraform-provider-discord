package guild_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ---- discord_guild (data source) ----

func TestAccGuildDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "discord_guild" "test" { id = %q }`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_guild.test", "id", guildID),
				resource.TestCheckResourceAttrSet("data.discord_guild.test", "name"),
			),
		}},
	})
}

// ---- discord_guild_settings ----

func TestAccGuildSettings(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "discord_guild_settings" "test" {
  guild_id                    = %q
  default_message_notifications = 0
  explicit_content_filter      = 0
}`, guildID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_settings.test", "guild_id", guildID),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "discord_guild_settings" "test" {
  guild_id                    = %q
  default_message_notifications = 1
  explicit_content_filter      = 0
}`, guildID),
				Check: resource.TestCheckResourceAttr("discord_guild_settings.test", "default_message_notifications", "1"),
			},
		},
	})
}

// ---- discord_guild_widget ----

func TestAccGuildWidget(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "discord_guild_widget" "test" {
  guild_id = %q
  enabled  = true
}`, guildID),
				Check: resource.TestCheckResourceAttr("discord_guild_widget.test", "enabled", "true"),
			},
			{
				Config: fmt.Sprintf(`
resource "discord_guild_widget" "test" {
  guild_id = %q
  enabled  = false
}`, guildID),
				Check: resource.TestCheckResourceAttr("discord_guild_widget.test", "enabled", "false"),
			},
		},
	})
}

// ---- discord_member_role ----

func TestAccMemberRole(t *testing.T) {
	guildID := acctest.TestGuildID()
	// Uses the bot itself as the "member" — bot is always in the guild.
	botID := acctest.TestApplicationID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource "discord_role" "acc_member" {
  guild_id = %q
  name     = "tf-acc-member-role"
}
resource "discord_member_role" "test" {
  guild_id = %q
  user_id  = %q
  role_id  = discord_role.acc_member.id
}`, guildID, guildID, botID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("discord_member_role.test", "guild_id", guildID),
				resource.TestCheckResourceAttr("discord_member_role.test", "user_id", botID),
			),
		}},
	})
}

// ---- discord_guild_template ----

func TestAccGuildTemplate(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "discord_guild_template" "test" {
  guild_id    = %q
  name        = "tf-acc-template"
  description = "acceptance test template"
}`, guildID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_template.test", "name", "tf-acc-template"),
					resource.TestCheckResourceAttrSet("discord_guild_template.test", "code"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "discord_guild_template" "test" {
  guild_id    = %q
  name        = "tf-acc-template-renamed"
  description = "updated"
}`, guildID),
				Check: resource.TestCheckResourceAttr("discord_guild_template.test", "name", "tf-acc-template-renamed"),
			},
			{
				ResourceName:                         "discord_guild_template.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_guild_template.test"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["guild_id"], rs.Primary.Attributes["code"]), nil
				},
				ImportStateVerifyIgnore: []string{"description"},
			},
		},
	})
}

// ---- discord_guilds data source ----

func TestAccGuildsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "discord_guilds" "test" {}`,
			Check:  resource.TestCheckResourceAttrSet("data.discord_guilds.test", "guilds.#"),
		}},
	})
}

// ---- discord_guild_welcome_screen ----
// Requires a Community guild. Set DISCORD_TEST_COMMUNITY_GUILD_ID to run this test.

func TestAccGuildWelcomeScreen(t *testing.T) {
	guildID := acctest.TestCommunityGuildID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWelcomeScreenConfig(guildID, "Welcome to the server!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_welcome_screen.test", "guild_id", guildID),
					resource.TestCheckResourceAttr("discord_guild_welcome_screen.test", "description", "Welcome to the server!"),
				),
			},
			{
				Config: testAccWelcomeScreenConfig(guildID, "Updated welcome!", false),
				Check:  resource.TestCheckResourceAttr("discord_guild_welcome_screen.test", "description", "Updated welcome!"),
			},
		},
	})
}

func testAccWelcomeScreenConfig(guildID, description string, enabled bool) string {
	return fmt.Sprintf(`
resource "discord_guild_welcome_screen" "test" {
  guild_id    = %q
  description = %q
  enabled     = %t
}`, guildID, description, enabled)
}

// ---- discord_guild_onboarding ----

func TestAccGuildOnboarding(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOnboardingConfig(guildID, false, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_guild_onboarding.test", "guild_id", guildID),
					resource.TestCheckResourceAttr("discord_guild_onboarding.test", "enabled", "false"),
				),
			},
			{
				Config: testAccOnboardingConfig(guildID, false, 1),
				Check:  resource.TestCheckResourceAttr("discord_guild_onboarding.test", "mode", "1"),
			},
		},
	})
}

func testAccOnboardingConfig(guildID string, enabled bool, mode int) string {
	return fmt.Sprintf(`
resource "discord_guild_onboarding" "test" {
  guild_id = %q
  enabled  = %t
  mode     = %d
}`, guildID, enabled, mode)
}

// ---- discord_roles data source ----

func TestAccRolesDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "discord_roles" "test" { guild_id = %q }`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_roles.test", "guild_id", guildID),
				resource.TestCheckResourceAttrSet("data.discord_roles.test", "roles.#"),
			),
		}},
	})
}

// ---- discord_guild_preview data source ----

func TestAccGuildPreviewDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "discord_guild_preview" "test" { id = %q }`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_guild_preview.test", "id", guildID),
				resource.TestCheckResourceAttrSet("data.discord_guild_preview.test", "name"),
			),
		}},
	})
}

// ---- discord_audit_log data source ----

func TestAccAuditLogDataSource(t *testing.T) {
	guildID := acctest.TestGuildID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
data "discord_audit_log" "test" {
  guild_id = %q
  limit    = 5
}`, guildID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_audit_log.test", "guild_id", guildID),
				resource.TestCheckResourceAttrSet("data.discord_audit_log.test", "entries.#"),
			),
		}},
	})
}
