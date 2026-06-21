package application_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCurrentApplicationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "discord_current_application" "test" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.discord_current_application.test", "id"),
				resource.TestCheckResourceAttrSet("data.discord_current_application.test", "name"),
			),
		}},
	})
}

func TestAccApplicationSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationSettingsConfig("A test bot"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("discord_application_settings.test", "id"),
					resource.TestCheckResourceAttr("discord_application_settings.test", "description", "A test bot"),
				),
			},
			{
				Config: testAccApplicationSettingsConfig("Updated description"),
				Check:  resource.TestCheckResourceAttr("discord_application_settings.test", "description", "Updated description"),
			},
		},
	})
}

func testAccApplicationSettingsConfig(description string) string {
	return fmt.Sprintf(`
resource "discord_application_settings" "test" {
  description = %q
}`, description)
}

func TestAccRoleConnectionMetadata(t *testing.T) {
	appID := acctest.TestApplicationID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConnectionConfig(appID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_application_role_connection_metadata.test", "application_id", appID),
					resource.TestCheckResourceAttr("discord_application_role_connection_metadata.test", "records.0.key", "level"),
				),
			},
			{
				ResourceName:                      "discord_application_role_connection_metadata.test",
				ImportState:                       true,
				ImportStateVerify:                 true,
				ImportStateVerifyIdentifierAttribute: "application_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["discord_application_role_connection_metadata.test"]
					return rs.Primary.Attributes["application_id"], nil
				},
			},
		},
	})
}

func testAccRoleConnectionConfig(appID string) string {
	return fmt.Sprintf(`
resource "discord_application_role_connection_metadata" "test" {
  application_id = %q
  records = [
    {
      key         = "level"
      name        = "Level"
      description = "Player level"
      type        = 2
    }
  ]
}`, appID)
}
