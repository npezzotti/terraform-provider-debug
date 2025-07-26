data "debug_dns_lookup" "example" {
  hostname = "does-not-exist.com"
}

output "ips" {
  value = data.debug_dns_lookup.example.result
}
