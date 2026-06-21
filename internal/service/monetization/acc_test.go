package monetization_test

import (
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSKUDataSource verifies the discord_sku data source reads application SKUs.
// This test requires a monetization-enabled application; it passes even with an
// empty SKU list as long as the attribute is set.
func TestAccSKUDataSource(t *testing.T) {
	appID := acctest.TestApplicationID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccSKUConfig(appID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.discord_sku.test", "skus.#"),
			),
		}},
	})
}

func testAccSKUConfig(appID string) string {
	return `data "discord_sku" "test" { application_id = "` + appID + `" }`
}

// TestAccEntitlementDataSource verifies the discord_entitlement data source reads
// application entitlements. An empty list is valid and still passes.
func TestAccEntitlementDataSource(t *testing.T) {
	appID := acctest.TestApplicationID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccEntitlementConfig(appID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.discord_entitlement.test", "entitlements.#"),
			),
		}},
	})
}

func testAccEntitlementConfig(appID string) string {
	return `data "discord_entitlement" "test" { application_id = "` + appID + `" }`
}
