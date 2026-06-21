data "discord_voice_regions" "all" {}

output "optimal_regions" {
  value = [for r in data.discord_voice_regions.all.regions : r.id if r.optimal]
}
