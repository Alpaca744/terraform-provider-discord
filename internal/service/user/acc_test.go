package user_test

import (
	"fmt"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCurrentUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "discord_current_user" "test" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.discord_current_user.test", "id"),
				resource.TestCheckResourceAttrSet("data.discord_current_user.test", "username"),
				resource.TestCheckResourceAttr("data.discord_current_user.test", "bot", "true"),
			),
		}},
	})
}

func TestAccUserDataSource(t *testing.T) {
	appID := acctest.TestApplicationID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "discord_user" "test" { id = %q }`, appID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.discord_user.test", "id", appID),
				resource.TestCheckResourceAttrSet("data.discord_user.test", "username"),
			),
		}},
	})
}
