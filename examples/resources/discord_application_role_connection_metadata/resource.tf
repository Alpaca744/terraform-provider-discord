resource "discord_application_role_connection_metadata" "this" {
  application_id = "123456789012345678"

  records = [
    {
      type        = 2 # INTEGER_GREATER_THAN_OR_EQUAL
      key         = "level"
      name        = "Level"
      description = "Minimum account level"
    },
  ]
}
