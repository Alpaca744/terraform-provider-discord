package voice_test

import (
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVoiceRegionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "discord_voice_regions" "test" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.discord_voice_regions.test", "regions.#"),
			),
		}},
	})
}
