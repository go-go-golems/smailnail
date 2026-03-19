output "alice_id" {
  value = try(keycloak_user.alice[0].id, null)
}
